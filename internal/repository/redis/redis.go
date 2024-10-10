package redi

import (
	"context"
	"encoding/json"
	"fmt"
	"songLibrary/internal/domain"

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

func (r *Redis) genKey(group, name string) (string, error) {
	switch {
	case group != "" && name == "":
		return "", domain.ErrSongNameIsNull
	case group == "" && name != "":
		return "", domain.ErrSongGroupIsNull
	case group == "" && name == "":
		return "", domain.ErrSongNameAndGroupIsNull
	default:
		return "song:" + group + "-" + name, nil
	}
}

func (r *Redis) Set(ctx context.Context, song *domain.Song) error {
	const op = "repository.Redis.Set"

	songJSON, err := json.Marshal(song)
	if err != nil {
		return fmt.Errorf("%s: could not marshal song to JSON: %w", op, err)
	}

	key, err := r.genKey(song.Group, song.Name)
	if err != nil {
		return fmt.Errorf("%s: could not generate key for song: %w", op, err)
	}

	err = r.cache.Set(ctx, key, songJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: could not set song JSON in Redis: %w", op, err)
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, group, name string) (*domain.Song, error) {
	const op = "repository.Redis.Get"

	key, err := r.genKey(group, name)
	if err != nil {
		return nil, fmt.Errorf("%s: could not generate key for song: %w", op, err)
	}

	songJSON, err := r.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("%s: song not found in Redis cache: %w", op, domain.ErrSongNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("%s: could not get song from Redis: %w", op, err)
	}

	var song *domain.Song
	err = json.Unmarshal([]byte(songJSON), &song)
	if err != nil {
		return nil, fmt.Errorf("%s: could not unmarshal JSON into song: %w", op, err)
	}

	return song, nil
}

func (r *Redis) Invalidate(ctx context.Context, group, name string) error {
	const op = "repository.Redis.Invalidate"

	key, err := r.genKey(group, name)
	if err != nil {
		return fmt.Errorf("%s: could not generate key for song: %w", op, err)
	}

	err = r.cache.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: could not delete song from Redis: %w", op, err)
	}

	return nil
}
