package repository

import (
	"context"
	"log/slog"
	"songLibrary/internal/domain"
	"songLibrary/pkg/logger/sl"
)

type Database interface {
	Create(ctx context.Context, song *domain.Song) error
	Read(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error
	Delete(ctx context.Context, song *domain.SongInfo) error

	ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error)
}

type Cache interface {
	Set(ctx context.Context, song *domain.Song) error
	Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Invalidate(ctx context.Context, song *domain.SongInfo) error
}

type IRepository interface {
	Create(ctx context.Context, song *domain.Song) error
	Read(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error
	Delete(ctx context.Context, song *domain.SongInfo) error

	ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error)
	CacheRecovery(ctx context.Context) error
}

type Repository struct {
	db    Database
	cache Cache
	log   *slog.Logger
}

func NewRepository(db Database, cache Cache, log *slog.Logger) *Repository {
	return &Repository{
		db:    db,
		cache: cache,
		log:   log,
	}
}

func (r *Repository) Create(ctx context.Context, song *domain.Song) error {
	const op = "Repository.Create"

	log := r.log.With(slog.String("op", op), slog.String("song_name", song.Name), slog.String("group_name", song.Group))

	log.Debug("creating song in database")
	err := r.db.Create(ctx, song)
	if err != nil {
		log.Error("failed to create song in database", sl.Err(err))
		return err
	}

	log.Debug("storing song in cache")
	err = r.cache.Set(ctx, song)
	if err != nil {
		log.Error("failed to store song in cache", sl.Err(err))
		return err
	}

	log.Debug("song successfully created and cached")
	return nil
}

func (r *Repository) Read(ctx context.Context, song *domain.SongInfo) (*domain.Song, error) {
	const op = "Repository.Read"

	log := r.log.With(slog.String("op", op), slog.String("song_name", song.Name), slog.String("group_name", song.Group))

	log.Debug("attempting to fetch song from cache")
	targetSong, err := r.cache.Get(ctx, song)
	if err != nil {
		log.Warn("song not found in cache, fetching from database", sl.Err(err))

		targetSong, err = r.db.Read(ctx, song)
		if err != nil {
			log.Error("failed to fetch song from database", sl.Err(err))
			return nil, err
		}

		log.Debug("storing song in cache after fetching from database")
		err = r.cache.Set(ctx, targetSong)
		if err != nil {
			log.Error("failed to store song in cache", sl.Err(err))
			return nil, err
		}

		return targetSong, nil
	}

	log.Debug("song successfully fetched from cache")
	return targetSong, nil
}

func (r *Repository) ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error) {
	const op = "Repository.ReadAllWithFilter"

	log := r.log.With(slog.String("op", op), slog.String("song_name", song.Name), slog.String("group_name", song.Group))

	log.Debug("attempting to fetch songs from database with filter")
	songs, err := r.db.ReadAllWithFilter(ctx, song, limit, offset)
	if err != nil {
		log.Error("failed to fetch songs from database with filter", sl.Err(err))
		return nil, err
	}

	log.Debug("songs successfully fetched from database")
	return songs, nil
}

func (r *Repository) Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error {
	const op = "Repository.Update"

	log := r.log.With(slog.String("op", op), slog.String("song_name", updatedSong.Name), slog.String("group_name", updatedSong.Group))

	log.Debug("updating song in database")
	err := r.db.Update(ctx, song, updatedSong)
	if err != nil {
		log.Error("failed to update song in database", sl.Err(err))
		return err
	}

	log.Debug("updating song in cache")
	err = r.cache.Set(ctx, updatedSong)
	if err != nil {
		log.Error("failed to update song in cache", sl.Err(err))
		return err
	}

	log.Debug("song successfully updated in database and cache")
	return nil
}

func (r *Repository) Delete(ctx context.Context, song *domain.SongInfo) error {
	const op = "Repository.Delete"

	log := r.log.With(slog.String("op", op), slog.String("song_id", song.ID.String()))

	log.Debug("deleting song from database")
	err := r.db.Delete(ctx, song)
	if err != nil {
		log.Error("failed to delete song from database", sl.Err(err))
		return err
	}

	log.Debug("invalidating song in cache")
	err = r.cache.Invalidate(ctx, song)
	if err != nil {
		log.Error("failed to invalidate song in cache", sl.Err(err))
		return err
	}

	log.Debug("song successfully deleted from database and cache invalidated")
	return nil
}

func (r *Repository) CacheRecovery(ctx context.Context) error {
	const op = "Repository.CacheRecovery"

	log := r.log.With(slog.String("op", op))

	log.Debug("attempting to recover cache from database")
	songs, err := r.db.ReadAllWithFilter(ctx, &domain.Song{}, 0, 0)
	if err != nil {
		log.Error("failed to fetch songs from database for cache recovery", sl.Err(err))
		return err
	}

	for _, song := range songs {
		log.Debug("caching song", slog.String("song_name", song.Name), slog.String("group_name", song.Group))
		err = r.cache.Set(ctx, song)
		if err != nil {
			log.Error("failed to cache song", sl.Err(err))
			return err
		}
	}

	log.Debug("cache recovery completed successfully")
	return nil
}
