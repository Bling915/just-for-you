package domain

import (
	"strings"
	"unicode/utf8"
)

type User struct {
	id          int64
	openID      string
	displayName string
	avatarRef   string
}

func NewUser(id int64, openID string) User { return User{id: id, openID: openID} }

func RestoreUser(id int64, openID, displayName, avatarRef string) User {
	return User{id: id, openID: openID, displayName: displayName, avatarRef: avatarRef}
}

func (u *User) UpdateProfile(name, avatar string) error {
	name = strings.TrimSpace(name)
	if count := utf8.RuneCountInString(name); count < 1 || count > 30 {
		return ErrInvalidNickname
	}
	u.displayName = name
	u.avatarRef = avatar
	return nil
}

func (u User) ID() int64           { return u.id }
func (u User) OpenID() string      { return u.openID }
func (u User) DisplayName() string { return u.displayName }
func (u User) AvatarRef() string   { return u.avatarRef }
