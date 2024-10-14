package redi

import (
	"context"
	"encoding/json"
	"errors"
	"songLibrary/internal/domain"
	"songLibrary/internal/dto"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRedis_Set_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	// Создаем тестовые данные
	song := &domain.Song{
		ID:          uuid.New(),
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		Link:        "https://link-to-song.com",
		ReleaseDate: time.Now(),
	}

	songDTO := dto.SongToDTO(song)
	songJSON, err := json.Marshal(songDTO)
	assert.NoError(t, err)

	// Ожидаем успешный Set запрос в Redis
	mock.ExpectSet(songDTO.ID.String(), songJSON, 0).SetVal("OK")

	// Вызов метода Set
	err = r.Set(ctx, song)
	assert.NoError(t, err)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Set_Overwrite(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	// Создаем первоначальные тестовые данные
	songOriginal := &domain.Song{
		ID:          uuid.New(),
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		Link:        "https://link-to-song.com",
		ReleaseDate: time.Now(),
	}

	songDTOOriginal := dto.SongToDTO(songOriginal)
	songJSONOriginal, err := json.Marshal(songDTOOriginal)
	assert.NoError(t, err)

	// Ожидаем успешный Set запрос для первоначальных данных в Redis
	mock.ExpectSet(songDTOOriginal.ID.String(), songJSONOriginal, 0).SetVal("OK")

	// Вызов метода Set для первоначальных данных
	err = r.Set(ctx, songOriginal)
	assert.NoError(t, err)

	// Создаем новые тестовые данные для перезаписи
	songUpdated := &domain.Song{
		ID:          songOriginal.ID, // Используем тот же ID
		Name:        "New Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me again...",
		Link:        "https://new-link-to-song.com",
		ReleaseDate: time.Now(),
	}

	songDTOUpdated := dto.SongToDTO(songUpdated)
	songJSONUpdated, err := json.Marshal(songDTOUpdated)
	assert.NoError(t, err)

	// Ожидаем успешный Set запрос для обновленных данных в Redis
	mock.ExpectSet(songDTOUpdated.ID.String(), songJSONUpdated, 0).SetVal("OK")

	// Вызов метода Set для обновленных данных (перезапись)
	err = r.Set(ctx, songUpdated)
	assert.NoError(t, err)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Get_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	songID := uuid.New()

	songDTO := &dto.SongDTO{
		ID:          songID,
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		Link:        "https://link-to-song.com",
		ReleaseDate: time.Now(),
	}

	songJSON, err := json.Marshal(songDTO)
	assert.NoError(t, err)

	// Ожидаем успешный Get запрос в Redis
	mock.ExpectGet(songID.String()).SetVal(string(songJSON))

	// Вызов метода Get
	songInfo := &domain.SongInfo{ID: songID}
	song, err := r.Get(ctx, songInfo)
	assert.NoError(t, err)

	// Проверяем полученные данные
	assert.Equal(t, songDTO.Name, song.Name)
	assert.Equal(t, songDTO.Group, song.Group)
	assert.Equal(t, songDTO.Text, song.Text)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	songID := uuid.New()

	// Ожидаем, что Redis вернет Nil
	mock.ExpectGet(songID.String()).RedisNil()

	// Вызов метода Get
	songInfo := &domain.SongInfo{ID: songID}
	song, err := r.Get(ctx, songInfo)

	// Проверяем результат
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "song not found in Redis cache")
	assert.Nil(t, song)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Get_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	songID := uuid.New()

	// Ожидаем, что Redis вернет некорректные данные
	mock.ExpectGet(songID.String()).SetVal("invalid JSON")

	// Вызов метода Get
	songInfo := &domain.SongInfo{ID: songID}
	song, err := r.Get(ctx, songInfo)

	// Проверяем результат
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal JSON into song")
	assert.Nil(t, song)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Invalidate_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	songID := uuid.New()

	// Ожидаем успешный Del запрос в Redis
	mock.ExpectDel(songID.String()).SetVal(1)

	// Вызов метода Invalidate
	songInfo := &domain.SongInfo{ID: songID}
	err := r.Invalidate(ctx, songInfo)
	assert.NoError(t, err)

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedis_Invalidate_Failure(t *testing.T) {
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()

	r := NewRedis(mockRedis)

	songID := uuid.New()

	// Ожидаем, что Redis вернет ошибку
	mock.ExpectDel(songID.String()).SetErr(errors.New("some redis error"))

	// Вызов метода Invalidate
	songInfo := &domain.SongInfo{ID: songID}
	err := r.Invalidate(ctx, songInfo)

	// Проверяем результат
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not delete song from Redis")

	// Проверяем все ожидания
	assert.NoError(t, mock.ExpectationsWereMet())
}
