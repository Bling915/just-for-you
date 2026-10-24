package mysql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	application "github.com/nanfang/just-for-you/server/internal/moment/application"
	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
	"strconv"
	"strings"
	"time"
)

type MySQLMomentStore struct{ db *sql.DB }

const momentSelectColumns = "id,author_user_id,couple_id,space_type,visibility,kind,source_moment_id,daily_question_id,body,status_code,place_name,latitude,longitude,occurred_at,created_at,updated_at,deleted_at"

func storedKind(value domain.Moment) string {
	if value.DailyQuestionID() != nil {
		return "daily_question"
	}
	return "standard"
}

func (r MySQLMomentStore) Calendar(ctx context.Context, coupleID int64, from, to time.Time) ([]application.CalendarDay, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT DATE_FORMAT(CONVERT_TZ(occurred_at,'+00:00','+08:00'),'%Y-%m-%d'),COUNT(*) FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL GROUP BY 1 ORDER BY 1", coupleID, from.UTC(), to.UTC())
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []application.CalendarDay{}
	for rows.Next() {
		var v application.CalendarDay
		if e = rows.Scan(&v.Date, &v.Count); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r MySQLMomentStore) Map(ctx context.Context, coupleID int64, from, to time.Time, limit int) ([]application.MapPoint, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,latitude,longitude,occurred_at,place_name FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND latitude IS NOT NULL AND longitude IS NOT NULL AND deleted_at IS NULL ORDER BY occurred_at DESC LIMIT ?", coupleID, from.UTC(), to.UTC(), limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []application.MapPoint{}
	for rows.Next() {
		var v application.MapPoint
		if e = rows.Scan(&v.MomentID, &v.Latitude, &v.Longitude, &v.OccurredAt, &v.PlaceName); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r MySQLMomentStore) MonthlyReport(ctx context.Context, coupleID int64, from, to time.Time) (application.MonthlyReport, error) {
	var v application.MonthlyReport
	e := r.db.QueryRowContext(ctx, "SELECT COUNT(*),(SELECT COUNT(*) FROM moment_photos p JOIN moments m2 ON m2.id=p.moment_id WHERE m2.couple_id=? AND m2.space_type='shared' AND m2.occurred_at>=? AND m2.occurred_at<? AND m2.deleted_at IS NULL) FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL", coupleID, from.UTC(), to.UTC(), coupleID, from.UTC(), to.UTC()).Scan(&v.MomentCount, &v.PhotoCount)
	if e != nil {
		return v, e
	}
	_ = r.db.QueryRowContext(ctx, "SELECT place_name FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL AND place_name<>'' GROUP BY place_name ORDER BY COUNT(*) DESC,place_name LIMIT 1", coupleID, from.UTC(), to.UTC()).Scan(&v.TopPlace)
	_ = r.db.QueryRowContext(ctx, "SELECT status_code FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL AND status_code<>'' GROUP BY status_code ORDER BY COUNT(*) DESC,status_code LIMIT 1", coupleID, from.UTC(), to.UTC()).Scan(&v.CommonStatus)
	return v, nil
}

func NewMySQLMomentStore(db *sql.DB) MySQLMomentStore { return MySQLMomentStore{db: db} }
func (r MySQLMomentStore) CreateGrant(ctx context.Context, userID int64, input application.UploadInput, now time.Time) (application.UploadGrant, error) {
	extension := map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}[input.ContentType]
	if extension == "" || input.Size < 1 || input.Size > 10<<20 {
		return application.UploadGrant{}, domain.ErrInvalidUpload
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return application.UploadGrant{}, err
	}
	key := fmt.Sprintf("users/%d/%04d/%02d/%x.%s", userID, now.Year(), now.Month(), random, extension)
	grant := application.UploadGrant{ObjectKey: key, OwnerID: userID, ContentType: input.ContentType, ExpiresAt: now.Add(10 * time.Minute)}
	_, err := r.db.ExecContext(ctx, "INSERT INTO upload_grants(user_id,object_key,content_type,expires_at) VALUES(?,?,?,?)", userID, key, input.ContentType, grant.ExpiresAt)
	return grant, err
}
func (r MySQLMomentStore) ValidateGrants(ctx context.Context, userID int64, keys []string, now time.Time) ([]application.UploadGrant, error) {
	grants := make([]application.UploadGrant, len(keys))
	for index, key := range keys {
		var grant application.UploadGrant
		var consumed sql.NullTime
		err := r.db.QueryRowContext(ctx, "SELECT user_id,content_type,expires_at,consumed_at FROM upload_grants WHERE object_key=?", key).Scan(&grant.OwnerID, &grant.ContentType, &grant.ExpiresAt, &consumed)
		if err != nil || grant.OwnerID != userID || consumed.Valid || !grant.ExpiresAt.After(now) {
			return nil, domain.ErrInvalidUpload
		}
		grant.ObjectKey = key
		grants[index] = grant
	}
	return grants, nil
}
func (r MySQLMomentStore) SaveWithGrants(ctx context.Context, moment domain.Moment, grants []application.UploadGrant) (domain.Moment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Moment{}, err
	}
	defer tx.Rollback()
	for _, grant := range grants {
		var owner int64
		var expires time.Time
		var consumed sql.NullTime
		if err = tx.QueryRowContext(ctx, "SELECT user_id,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", grant.ObjectKey).Scan(&owner, &expires, &consumed); err != nil || owner != moment.AuthorID() || consumed.Valid || !expires.After(moment.CreatedAt()) {
			return domain.Moment{}, domain.ErrInvalidUpload
		}
	}
	if moment.CoupleID() > 0 {
		var relationshipID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM couples WHERE id=? AND status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", moment.CoupleID(), moment.AuthorID(), moment.AuthorID()).Scan(&relationshipID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Moment{}, coupledomain.ErrNoActiveCouple
		}
		if err != nil {
			return domain.Moment{}, err
		}
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO moments(couple_id,author_user_id,space_type,visibility,kind,source_moment_id,daily_question_id,body,status_code,place_name,latitude,longitude,occurred_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", nullablePositiveID(moment.CoupleID()), moment.AuthorID(), string(moment.SpaceType()), string(moment.Visibility()), storedKind(moment), moment.SourceMomentID(), moment.DailyQuestionID(), moment.Text(), string(moment.Status()), moment.Place(), moment.Latitude(), moment.Longitude(), moment.OccurredAt(), moment.CreatedAt(), moment.UpdatedAt())
	if err != nil {
		return domain.Moment{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Moment{}, err
	}
	for _, photo := range moment.Photos() {
		if _, err = tx.ExecContext(ctx, "INSERT INTO moment_photos(moment_id,object_key,sort_order) VALUES(?,?,?)", id, photo.ObjectKey(), photo.Order()); err != nil {
			return domain.Moment{}, err
		}
	}
	for _, grant := range grants {
		if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", moment.CreatedAt(), grant.ObjectKey); err != nil {
			return domain.Moment{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Moment{}, err
	}
	return domain.RestoreMomentDetails(domain.RestoreInput{ID: id, AuthorID: moment.AuthorID(), CoupleID: moment.CoupleID(), Text: moment.Text(), Status: moment.Status(), Place: moment.Place(), Latitude: moment.Latitude(), Longitude: moment.Longitude(), Photos: moment.Photos(), OccurredAt: moment.OccurredAt(), CreatedAt: moment.CreatedAt(), UpdatedAt: moment.UpdatedAt(), SpaceType: moment.SpaceType(), Visibility: moment.Visibility(), SourceMomentID: moment.SourceMomentID(), DailyQuestionID: moment.DailyQuestionID()}), nil
}

func nullablePositiveID(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}
func (r MySQLMomentStore) UpdateWithGrants(ctx context.Context, moment domain.Moment, version time.Time, grants []application.UploadGrant) (domain.Moment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Moment{}, err
	}
	defer tx.Rollback()
	for _, grant := range grants {
		var owner int64
		var expires time.Time
		var consumed sql.NullTime
		if err = tx.QueryRowContext(ctx, "SELECT user_id,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", grant.ObjectKey).Scan(&owner, &expires, &consumed); err != nil || owner != moment.AuthorID() || consumed.Valid || !expires.After(moment.UpdatedAt()) {
			return domain.Moment{}, domain.ErrInvalidUpload
		}
	}
	if moment.CoupleID() > 0 {
		var relationshipID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM couples WHERE id=? AND status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", moment.CoupleID(), moment.AuthorID(), moment.AuthorID()).Scan(&relationshipID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Moment{}, coupledomain.ErrNoActiveCouple
		}
		if err != nil {
			return domain.Moment{}, err
		}
	}
	result, err := tx.ExecContext(ctx, "UPDATE moments SET couple_id=?,visibility=?,body=?,status_code=?,place_name=?,latitude=?,longitude=?,occurred_at=?,updated_at=? WHERE id=? AND author_user_id=? AND space_type=? AND updated_at=? AND deleted_at IS NULL", nullablePositiveID(moment.CoupleID()), string(moment.Visibility()), moment.Text(), string(moment.Status()), moment.Place(), moment.Latitude(), moment.Longitude(), moment.OccurredAt(), moment.UpdatedAt(), moment.ID(), moment.AuthorID(), string(moment.SpaceType()), version)
	if err != nil {
		return domain.Moment{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return domain.Moment{}, err
	}
	if count != 1 {
		return domain.Moment{}, domain.ErrMomentConflict
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM moment_photos WHERE moment_id=?", moment.ID()); err != nil {
		return domain.Moment{}, err
	}
	for _, photo := range moment.Photos() {
		if _, err = tx.ExecContext(ctx, "INSERT INTO moment_photos(moment_id,object_key,sort_order) VALUES(?,?,?)", moment.ID(), photo.ObjectKey(), photo.Order()); err != nil {
			return domain.Moment{}, err
		}
	}
	for _, grant := range grants {
		if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", moment.UpdatedAt(), grant.ObjectKey); err != nil {
			return domain.Moment{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Moment{}, err
	}
	return moment, nil
}
func (r MySQLMomentStore) List(ctx context.Context, filter application.TimelineFilter, cursor string, limit int) (application.TimelinePage, error) {
	query := "SELECT " + momentSelectColumns + " FROM moments WHERE "
	args := []any{}
	switch {
	case filter.SpaceType == domain.SpaceShared && filter.CoupleID > 0 && filter.OwnerID == 0:
		query += "couple_id=? AND space_type='shared' AND deleted_at IS NULL"
		args = append(args, filter.CoupleID)
	case filter.SpaceType == domain.SpacePersonal && filter.OwnerID == filter.ViewerID && filter.OwnerID > 0:
		query += "author_user_id=? AND space_type='personal' AND deleted_at IS NULL"
		args = append(args, filter.OwnerID)
	case filter.SpaceType == domain.SpacePersonal && filter.CoupleID > 0 && filter.OwnerID > 0 && filter.OwnerID != filter.ViewerID:
		query += "couple_id=? AND author_user_id=? AND space_type='personal' AND visibility='partner' AND deleted_at IS NULL"
		args = append(args, filter.CoupleID, filter.OwnerID)
	default:
		return application.TimelinePage{}, application.ErrInvalidSpaceQuery
	}
	if cursor != "" {
		at, id, err := decodeCursor(cursor)
		if err != nil {
			return application.TimelinePage{}, err
		}
		query += " AND (occurred_at < ? OR (occurred_at = ? AND id < ?))"
		args = append(args, at, at, id)
	}
	query += " ORDER BY occurred_at DESC,id DESC LIMIT ?"
	args = append(args, limit+1)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return application.TimelinePage{}, err
	}
	defer rows.Close()
	moments := []domain.Moment{}
	for rows.Next() {
		moment, err := r.scanMoment(ctx, rows)
		if err != nil {
			return application.TimelinePage{}, err
		}
		moments = append(moments, moment)
	}
	page := application.TimelinePage{Moments: moments}
	if len(moments) > limit {
		last := moments[limit-1]
		page.NextCursor = encodeCursor(last.OccurredAt(), last.ID())
		page.Moments = moments[:limit]
	}
	return page, rows.Err()
}

type scanner interface{ Scan(...any) error }

func (r MySQLMomentStore) scanMoment(ctx context.Context, row scanner) (domain.Moment, error) {
	var id, author int64
	var relationship, sourceMoment, dailyQuestion sql.NullInt64
	var text, status, place, spaceType, visibility, kind string
	var latitude, longitude sql.NullFloat64
	var occurred, created, updated time.Time
	var deleted sql.NullTime
	if err := row.Scan(&id, &author, &relationship, &spaceType, &visibility, &kind, &sourceMoment, &dailyQuestion, &text, &status, &place, &latitude, &longitude, &occurred, &created, &updated, &deleted); err != nil {
		return domain.Moment{}, err
	}
	photoRows, err := r.db.QueryContext(ctx, "SELECT object_key,sort_order FROM moment_photos WHERE moment_id=? ORDER BY sort_order", id)
	if err != nil {
		return domain.Moment{}, err
	}
	defer photoRows.Close()
	photos := []domain.Photo{}
	for photoRows.Next() {
		var key string
		var order int
		if err = photoRows.Scan(&key, &order); err != nil {
			return domain.Moment{}, err
		}
		photos = append(photos, domain.NewPhoto(key, order))
	}
	var deletedAt *time.Time
	if deleted.Valid {
		deletedAt = &deleted.Time
	}
	var latitudeValue, longitudeValue *float64
	if latitude.Valid {
		latitudeValue = &latitude.Float64
	}
	if longitude.Valid {
		longitudeValue = &longitude.Float64
	}
	var sourceMomentID, dailyQuestionID *int64
	if sourceMoment.Valid {
		sourceMomentID = &sourceMoment.Int64
	}
	if dailyQuestion.Valid {
		dailyQuestionID = &dailyQuestion.Int64
	}
	return domain.RestoreMomentDetails(domain.RestoreInput{ID: id, AuthorID: author, CoupleID: relationship.Int64, Text: text, Status: domain.Status(status), Place: place, Latitude: latitudeValue, Longitude: longitudeValue, Photos: photos, OccurredAt: occurred, CreatedAt: created, UpdatedAt: updated, DeletedAt: deletedAt, SpaceType: domain.SpaceType(spaceType), Visibility: domain.Visibility(visibility), SourceMomentID: sourceMomentID, DailyQuestionID: dailyQuestionID}), nil
}
func (r MySQLMomentStore) Recent(ctx context.Context, filter application.TimelineFilter, limit int) ([]domain.Moment, error) {
	page, err := r.List(ctx, filter, "", limit)
	return page.Moments, err
}
func (r MySQLMomentStore) LatestStatusByAuthor(ctx context.Context, relationshipID, authorID int64) (*domain.Moment, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+momentSelectColumns+" FROM moments WHERE couple_id=? AND space_type='shared' AND author_user_id=? AND status_code<>'' AND deleted_at IS NULL ORDER BY occurred_at DESC,id DESC LIMIT 1", relationshipID, authorID)
	moment, err := r.scanMoment(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &moment, err
}
func (r MySQLMomentStore) Find(ctx context.Context, lookup application.MomentLookup) (domain.Moment, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+momentSelectColumns+" FROM moments WHERE id=? AND deleted_at IS NULL AND ((author_user_id=? AND space_type='personal') OR (couple_id=? AND space_type='shared') OR (couple_id=? AND space_type='personal' AND visibility='partner' AND author_user_id<>?))", lookup.MomentID, lookup.ViewerID, lookup.CoupleID, lookup.CoupleID, lookup.ViewerID)
	moment, err := r.scanMoment(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrMomentNotFound
	}
	return moment, err
}
func (r MySQLMomentStore) Delete(ctx context.Context, momentID int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE moments SET deleted_at=CURRENT_TIMESTAMP(6) WHERE id=? AND deleted_at IS NULL", momentID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return domain.ErrMomentNotFound
	}
	return nil
}

