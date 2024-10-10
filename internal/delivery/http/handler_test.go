package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	delivery "songLibrary/internal/delivery/http"
	"songLibrary/internal/delivery/http/mocks"
	"songLibrary/internal/domain"
	"songLibrary/pkg/logger/logger/handlers/slogdiscard"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddSong_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	reqBody := delivery.SongRequest{
		Name:  "Hysteria",
		Group: "Muse",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.EXPECT().Add(gomock.Any(), &domain.SongInfo{
		Name:  reqBody.Name,
		Group: reqBody.Group,
	}).Return(nil)

	h.Add(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "song added successfully", respBody["message"])
}

func TestAddSong_Failure_DecodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewBuffer([]byte("{invalid-json")))
	w := httptest.NewRecorder()

	h.Add(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid request", respBody["error"])
}

func TestAddSong_Failure_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	reqBody := delivery.SongRequest{
		Name:  "Hysteria",
		Group: "Muse",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.EXPECT().Add(gomock.Any(), &domain.SongInfo{
		Name:  reqBody.Name,
		Group: reqBody.Group,
	}).Return(errors.New("service error"))

	h.Add(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "internal error", respBody["error"])
}

func TestGetAllWithFilter_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodGet, "/songs?page=1&page_size=2&name=Hysteria&group=Muse", nil)
	w := httptest.NewRecorder()

	expectedSongs := []*domain.Song{
		{
			ID:          uuid.New(),
			Name:        "Hysteria",
			Group:       "Muse",
			Text:        "Ooh baby, don't you know I suffer?",
			Link:        "http://example.com",
			ReleaseDate: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockService.EXPECT().GetAllWithFilter(gomock.Any(), &domain.SongSearch{
		Name:  "Hysteria",
		Group: "Muse",
	}, 1, 2).Return(expectedSongs, nil)

	h.GetAllWithFilter(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var songsResponse []delivery.SongResponse
	err := json.NewDecoder(resp.Body).Decode(&songsResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(songsResponse))
	assert.Equal(t, "Hysteria", songsResponse[0].Name)
	assert.Equal(t, "Muse", songsResponse[0].Group)
}

func TestGetAllWithFilter_InvalidPageSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	req := httptest.NewRequest(http.MethodGet, "/songs?page=1&page_size=-1", nil)
	w := httptest.NewRecorder()

	h.GetAllWithFilter(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid page_size parameter", respBody["error"])
}

func TestDeleteSong_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := delivery.NewHandler(mockService, mockLog)

	// Создаем маршрутизатор и регистрируем маршруты
	r := chi.NewRouter()
	r.Delete("/songs/{id}", h.Delete)

	songID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/songs/"+songID.String(), nil)
	w := httptest.NewRecorder()

	mockService.EXPECT().Delete(gomock.Any(), &domain.SongInfo{
		ID: songID,
	}).Return(nil)

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "song deleted successfully", respBody["message"])
}

func TestDeleteSong_SongNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Инициализация обработчика и маршрутизатора
	h := delivery.NewHandler(mockService, mockLog)
	r := chi.NewRouter()
	r.Delete("/songs/{id}", h.Delete)

	songID := uuid.New()

	// Запрос для удаления
	req := httptest.NewRequest(http.MethodDelete, "/songs/"+songID.String(), nil)
	w := httptest.NewRecorder()

	// Ожидание вызова метода Delete с ошибкой "song not found"
	mockService.EXPECT().Delete(gomock.Any(), &domain.SongInfo{
		ID: songID,
	}).Return(domain.ErrSongNotFound)

	// Выполнение запроса через маршрутизатор
	r.ServeHTTP(w, req)

	// Проверка ответа
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "song not found", respBody["error"])
}

func TestGetPaginatedText_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Инициализация обработчика и маршрутизатора
	h := delivery.NewHandler(mockService, mockLog)
	r := chi.NewRouter()
	r.Get("/songs/{id}/text", h.GetPaginatedText)

	songID := uuid.New()
	expectedVerses := []string{
		"Ooh baby, don't you know I suffer?",
		"Ooh baby, can you hear me moan?",
	}

	// Запрос для получения пагинированного текста
	req := httptest.NewRequest(http.MethodGet, "/songs/"+songID.String()+"/text", nil)
	w := httptest.NewRecorder()

	// Ожидание вызова метода GetPaginatedText
	mockService.EXPECT().GetPaginatedText(gomock.Any(), &domain.SongInfo{
		ID: songID,
	}).Return(expectedVerses, nil)

	// Выполнение запроса через маршрутизатор
	r.ServeHTTP(w, req)

	// Проверка ответа
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody delivery.PaginatedTextResponse
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedVerses, respBody.Text)
}

func TestGetPaginatedText_SongNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	// Инициализация обработчика и маршрутизатора
	h := delivery.NewHandler(mockService, mockLog)
	r := chi.NewRouter()
	r.Get("/songs/{id}/text", h.GetPaginatedText)

	songID := uuid.New()

	// Запрос для получения пагинированного текста
	req := httptest.NewRequest(http.MethodGet, "/songs/"+songID.String()+"/text", nil)
	w := httptest.NewRecorder()

	// Ожидание вызова метода GetPaginatedText с ошибкой "song not found"
	mockService.EXPECT().GetPaginatedText(gomock.Any(), &domain.SongInfo{
		ID: songID,
	}).Return(nil, domain.ErrSongNotFound)

	// Выполнение запроса через маршрутизатор
	r.ServeHTTP(w, req)

	// Проверка ответа
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "song not found", respBody["error"])
}
