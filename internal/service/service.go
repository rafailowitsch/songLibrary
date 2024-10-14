package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"songLibrary/internal/domain"
	"songLibrary/pkg/logger/sl"
	"strings"
	"time"
)

type Repository interface {
	Create(ctx context.Context, song *domain.Song) error
	Read(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error
	Delete(ctx context.Context, song *domain.SongInfo) error

	ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error)
}

type MusicInfo interface {
	FetchMusicInfo(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
}

type IService interface {
	Add(ctx context.Context, song *domain.SongInfo) error
	Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error
	Delete(ctx context.Context, song *domain.SongInfo) error

	GetAllWithFilter(ctx context.Context, song *domain.Song, page, pageSize int) ([]*domain.Song, error)
	GetPaginatedText(ctx context.Context, song *domain.SongInfo) ([]string, error)
}

type Service struct {
	Repo      Repository
	MusicInfo MusicInfo
	log       *slog.Logger
}

func NewService(r Repository, mi MusicInfo, log *slog.Logger) *Service {
	return &Service{
		Repo:      r,
		MusicInfo: mi,
		log:       log,
	}
}

// Add method to add a new song to the system.
func (s *Service) Add(ctx context.Context, songInfo *domain.SongInfo) error {
	const op = "Service.Add"

	log := s.log.With(
		slog.String("op", op),
		slog.String("song_name", songInfo.Name),
		slog.String("group_name", songInfo.Group),
	)

	log.Info("attempting to add a new song")

	// Fetch music info from external API
	// Fetch music info from external API
	song, err := s.MusicInfo.FetchMusicInfo(ctx, songInfo)
	if err != nil {
		var httpErr *domain.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusBadRequest {
			// Log and return a special error for bad request from MusicInfo
			log.Warn("failed to fetch song info: bad request from MusicInfo", sl.Err(err))
			return fmt.Errorf("%s: bad request from MusicInfo: %w", op, err)
		}
		log.Error("failed to fetch song info", sl.Err(err))
		return fmt.Errorf("%s: failed to fetch song info: %w", op, err)
	}

	log.Debug("fetched song info successfully")

	// Save the song to the repository
	err = s.Repo.Create(ctx, song)
	if err != nil {
		if errors.Is(err, domain.ErrSongExists) {
			log.Warn("song already exists", sl.Err(err))
			return fmt.Errorf("%s: song already exists: %w", op, domain.ErrSongExists)
		}
		log.Error("failed to save song", sl.Err(err))
		return fmt.Errorf("%s: failed to save song: %w", op, err)
	}

	log.Info("song successfully added")
	return nil
}

// Get method to fetch a song by group and name.
func (s *Service) Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error) {
	const op = "Service.Get"

	log := s.log.With(
		slog.String("op", op),
		slog.String("song_name", song.Name),
		slog.String("group_name", song.Group),
	)

	log.Info("attempting to fetch song")

	// Try to get the song from the repository
	targetSong, err := s.Repo.Read(ctx, song)
	if err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Warn("song not found", sl.Err(err))
			return nil, fmt.Errorf("%s: song not found: %w", op, domain.ErrSongNotFound)
		}
		log.Error("failed to read song", sl.Err(err))
		return nil, fmt.Errorf("%s: failed to read song: %w", op, err)
	}

	log.Info("song successfully fetched")
	return targetSong, nil
}

// Update method to update an existing song's information.
func (s *Service) Update(ctx context.Context, songInfo *domain.SongInfo, updatedSong *domain.Song) error {
	const op = "Service.Update"

	log := s.log.With(
		slog.String("op", op),
		slog.String("song_name", songInfo.Name),
		slog.String("group_name", songInfo.Group),
	)

	log.Info("attempting to update song")

	// Fetch the existing song information
	targetSong, err := s.Get(ctx, songInfo)
	if err != nil {
		log.Error("failed to fetch song", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	// Merge updatedSong with targetSong
	mergedSong := mergeSongs(updatedSong, targetSong)

	// Update the song in the repository
	err = s.Repo.Update(ctx, songInfo, mergedSong)
	if err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Warn("song not found during update", sl.Err(err))
			return fmt.Errorf("%s: song not found: %w", op, domain.ErrSongNotFound)
		}
		log.Error("failed to update song", sl.Err(err))
		return fmt.Errorf("%s: failed to update song: %w", op, err)
	}

	log.Info("song successfully updated")
	return nil
}

