package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSongExists   = errors.New("song already exists")
	ErrSongNotFound = errors.New("song not found")

	ErrSongNameIsNull         = errors.New("song name is null")
	ErrSongGroupIsNull        = errors.New("song group is null")
	ErrSongNameAndGroupIsNull = errors.New("song name and group is null")

	ErrInvalidSongID    = errors.New("invalid song ID")
	ErrInvalidSongName  = errors.New("invalid song name")
	ErrInvalidSongGroup = errors.New("invalid song group")
	ErrInvalidSongText  = errors.New("invalid song text")
)

type SongInfo SongSearch

type SongSearch struct {
	ID    uuid.UUID
	Name  string
	Group string
}

type Song struct {
	ID          uuid.UUID
	Name        string
	Group       string
	Text        string
	Link        string
	ReleaseDate time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
