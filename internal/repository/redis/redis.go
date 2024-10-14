package redi

import (
	"context"
	"encoding/json"
	"fmt"
	"songLibrary/internal/domain"
	"songLibrary/internal/dto"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	cache *redis.Client
}

func NewRedis(cache *redis.Client) *Redis {
	return &Redis{
		cache: cache,
	}
}

func (r *Redis) Set(ctx context.Context, song *domain.Song) error {
	const op = "repository.Redis.Set"

	songDTO := dto.SongToDTO(song)
	songJSON, err := json.Marshal(songDTO)
	if err != nil {
		return fmt.Errorf("%s: could not marshal song to JSON: %w", op, err)
	}

	key := songDTO.ID.String()
	err = r.cache.Set(ctx, key, songJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: could not set song JSON in Redis: %w", op, err)
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error) {
	const op = "repository.Redis.Get"

	key := song.ID.String()
	songJSON, err := r.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("%s: song not found in Redis cache: %w", op, domain.ErrSongNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("%s: could not get song from Redis: %w", op, err)
	}

	var targetSong *domain.Song
	err = json.Unmarshal([]byte(songJSON), &targetSong)
	if err != nil {
		return nil, fmt.Errorf("%s: could not unmarshal JSON into song: %w", op, err)
	}

	return targetSong, nil
}

func (r *Redis) Invalidate(ctx context.Context, song *domain.SongInfo) error {
	const op = "repository.Redis.Invalidate"

	key := song.ID.String()
	err := r.cache.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: could not delete song from Redis: %w", op, err)
	}

	return nil
}
