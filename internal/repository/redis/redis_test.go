package redi_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"songLibrary/internal/domain"
	redi "songLibrary/internal/repository/redis"

	"github.com/google/uuid"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

// func TestRedis_GenKey(t *testing.T) {
// 	redisClient, _ := redismock.NewClientMock()
// 	r := redis.NewRedis(redisClient)

// 	ctx := context.Background()

// 	// Успешная генерация ключа
// 	key, err := r.GenKey(ctx, "Muse", "Hysteria")
// 	assert.NoError(t, err)
// 	assert.Equal(t, "song:Muse-Hysteria", key)

// 	// Ошибки при некорректных параметрах
// 	_, err = r.GenKey(ctx, "Muse", "")
// 	assert.ErrorIs(t, err, domain.ErrSongNameIsNull)

// 	_, err = r.GenKey(ctx, "", "Hysteria")
// 	assert.ErrorIs(t, err, domain.ErrSongGroupIsNull)

// 	_, err = r.GenKey(ctx, "", "")
// 	assert.ErrorIs(t, err, domain.ErrSongNameAndGroupIsNull)
// }

func TestRedis_Set(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()
	r := redi.NewRedis(redisClient)

	ctx := context.Background()
	song := &domain.Song{
		ID:          uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		Link:        "https://link-to-song.com",
		ReleaseDate: time.Now(),
	}

	// Успешное сохранение песни в кэш
	key := "song:Muse-Hysteria"
	songJSON, _ := json.Marshal(song)
	mock.ExpectSet(key, songJSON, 0).SetVal("OK")

	err := r.Set(ctx, song)
	assert.NoError(t, err)

	// Ошибка при сериализации данных
	song.Name = string([]byte{0xff, 0xfe, 0xfd}) // Невалидные байты
	err = r.Set(ctx, song)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not marshal song to JSON")

	// Ошибка при сохранении данных в Redis
	song.Name = "Hysteria" // Корректируем данные
	mock.ExpectSet(key, songJSON, 0).SetErr(errors.New("redis error"))
	err = r.Set(ctx, song)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not set song JSON in Redis")
}

// func TestRedis_Get(t *testing.T) {
// 	redisClient, mock := redismock.NewClientMock()
// 	r := redis.NewRedis(redisClient)

// 	ctx := context.Background()
// 	key := "song:Muse-Hysteria"

// 	song := &domain.Song{
// 		ID:          "12345",
// 		Name:        "Hysteria",
// 		Group:       "Muse",
// 		Text:        "It's bugging me...",
// 		Link:        "https://link-to-song.com",
// 		ReleaseDate: time.Now(),
// 	}
// 	songJSON, _ := json.Marshal(song)

// 	// Успешное получение песни из кэша
// 	mock.ExpectGet(key).SetVal(string(songJSON))

// 	cachedSong, err := r.Get(ctx, "Muse", "Hysteria")
// 	assert.NoError(t, err)
// 	assert.Equal(t, song, cachedSong)

// 	// Ошибка: песня не найдена в кэше
// 	mock.ExpectGet(key).RedisNil()

// 	cachedSong, err = r.Get(ctx, "Muse", "NonExistingSong")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "song not found in Redis cache")
// 	assert.Nil(t, cachedSong)

// 	// Ошибка: некорректные данные в Redis (ошибка десериализации)
// 	mock.ExpectGet(key).SetVal("invalid JSON")

// 	cachedSong, err = r.Get(ctx, "Muse", "Hysteria")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "could not unmarshal JSON into song")
// 	assert.Nil(t, cachedSong)

// 	// Ошибка при получении данных из Redis
// 	mock.ExpectGet(key).SetErr(errors.New("redis error"))

// 	cachedSong, err = r.Get(ctx, "Muse", "Hysteria")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "could not get song from Redis")
// 	assert.Nil(t, cachedSong)
// }

// func TestRedis_Invalidate(t *testing.T) {
// 	redisClient, mock := redismock.NewClientMock()
// 	r := redis.NewRedis(redisClient)

// 	ctx := context.Background()
// 	key := "song:Muse-Hysteria"

// 	// Успешная инвалидизация кэша
// 	mock.ExpectDel(key).SetVal(1)

// 	err := r.Invalidate(ctx, "Muse", "Hysteria")
// 	assert.NoError(t, err)

// 	// Ошибка при попытке удалить песню из кэша
// 	mock.ExpectDel(key).SetErr(errors.New("redis error"))

// 	err = r.Invalidate(ctx, "Muse", "Hysteria")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "could not delete song from Redis")
// }