// Delete method to remove a song from the system.
func (s *Service) Delete(ctx context.Context, songSearch *domain.SongInfo) error {
	const op = "Service.Delete"

	log := s.log.With(
		slog.String("op", op),
		slog.String("song_name", songSearch.Name),
		slog.String("group_name", songSearch.Group),
	)

	log.Info("attempting to delete song")

	// Delete the song from the repository
	err := s.Repo.Delete(ctx, songSearch)
	if err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Warn("song not found during deletion", sl.Err(err))
			return fmt.Errorf("%s: song not found: %w", op, domain.ErrSongNotFound)
		}
		log.Error("failed to delete song", sl.Err(err))
		return fmt.Errorf("%s: failed to delete song: %w", op, err)
	}

	log.Info("song successfully deleted")
	return nil
}

// GetAllWithFilter retrieves all songs with filtering and pagination.
func (s *Service) GetAllWithFilter(ctx context.Context, song *domain.Song, page, pageSize int) ([]*domain.Song, error) {
	const op = "Service.GetAllWithFilter"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("page", page),
		slog.Int("pageSize", pageSize),
	)

	offset := (page - 1) * pageSize
	log.Info("attempting to fetch songs with filter", slog.Int("offset", offset))

	// Fetch songs with filtering from the repository
	songs, err := s.Repo.ReadAllWithFilter(ctx, song, pageSize, offset)
	if err != nil {
		log.Error("failed to fetch songs with filter", sl.Err(err))
		return nil, fmt.Errorf("%s: failed to fetch songs with filter: %w", op, err)
	}

	log.Info("songs successfully fetched", slog.Int("count", len(songs)))
	return songs, nil
}

// GetPaginatedText retrieves the song's text with pagination by verses.
func (s *Service) GetPaginatedText(ctx context.Context, song *domain.SongInfo) ([]string, error) {
	const op = "Service.GetPaginatedText"

	log := s.log.With(
		slog.String("op", op),
		slog.String("song_name", song.Name),
		slog.String("group_name", song.Group),
	)

	log.Info("attempting to fetch and paginate song text")

	// Try to get the song from the repository
	targetSong, err := s.Repo.Read(ctx, song)
	if err != nil {
		if err == domain.ErrSongNotFound {
			log.Warn("song not found", sl.Err(err))
			return nil, fmt.Errorf("%s: song not found: %w", op, domain.ErrSongNotFound)
		}
		log.Error("failed to fetch song from repository", sl.Err(err))
		return nil, fmt.Errorf("%s: failed to fetch song: %w", op, err)
	}

	if targetSong.Text == "" {
		log.Warn("song text is empty", slog.String("song_name", targetSong.Name), slog.String("group_name", targetSong.Group))
		return nil, fmt.Errorf("%s: song text is empty", op)
	}

	// Split the song text into verses by detecting double newlines (\n\n)
	verses := strings.Split(targetSong.Text, "\n\n")

	log.Debug("successfully paginated song text", slog.Int("verses_count", len(verses)))

	log.Info("song text successfully paginated", slog.String("song_name", targetSong.Name), slog.Int("verses_count", len(verses)))

	return verses, nil
}

func mergeSongs(updatedSong, targetSong *domain.Song) *domain.Song {
	updatedSong.ID = targetSong.ID

	if updatedSong.Name == "" {
		updatedSong.Name = targetSong.Name
	}
	if updatedSong.Group == "" {
		updatedSong.Group = targetSong.Group
	}
	if updatedSong.Text == "" {
		updatedSong.Text = targetSong.Text
	}
	if updatedSong.Link == "" {
		updatedSong.Link = targetSong.Link
	}
	if updatedSong.ReleaseDate.IsZero() {
		updatedSong.ReleaseDate = targetSong.ReleaseDate
	}
	if updatedSong.CreatedAt.IsZero() {
		updatedSong.CreatedAt = targetSong.CreatedAt
	}
	// UpdatedAt устанавливаем заново для актуального времени
	updatedSong.UpdatedAt = time.Now()

	return updatedSong
}
