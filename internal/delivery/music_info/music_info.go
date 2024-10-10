package musicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"songLibrary/internal/domain"
	"time"

	"github.com/google/uuid"
)

type SongResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	ReleaseDate time.Time `json:"release_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type IMusicInfo interface {
	FetchMusicInfo(ctx context.Context, name, group string) (*domain.Song, error)
}

type MusicInfo struct {
	BaseURL string
	Client  *http.Client
}

func NewMusicInfo(baseURL string) *MusicInfo {
	return &MusicInfo{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (api *MusicInfo) FetchMusicInfo(ctx context.Context, song *domain.SongInfo) (*domain.Song, error) {
	url := fmt.Sprintf("%s/info?group=%s&song=%s", api.BaseURL, song.Group, song.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch song details, status code: %d", resp.StatusCode)
	}

	var songResponse SongResponse
	if err := json.NewDecoder(resp.Body).Decode(&songResponse); err != nil {
		return nil, err
	}

	return MustConvertResponseToSong(&songResponse), nil
}

func ConvertResponseToSong(response *SongResponse) (*domain.Song, error) {
	if response.Name == "" {
		return nil, domain.ErrInvalidSongName
	}

	if response.Group == "" {
		return nil, domain.ErrInvalidSongGroup
	}

	if response.Text == "" {
		return nil, domain.ErrInvalidSongText
	}

	song := &domain.Song{
		Name:        response.Name,
		Group:       response.Group,
		Text:        response.Text,
		Link:        response.Link,
		ReleaseDate: response.ReleaseDate,
	}

	return song, nil
}

func MustConvertResponseToSong(songResponse *SongResponse) *domain.Song {
	song, _ := ConvertResponseToSong(songResponse)
	return song
}
