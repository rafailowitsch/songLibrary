package dto

import (
	"songLibrary/internal/domain"
	"time"

	"github.com/google/uuid"
)

type AddSongRequest struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type UpdateSongRequest struct {
	Name  string `json:"name"`
	Group string `json:"group"`
	Text  string `json:"text,omitempty"`
	Link  string `json:"link,omitempty"`
}

type SongResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Text        string    `json:"text,omitempty"`
	Link        string    `json:"link,omitempty"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GetAllSongsFilter struct {
	Name        string `json:"name,omitempty"`
	Group       string `json:"group,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

type PaginatedTextResponse struct {
	Text []string `json:"text"`
}

type SongDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	ReleaseDate time.Time `json:"release_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func SongToDTO(song *domain.Song) *SongDTO {
	return &SongDTO{
		ID:          song.ID,
		Name:        song.Name,
		Group:       song.Group,
		Text:        song.Text,
		Link:        song.Link,
		ReleaseDate: song.ReleaseDate,
		CreatedAt:   song.CreatedAt,
		UpdatedAt:   song.UpdatedAt,
	}
}

func DTOToSong(dto *SongDTO) *domain.Song {
	return &domain.Song{
		ID:          dto.ID,
		Name:        dto.Name,
		Group:       dto.Group,
		Text:        dto.Text,
		Link:        dto.Link,
		ReleaseDate: dto.ReleaseDate,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}
