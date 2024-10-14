package musicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"songLibrary/internal/domain"
	"songLibrary/internal/dto"
	"songLibrary/pkg/logger/sl"
)

type SongResponse dto.SongDTO

type IMusicInfo interface {
	FetchMusicInfo(ctx context.Context, name, group string) (*domain.Song, error)
}

type MusicInfo struct {
	BaseURL string
	Client  *http.Client
	log     *slog.Logger
}

func NewMusicInfo(baseURL string, log *slog.Logger) *MusicInfo {
	return &MusicInfo{
		BaseURL: baseURL,
		Client:  &http.Client{},
		log:     log,
	}
}

func (api *MusicInfo) FetchMusicInfo(ctx context.Context, song *domain.SongInfo) (*domain.Song, error) {
	const op = "MusicInfo.FetchMusicInfo"

	group := url.QueryEscape(song.Group)
	name := url.QueryEscape(song.Name)
	url := fmt.Sprintf("http://%s/info?group=%s&song=%s", api.BaseURL, group, name)

	// Добавляем логирование начала операции
	log := api.log.With(
		slog.String("op", op),
		slog.String("url", url),
		slog.String("song_name", song.Name),
		slog.String("group_name", song.Group),
	)
	log.Info("attempting to fetch song info from external API")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error("failed to create request", sl.Err(err))
		return nil, err
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		log.Error("failed to send request to external API", sl.Err(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("external API returned non-OK status", slog.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("failed to fetch song details, status code: %d", resp.StatusCode)
	}

	var songResponse SongResponse
	if err := json.NewDecoder(resp.Body).Decode(&songResponse); err != nil {
		log.Error("failed to decode response from external API", sl.Err(err))
		return nil, err
	}

	log.Info("successfully fetched song info from external API", slog.String("song_name", songResponse.Name), slog.String("group_name", songResponse.Group))

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