func (r MySQLMomentStore) ShareToCouple(ctx context.Context, sourceID, authorID, relationshipID int64, message string, now time.Time) (domain.Moment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Moment{}, err
	}
	defer tx.Rollback()
	var sourceAuthor int64
	var spaceType, kind, body, status, place string
	var latitude, longitude sql.NullFloat64
	var occurredAt time.Time
	err = tx.QueryRowContext(ctx, "SELECT author_user_id,space_type,kind,body,status_code,place_name,latitude,longitude,occurred_at FROM moments WHERE id=? AND author_user_id=? AND deleted_at IS NULL FOR UPDATE", sourceID, authorID).Scan(&sourceAuthor, &spaceType, &kind, &body, &status, &place, &latitude, &longitude, &occurredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Moment{}, domain.ErrMomentNotFound
	}
	if err != nil {
		return domain.Moment{}, err
	}
	if spaceType != string(domain.SpacePersonal) || kind != "standard" {
		return domain.Moment{}, application.ErrCannotShareMoment
	}
	photoRows, err := tx.QueryContext(ctx, "SELECT object_key,sort_order FROM moment_photos WHERE moment_id=? ORDER BY sort_order", sourceID)
	if err != nil {
		return domain.Moment{}, err
	}
	photos := []domain.Photo{}
	for photoRows.Next() {
		var key string
		var order int
		if err = photoRows.Scan(&key, &order); err != nil {
			photoRows.Close()
			return domain.Moment{}, err
		}
		photos = append(photos, domain.NewPhoto(key, order))
	}
	if err = photoRows.Close(); err != nil {
		return domain.Moment{}, err
	}
	var activeID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM couples WHERE id=? AND status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", relationshipID, authorID, authorID).Scan(&activeID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Moment{}, coupledomain.ErrNoActiveCouple
	}
	if err != nil {
		return domain.Moment{}, err
	}
	text := strings.TrimSpace(message)
	if text == "" {
		text = body
	}
	latitudeValue, longitudeValue := nullFloatPointer(latitude), nullFloatPointer(longitude)
	copyMoment, err := domain.NewMomentDetails(domain.CreateInput{AuthorID: authorID, CoupleID: relationshipID, Text: text, Status: domain.Status(status), Place: place, Latitude: latitudeValue, Longitude: longitudeValue, PhotoKeys: photoObjectKeys(photos), OccurredAt: occurredAt, CreatedAt: now, SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, SourceMomentID: &sourceID})
	if err != nil {
		return domain.Moment{}, err
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO moments(couple_id,author_user_id,space_type,visibility,kind,source_moment_id,daily_question_id,body,status_code,place_name,latitude,longitude,occurred_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", relationshipID, authorID, "shared", "couple", "standard", sourceID, nil, copyMoment.Text(), string(copyMoment.Status()), copyMoment.Place(), copyMoment.Latitude(), copyMoment.Longitude(), copyMoment.OccurredAt(), now, now)
	if err != nil {
		return domain.Moment{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Moment{}, err
	}
	if len(photos) > 0 {
		if _, err = tx.ExecContext(ctx, "INSERT INTO moment_photos(moment_id,object_key,sort_order) SELECT ?,object_key,sort_order FROM moment_photos WHERE moment_id=?", id, sourceID); err != nil {
			return domain.Moment{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Moment{}, err
	}
	return domain.RestoreMomentDetails(domain.RestoreInput{ID: id, AuthorID: authorID, CoupleID: relationshipID, Text: copyMoment.Text(), Status: copyMoment.Status(), Place: copyMoment.Place(), Latitude: copyMoment.Latitude(), Longitude: copyMoment.Longitude(), Photos: photos, OccurredAt: occurredAt, CreatedAt: now, UpdatedAt: now, SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, SourceMomentID: &sourceID}), nil
}

func (r MySQLMomentStore) UpsertReaction(ctx context.Context, reaction domain.Reaction, relationshipID int64) (domain.Reaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Reaction{}, err
	}
	defer tx.Rollback()
	var authorID, coupleID int64
	var spaceType, visibility string
	err = tx.QueryRowContext(ctx, "SELECT author_user_id,space_type,visibility,couple_id FROM moments WHERE id=? AND couple_id=? AND deleted_at IS NULL FOR UPDATE", reaction.MomentID(), relationshipID).Scan(&authorID, &spaceType, &visibility, &coupleID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Reaction{}, domain.ErrMomentNotFound
	}
	if err != nil {
		return domain.Reaction{}, err
	}
	readableByPartner := (spaceType == string(domain.SpaceShared) && visibility == string(domain.VisibilityCouple)) ||
		(spaceType == string(domain.SpacePersonal) && visibility == string(domain.VisibilityPartner))
	if authorID == reaction.UserID() || coupleID != relationshipID || !readableByPartner {
		return domain.Reaction{}, application.ErrMomentNotVisible
	}
	var activeID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM couples WHERE id=? AND status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", relationshipID, reaction.UserID(), reaction.UserID()).Scan(&activeID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Reaction{}, coupledomain.ErrNoActiveCouple
	}
	if err != nil {
		return domain.Reaction{}, err
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO moment_reactions(moment_id,user_id,reaction_type,text_content,created_at,updated_at) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE text_content=VALUES(text_content),updated_at=VALUES(updated_at),id=LAST_INSERT_ID(id)", reaction.MomentID(), reaction.UserID(), string(reaction.Type()), reaction.Text(), reaction.CreatedAt(), reaction.UpdatedAt())
	if err != nil {
		return domain.Reaction{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Reaction{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Reaction{}, err
	}
	return domain.RestoreReaction(id, reaction.MomentID(), reaction.UserID(), reaction.Type(), reaction.Text(), reaction.CreatedAt(), reaction.UpdatedAt()), nil
}

func (r MySQLMomentStore) DeleteReaction(ctx context.Context, reactionID, userID int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM moment_reactions WHERE id=? AND user_id=?", reactionID, userID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrReactionNotFound
	}
	return nil
}

func (r MySQLMomentStore) ListReactions(ctx context.Context, momentIDs []int64) (map[int64][]domain.Reaction, error) {
	result := map[int64][]domain.Reaction{}
	if len(momentIDs) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(momentIDs)), ",")
	args := make([]any, len(momentIDs))
	for index, id := range momentIDs {
		args[index] = id
	}
	rows, err := r.db.QueryContext(ctx, "SELECT id,moment_id,user_id,reaction_type,text_content,created_at,updated_at FROM moment_reactions WHERE moment_id IN ("+placeholders+") ORDER BY created_at,id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, momentID, userID int64
		var reactionType, text string
		var createdAt, updatedAt time.Time
		if err = rows.Scan(&id, &momentID, &userID, &reactionType, &text, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		result[momentID] = append(result[momentID], domain.RestoreReaction(id, momentID, userID, domain.ReactionType(reactionType), text, createdAt, updatedAt))
	}
	return result, rows.Err()
}

func (r MySQLMomentStore) SavePerspectiveWithGrants(ctx context.Context, value domain.Perspective, grants []application.UploadGrant, relationshipID int64) (domain.Perspective, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Perspective{}, err
	}
	defer tx.Rollback()
	var sourceAuthor, activeID int64
	var space string
	err = tx.QueryRowContext(ctx, "SELECT author_user_id,space_type FROM moments WHERE id=? AND couple_id=? AND deleted_at IS NULL FOR UPDATE", value.MomentID(), relationshipID).Scan(&sourceAuthor, &space)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Perspective{}, domain.ErrMomentNotFound
	}
	if err != nil {
		return domain.Perspective{}, err
	}
	if space != "shared" {
		return domain.Perspective{}, application.ErrPerspectiveOnlyForShared
	}
	if sourceAuthor == value.AuthorID() {
		return domain.Perspective{}, application.ErrPerspectiveRequiresPartner
	}
	err = tx.QueryRowContext(ctx, "SELECT id FROM couples WHERE id=? AND status='active' AND (user_a_id=? OR user_b_id=?) FOR UPDATE", relationshipID, value.AuthorID(), value.AuthorID()).Scan(&activeID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Perspective{}, coupledomain.ErrNoActiveCouple
	}
	if err != nil {
		return domain.Perspective{}, err
	}
	for _, grant := range grants {
		var owner int64
		var expires time.Time
		var consumed sql.NullTime
		if err = tx.QueryRowContext(ctx, "SELECT user_id,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", grant.ObjectKey).Scan(&owner, &expires, &consumed); err != nil || owner != value.AuthorID() || consumed.Valid || !expires.After(value.UpdatedAt()) {
			return domain.Perspective{}, domain.ErrInvalidUpload
		}
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO moment_perspectives(moment_id,author_user_id,body,created_at,updated_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE body=VALUES(body),updated_at=VALUES(updated_at),id=LAST_INSERT_ID(id)", value.MomentID(), value.AuthorID(), value.Text(), value.CreatedAt(), value.UpdatedAt())
	if err != nil {
		return domain.Perspective{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Perspective{}, err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM moment_perspective_photos WHERE perspective_id=?", id); err != nil {
		return domain.Perspective{}, err
	}
	for _, photo := range value.Photos() {
		if _, err = tx.ExecContext(ctx, "INSERT INTO moment_perspective_photos(perspective_id,object_key,sort_order) VALUES(?,?,?)", id, photo.ObjectKey(), photo.Order()); err != nil {
			return domain.Perspective{}, err
		}
	}
	for _, grant := range grants {
		if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", value.UpdatedAt(), grant.ObjectKey); err != nil {
			return domain.Perspective{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Perspective{}, err
	}
	return domain.RestorePerspective(id, value.MomentID(), value.AuthorID(), value.Text(), value.Photos(), value.CreatedAt(), value.UpdatedAt()), nil
}

func (r MySQLMomentStore) DeletePerspective(ctx context.Context, momentID, userID int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM moment_perspectives WHERE moment_id=? AND author_user_id=?", momentID, userID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrNotPerspectiveAuthor
	}
	return nil
}

func (r MySQLMomentStore) ListPerspectives(ctx context.Context, momentIDs []int64) (map[int64]domain.Perspective, error) {
	result := map[int64]domain.Perspective{}
	if len(momentIDs) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(momentIDs)), ",")
	args := make([]any, len(momentIDs))
	for i, id := range momentIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, "SELECT id,moment_id,author_user_id,body,created_at,updated_at FROM moment_perspectives WHERE moment_id IN ("+placeholders+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, momentID, authorID int64
		var body string
		var created, updated time.Time
		if err = rows.Scan(&id, &momentID, &authorID, &body, &created, &updated); err != nil {
			return nil, err
		}
		photoRows, queryErr := r.db.QueryContext(ctx, "SELECT object_key,sort_order FROM moment_perspective_photos WHERE perspective_id=? ORDER BY sort_order", id)
		if queryErr != nil {
			return nil, queryErr
		}
		photos := []domain.Photo{}
		for photoRows.Next() {
			var key string
			var order int
			if queryErr = photoRows.Scan(&key, &order); queryErr != nil {
				photoRows.Close()
				return nil, queryErr
			}
			photos = append(photos, domain.NewPhoto(key, order))
		}
		if queryErr = photoRows.Close(); queryErr != nil {
			return nil, queryErr
		}
		result[momentID] = domain.RestorePerspective(id, momentID, authorID, body, photos, created, updated)
	}
	return result, rows.Err()
}

func (r MySQLMomentStore) ResurfaceCandidates(ctx context.Context, relationshipID int64, now time.Time, limit int) ([]domain.Moment, error) {
	local := now.In(time.FixedZone("CST", 8*3600))
	today := local.Format("01-02")
	monthAgo := local.AddDate(0, -1, 0).Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx, "SELECT "+momentSelectColumns+" FROM moments WHERE couple_id=? AND space_type='shared' AND visibility='couple' AND deleted_at IS NULL AND occurred_at<? ORDER BY (DATE_FORMAT(CONVERT_TZ(occurred_at,'+00:00','+08:00'),'%m-%d')=?) DESC,(DATE(CONVERT_TZ(occurred_at,'+00:00','+08:00'))=?) DESC,occurred_at ASC,id ASC LIMIT ?", relationshipID, now.AddDate(0, 0, -7), today, monthAgo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []domain.Moment{}
	for rows.Next() {
		value, e := r.scanMoment(ctx, rows)
		if e != nil {
			return nil, e
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func nullFloatPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func photoObjectKeys(photos []domain.Photo) []string {
	keys := make([]string, len(photos))
	for index, photo := range photos {
		keys[index] = photo.ObjectKey()
	}
	return keys
}
func encodeCursor(at time.Time, id int64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(at.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(id, 10)))
}
func decodeCursor(raw string) (time.Time, int64, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return time.Time{}, 0, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("invalid cursor")
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, err
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	return at, id, err
}
