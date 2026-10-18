package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalidProfile = errors.New("invalid profile")

type User struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
	AvatarRef   string `json:"avatarRef"`
}
type UpdateInput struct {
	DisplayName string `json:"displayName"`
	AvatarRef   string `json:"avatarRef"`
}

func ValidateUpdate(in UpdateInput) (UpdateInput, error) {
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if n := utf8.RuneCountInString(in.DisplayName); n < 1 || n > 30 {
		return UpdateInput{}, ErrInvalidProfile
	}
	return in, nil
}

type Service struct{ DB *sql.DB }

func (s Service) UpdateMe(ctx context.Context, id int64, in UpdateInput) (User, error) {
	in, e := ValidateUpdate(in)
	if e != nil {
		return User{}, e
	}
	_, e = s.DB.ExecContext(ctx, "UPDATE users SET display_name=?,avatar_ref=? WHERE id=?", in.DisplayName, in.AvatarRef, id)
	return User{ID: id, DisplayName: in.DisplayName, AvatarRef: in.AvatarRef}, e
}
