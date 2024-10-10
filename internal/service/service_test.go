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
	"songLibrary/pkg/logger/logger/handlers/slogdiscard"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Add_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocking repository and music info
	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Mocked service
	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	// Input song
	song := &domain.Song{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock the FetchMusicInfo call
	mockMusicInfo.EXPECT().FetchMusicInfo(gomock.Any(), gomock.Any()).Return(&domain.Song{
		Text:        "Lyrics",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}, nil)

	// Mock the repository Create call
	mockRepo.EXPECT().Create(gomock.Any(), song).Return(nil)

	// Call the Add method
	err := svc.Add(context.Background(), song)
	assert.NoError(t, err)
	assert.Equal(t, "Lyrics", song.Text)
	assert.NotZero(t, song.CreatedAt)
}

func TestService_Add_SongAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	song := &domain.Song{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock FetchMusicInfo call
	mockMusicInfo.EXPECT().FetchMusicInfo(gomock.Any(), gomock.Any()).Return(&domain.Song{
		Text:        "Lyrics",
		ReleaseDate: time.Now(),
		Link:        "https://example.com",
	}, nil)

	// Mock repository Create call that returns song already exists error
	mockRepo.EXPECT().Create(gomock.Any(), song).Return(domain.ErrSongExists)

	// Call Add method
	err := svc.Add(context.Background(), song)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrSongExists))
}

func TestService_Add_FetchMusicInfoFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	song := &domain.Song{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock FetchMusicInfo call that fails
	mockMusicInfo.EXPECT().FetchMusicInfo(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to fetch info"))

	// Call Add method
	err := svc.Add(context.Background(), song)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch info")
}

func TestService_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock repository Read call
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(&domain.Song{
		Name:  "Song Name",
		Group: "Group Name",
		Text:  "Lyrics",
	}, nil)

	// Call Get method
	result, err := svc.Get(context.Background(), songInfo)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Lyrics", result.Text)
}

func TestService_Get_SongNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Non-existing Song",
		Group: "Non-existing Group",
	}

	// Mock repository Read call that returns song not found error
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(nil, domain.ErrSongNotFound)

	// Call Get method
	result, err := svc.Get(context.Background(), songInfo)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrSongNotFound))
}

func TestService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock repository Read call
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(&domain.Song{
		Name:  "Song Name",
		Group: "Group Name",
		Text:  "Old Lyrics",
	}, nil)

	// Mock repository Update call
	mockRepo.EXPECT().Update(gomock.Any(), songInfo, gomock.Any()).Return(nil)

	// Call Update method
	err := svc.Update(context.Background(), songInfo)
	assert.NoError(t, err)
}

func TestService_Update_SongNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Non-existing Song",
		Group: "Non-existing Group",
	}

	// Mock repository Read call
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(nil, domain.ErrSongNotFound)

	// Call Update method
	err := svc.Update(context.Background(), songInfo)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrSongNotFound))
}

func TestService_GetPaginatedText_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock repository Read call
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(&domain.Song{
		Text: "Verse 1\n\nVerse 2\n\nVerse 3",
	}, nil)

	// Call GetPaginatedText method
	verses, err := svc.GetPaginatedText(context.Background(), songInfo)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Verse 1", "Verse 2", "Verse 3"}, verses)
}

func TestService_GetPaginatedText_EmptyText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockMusicInfo := mocks.NewMockMusicInfo(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	svc := service.NewService(mockRepo, mockMusicInfo, mockLog)

	songInfo := &domain.SongInfo{
		Name:  "Song Name",
		Group: "Group Name",
	}

	// Mock repository Read call
	mockRepo.EXPECT().Read(gomock.Any(), songInfo).Return(&domain.Song{
		Text: "",
	}, nil)

	// Call GetPaginatedText method
	verses, err := svc.GetPaginatedText(context.Background(), songInfo)
	assert.Error(t, err)
	assert.Nil(t, verses)
	assert.Contains(t, err.Error(), "song text is empty")
}
