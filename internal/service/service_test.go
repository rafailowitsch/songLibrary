package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"songLibrary/internal/domain"
	"songLibrary/internal/service"
	"songLibrary/internal/service/mocks"
	"songLibrary/pkg/logger/handlers/slogdiscard"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_Add_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	song := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}

	mockMusicInfo.EXPECT().FetchMusicInfo(gomock.Any(), songInfo).Return(song, nil)
	mockRepo.EXPECT().Create(gomock.Any(), song).Return(nil)

	err := service.Add(context.Background(), songInfo)
	assert.NoError(t, err)
}
func TestService_Add_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	song := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}

	mockMusicInfo.EXPECT().FetchMusicInfo(gomock.Any(), songInfo).Return(song, nil)
	mockRepo.EXPECT().Create(gomock.Any(), song).Return(domain.ErrSongExists)

	err := service.Add(context.Background(), songInfo)
	assert.ErrorIs(t, err, domain.ErrSongExists)
}

func TestService_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	song := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}

	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(song, nil)

	result, err := service.Get(context.Background(), songInfo)
	assert.NoError(t, err)
	assert.Equal(t, song, result)
}

func TestService_Get_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(nil, domain.ErrSongNotFound)

	_, err := service.Get(context.Background(), songInfo)
	assert.ErrorIs(t, err, domain.ErrSongNotFound)
}

func TestService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	originalSong := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}

	updatedSong := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "Updated text",
		ReleaseDate: time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(originalSong, nil)
	mockRepo.EXPECT().Update(gomock.Any(), songInfo, updatedSong).Return(nil)

	err := service.Update(context.Background(), songInfo, updatedSong)
	assert.NoError(t, err)
}

func TestService_Update_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	originalSong := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}

	updatedSong := &domain.Song{}

	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(originalSong, nil)
	mockRepo.EXPECT().Update(gomock.Any(), songInfo, gomock.Any()).Return(domain.ErrSongNotFound)

	err := service.Update(context.Background(), songInfo, updatedSong)
	assert.ErrorIs(t, err, domain.ErrSongNotFound)
}

func TestService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	mockRepo.EXPECT().Delete(gomock.Any(), songInfo).Return(nil)

	err := service.Delete(context.Background(), songInfo)
	assert.NoError(t, err)
}

func TestService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	service := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	mockRepo.EXPECT().Delete(gomock.Any(), songInfo).Return(domain.ErrSongNotFound)

	err := service.Delete(context.Background(), songInfo)
	assert.ErrorIs(t, err, domain.ErrSongNotFound)
}

func TestService_GetAllWithFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, nil, mockLog)

	songFilter := &domain.Song{
		Name:  "Hysteria",
		Group: "Muse",
	}

	expectedSongs := []*domain.Song{
		{
			ID:          uuid.New(),
			Name:        "Hysteria",
			Group:       "Muse",
			Text:        "It's bugging me...",
			Link:        "https://link-to-song1.com",
			ReleaseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize

	// Ожидаем вызов метода ReadAllWithFilter репозитория
	mockRepo.EXPECT().
		ReadAllWithFilter(gomock.Any(), songFilter, pageSize, offset).
		Return(expectedSongs, nil)

	// Выполняем тестируемую функцию
	songs, err := svc.GetAllWithFilter(context.Background(), songFilter, page, pageSize)

	assert.NoError(t, err)
	assert.Len(t, songs, 1)
	assert.Equal(t, expectedSongs, songs)
}

func TestService_GetAllWithFilter_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, nil, mockLog)

	songFilter := &domain.Song{
		Name:  "Hysteria",
		Group: "Muse",
	}

	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize

	// Ожидаем, что репозиторий вернет ошибку
	mockRepo.EXPECT().
		ReadAllWithFilter(gomock.Any(), songFilter, pageSize, offset).
		Return(nil, errors.New("database error"))

	// Выполняем тестируемую функцию
	songs, err := svc.GetAllWithFilter(context.Background(), songFilter, page, pageSize)

	assert.Error(t, err)
	assert.Nil(t, songs)
	assert.Contains(t, err.Error(), "failed to fetch songs with filter")
}

func TestService_GetPaginatedText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, nil, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	expectedSong := &domain.Song{
		ID:    uuid.New(),
		Name:  "Hysteria",
		Group: "Muse",
		Text:  "It's bugging me...\n\nI can't control...",
	}

	// Ожидаем вызов метода Read репозитория
	mockRepo.EXPECT().
		Read(gomock.Any(), songInfo).
		Return(expectedSong, nil)

	// Выполняем тестируемую функцию
	verses, err := svc.GetPaginatedText(context.Background(), songInfo)

	assert.NoError(t, err)
	assert.Len(t, verses, 2)
	assert.Equal(t, []string{
		"It's bugging me...",
		"I can't control...",
	}, verses)
}

func TestService_GetPaginatedText_EmptyText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, nil, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	expectedSong := &domain.Song{
		ID:    uuid.New(),
		Name:  "Hysteria",
		Group: "Muse",
		Text:  "",
	}

	// Ожидаем вызов метода Read репозитория
	mockRepo.EXPECT().
		Read(gomock.Any(), songInfo).
		Return(expectedSong, nil)

	// Выполняем тестируемую функцию
	verses, err := svc.GetPaginatedText(context.Background(), songInfo)

	assert.Error(t, err)
	assert.Nil(t, verses)
	assert.Contains(t, err.Error(), "song text is empty")
}

func TestService_GetPaginatedText_SongNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, nil, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Hysteria",
		Group: "Muse",
	}

	// Ожидаем, что репозиторий вернет ошибку, что песня не найдена
	mockRepo.EXPECT().
		Read(gomock.Any(), songInfo).
		Return(nil, domain.ErrSongNotFound)

	// Выполняем тестируемую функцию
	verses, err := svc.GetPaginatedText(context.Background(), songInfo)

	assert.Error(t, err)
	assert.Nil(t, verses)
	assert.Contains(t, err.Error(), "song not found")
}
