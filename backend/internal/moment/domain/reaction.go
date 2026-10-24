package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type ReactionType string

const (
	ReactionSeen     ReactionType = "seen"
	ReactionHug      ReactionType = "hug"
	ReactionTogether ReactionType = "together"
	ReactionCollect  ReactionType = "collect"
	ReactionText     ReactionType = "text"
)

func (r ReactionType) Valid() bool {
	return r == ReactionSeen || r == ReactionHug || r == ReactionTogether || r == ReactionCollect || r == ReactionText
}

type Reaction struct {
	id, momentID, userID int64
	reactionType         ReactionType
	text                 string
	createdAt, updatedAt time.Time
}

func NewReaction(momentID, userID int64, reactionType ReactionType, text string, now time.Time) (Reaction, error) {
	text = strings.TrimSpace(text)
	if momentID <= 0 || userID <= 0 || !reactionType.Valid() {
		return Reaction{}, ErrInvalidReaction
	}
	if reactionType == ReactionText {
		if text == "" {
			return Reaction{}, ErrEmptyReactionText
		}
		if utf8.RuneCountInString(text) > 120 {
			return Reaction{}, ErrReactionTextTooLong
		}
	} else if text != "" {
		return Reaction{}, ErrInvalidReaction
	}
	return Reaction{momentID: momentID, userID: userID, reactionType: reactionType, text: text, createdAt: now, updatedAt: now}, nil
}

func RestoreReaction(id, momentID, userID int64, reactionType ReactionType, text string, createdAt, updatedAt time.Time) Reaction {
	return Reaction{id: id, momentID: momentID, userID: userID, reactionType: reactionType, text: text, createdAt: createdAt, updatedAt: updatedAt}
}

func (r Reaction) ID() int64            { return r.id }
func (r Reaction) MomentID() int64      { return r.momentID }
func (r Reaction) UserID() int64        { return r.userID }
func (r Reaction) Type() ReactionType   { return r.reactionType }
func (r Reaction) Text() string         { return r.text }
func (r Reaction) CreatedAt() time.Time { return r.createdAt }
func (r Reaction) UpdatedAt() time.Time { return r.updatedAt }
func (r Reaction) Identity() string {
	return fmt.Sprintf("%s:%s:%s", strconv.FormatInt(r.momentID, 10), strconv.FormatInt(r.userID, 10), r.reactionType)
}
