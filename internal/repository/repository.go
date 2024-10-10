package repository

import (
	"context"
	"songLibrary/internal/domain"
)

type Database interface {
	Create(ctx context.Context, song *domain.Song) error
	Read(ctx context.Context, song *domain.SongSearch) (*domain.Song, error)
	Update(ctx context.Context, songSearch *domain.SongSearch, updatedSong *domain.Song) error
	Delete(ctx context.Context, song *domain.SongSearch) error

	ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error)
}

type Cache interface {
	Set(ctx context.Context, song *domain.Song) error
	Get(ctx context.Context, song *domain.SongSearch) (*domain.Song, error)
	Invalidate(ctx context.Context, song *domain.SongSearch) error
}

type IRepository interface {
	Create(ctx context.Context, song *domain.Song) error
	Read(ctx context.Context, song *domain.SongSearch) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongSearch) error
	Delete(ctx context.Context, song *domain.SongSearch) error

	ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error)
	CacheRecovery(ctx context.Context) error
}

type Repository struct {
	db    Database
	cache Cache
}

func NewRepository(db Database, cache Cache) *Repository {
	return &Repository{
		db:    db,
		cache: cache,
	}
}

func (r *Repository) Create(ctx context.Context, song *domain.Song) error {
	err := r.db.Create(ctx, song)
	if err != nil {
		return err
	}

	err = r.cache.Set(ctx, song)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Read(ctx context.Context, song *domain.SongSearch) (*domain.Song, error) {
	song, err := r.cache.Get(ctx, song)
	if err != nil {
		song, err = r.db.Read(ctx, song)
		if err != nil {
			return nil, err
		}

		err = r.cache.Set(ctx, song)
		if err != nil {
			return nil, err
		}

		return song, nil
	}

	return song, nil
}

func (r *Repository) ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error) {
	songs, err := r.db.ReadAllWithFilter(ctx, song, limit, offset)
	if err != nil {
		return nil, err
	}

	return songs, nil
}

func (r *Repository) Update(ctx context.Context, songSearch *domain.SongSearch, updatedSong *domain.Song) error {
	err := r.db.Update(ctx, songSearch, updatedSong)
	if err != nil {
		return err
	}

	err = r.cache.Set(ctx, songSearch)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, song *domain.SongSearch) error {
	err := r.db.Delete(ctx, song)
	if err != nil {
		return err
	}

	err = r.cache.Invalidate(ctx, song)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CacheRecovery(ctx context.Context) error {
	songs, err := r.db.ReadAllWithFilter(ctx, &domain.Song{}, 0, 0)
	if err != nil {
		return err
	}

	for _, song := range songs {
		err = r.cache.Set(ctx, song)
		if err != nil {
			return err
		}
	}

	return nil
}
