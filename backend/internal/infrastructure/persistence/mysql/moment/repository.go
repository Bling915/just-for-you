package infrastructure

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"strconv"
	"strings"
	"time"
)

type MySQLRepository struct{ db *sql.DB }

const momentSelectColumns = "id,author_user_id,couple_id,space_type,visibility,kind,source_moment_id,daily_question_id,body,status_code,place_name,latitude,longitude,occurred_at,created_at,updated_at,deleted_at"

func (r MySQLRepository) Calendar(ctx context.Context, coupleID int64, from, to time.Time) ([]application.CalendarDay, error) {
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
func (r MySQLRepository) Map(ctx context.Context, coupleID int64, from, to time.Time, limit int) ([]application.MapPoint, error) {
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
func (r MySQLRepository) MonthlyReport(ctx context.Context, coupleID int64, from, to time.Time) (application.MonthlyReport, error) {
	var v application.MonthlyReport
	e := r.db.QueryRowContext(ctx, "SELECT COUNT(*),(SELECT COUNT(*) FROM moment_photos p JOIN moments m2 ON m2.id=p.moment_id WHERE m2.couple_id=? AND m2.space_type='shared' AND m2.occurred_at>=? AND m2.occurred_at<? AND m2.deleted_at IS NULL) FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL", coupleID, from.UTC(), to.UTC(), coupleID, from.UTC(), to.UTC()).Scan(&v.MomentCount, &v.PhotoCount)
	if e != nil {
		return v, e
	}
	_ = r.db.QueryRowContext(ctx, "SELECT place_name FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL AND place_name<>'' GROUP BY place_name ORDER BY COUNT(*) DESC,place_name LIMIT 1", coupleID, from.UTC(), to.UTC()).Scan(&v.TopPlace)
	_ = r.db.QueryRowContext(ctx, "SELECT status_code FROM moments WHERE couple_id=? AND space_type='shared' AND occurred_at>=? AND occurred_at<? AND deleted_at IS NULL AND status_code<>'' GROUP BY status_code ORDER BY COUNT(*) DESC,status_code LIMIT 1", coupleID, from.UTC(), to.UTC()).Scan(&v.CommonStatus)
	return v, nil
}

func NewMySQLRepository(db *sql.DB) MySQLRepository { return MySQLRepository{db: db} }
func (r MySQLRepository) CreateGrant(ctx context.Context, userID int64, input application.UploadInput, now time.Time) (application.UploadGrant, error) {
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
func (r MySQLRepository) ValidateGrants(ctx context.Context, userID int64, keys []string, now time.Time) ([]application.UploadGrant, error) {
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
func (r MySQLRepository) SaveWithGrants(ctx context.Context, moment domain.Moment, grants []application.UploadGrant) (domain.Moment, error) {
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
	result, err := tx.ExecContext(ctx, "INSERT INTO moments(couple_id,author_user_id,space_type,visibility,kind,source_moment_id,daily_question_id,body,status_code,place_name,latitude,longitude,occurred_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", nullablePositiveID(moment.RelationshipID()), moment.AuthorID(), string(moment.SpaceType()), string(moment.Visibility()), string(moment.Kind()), moment.SourceMomentID(), moment.DailyQuestionID(), moment.Text(), string(moment.Status()), moment.Place(), moment.Latitude(), moment.Longitude(), moment.OccurredAt(), moment.CreatedAt(), moment.UpdatedAt())
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
	return domain.RestoreMomentDetails(domain.RestoreInput{ID: id, AuthorID: moment.AuthorID(), RelationshipID: moment.RelationshipID(), Text: moment.Text(), Status: moment.Status(), Place: moment.Place(), Latitude: moment.Latitude(), Longitude: moment.Longitude(), Photos: moment.Photos(), OccurredAt: moment.OccurredAt(), CreatedAt: moment.CreatedAt(), UpdatedAt: moment.UpdatedAt(), SpaceType: moment.SpaceType(), Visibility: moment.Visibility(), Kind: moment.Kind(), SourceMomentID: moment.SourceMomentID(), DailyQuestionID: moment.DailyQuestionID()}), nil
}

func nullablePositiveID(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}
func (r MySQLRepository) UpdateWithGrants(ctx context.Context, moment domain.Moment, version time.Time, grants []application.UploadGrant) (domain.Moment, error) {
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
	result, err := tx.ExecContext(ctx, "UPDATE moments SET body=?,status_code=?,place_name=?,latitude=?,longitude=?,occurred_at=?,updated_at=? WHERE id=? AND couple_id=? AND updated_at=? AND deleted_at IS NULL", moment.Text(), string(moment.Status()), moment.Place(), moment.Latitude(), moment.Longitude(), moment.OccurredAt(), moment.UpdatedAt(), moment.ID(), moment.RelationshipID(), version)
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
func (r MySQLRepository) List(ctx context.Context, filter application.TimelineFilter, cursor string, limit int) (application.TimelinePage, error) {
	query := "SELECT " + momentSelectColumns + " FROM moments WHERE "
	args := []any{}
	switch {
	case filter.SpaceType == domain.SpaceShared && filter.RelationshipID > 0 && filter.OwnerID == 0:
		query += "couple_id=? AND space_type='shared' AND deleted_at IS NULL"
		args = append(args, filter.RelationshipID)
	case filter.SpaceType == domain.SpacePersonal && filter.OwnerID == filter.ViewerID && filter.OwnerID > 0:
		query += "author_user_id=? AND space_type='personal' AND deleted_at IS NULL"
		args = append(args, filter.OwnerID)
	case filter.SpaceType == domain.SpacePersonal && filter.RelationshipID > 0 && filter.OwnerID > 0 && filter.OwnerID != filter.ViewerID:
		query += "couple_id=? AND author_user_id=? AND space_type='personal' AND visibility='partner' AND deleted_at IS NULL"
		args = append(args, filter.RelationshipID, filter.OwnerID)
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

func (r MySQLRepository) scanMoment(ctx context.Context, row scanner) (domain.Moment, error) {
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
	return domain.RestoreMomentDetails(domain.RestoreInput{ID: id, AuthorID: author, RelationshipID: relationship.Int64, Text: text, Status: domain.Status(status), Place: place, Latitude: latitudeValue, Longitude: longitudeValue, Photos: photos, OccurredAt: occurred, CreatedAt: created, UpdatedAt: updated, DeletedAt: deletedAt, SpaceType: domain.SpaceType(spaceType), Visibility: domain.Visibility(visibility), Kind: domain.Kind(kind), SourceMomentID: sourceMomentID, DailyQuestionID: dailyQuestionID}), nil
}
func (r MySQLRepository) Recent(ctx context.Context, filter application.TimelineFilter, limit int) ([]domain.Moment, error) {
	page, err := r.List(ctx, filter, "", limit)
	return page.Moments, err
}
func (r MySQLRepository) LatestStatusByAuthor(ctx context.Context, relationshipID, authorID int64) (*domain.Moment, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+momentSelectColumns+" FROM moments WHERE couple_id=? AND space_type='shared' AND author_user_id=? AND status_code<>'' AND deleted_at IS NULL ORDER BY occurred_at DESC,id DESC LIMIT 1", relationshipID, authorID)
	moment, err := r.scanMoment(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &moment, err
}
func (r MySQLRepository) Find(ctx context.Context, lookup application.MomentLookup) (domain.Moment, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+momentSelectColumns+" FROM moments WHERE id=? AND deleted_at IS NULL AND ((author_user_id=? AND space_type='personal') OR (couple_id=? AND space_type='shared') OR (couple_id=? AND space_type='personal' AND visibility='partner' AND author_user_id<>?))", lookup.MomentID, lookup.ViewerID, lookup.RelationshipID, lookup.RelationshipID, lookup.ViewerID)
	moment, err := r.scanMoment(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrMomentNotFound
	}
	return moment, err
}
func (r MySQLRepository) Delete(ctx context.Context, momentID int64) error {
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
