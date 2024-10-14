package deliveryHttp_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	handler "songLibrary/internal/delivery/http"
	"songLibrary/internal/delivery/http/mocks"
	"songLibrary/internal/domain"
	"songLibrary/internal/dto"
	"songLibrary/pkg/logger/handlers/slogdiscard"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddSong_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := handler.NewHandler(mockService, mockLog)

	reqBody := dto.AddSongRequest{
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

func TestAddSong_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := handler.NewHandler(mockService, mockLog)

	// Запрос без поля name и group
	req := httptest.NewRequest(http.MethodPost, "/songs", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()

	h.Add(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "name and group are required", respBody["error"])
}

func TestAddSong_Failure_DecodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockLog := slog.New(slogdiscard.NewDiscardHandler())

	h := handler.NewHandler(mockService, mockLog)

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

	h := handler.NewHandler(mockService, mockLog)

	reqBody := dto.AddSongRequest{
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

func TestHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout))

	handler := NewHandler(mockService, logger)
	songID := uuid.New()
	reqBody := `{"name": "Updated Song", "group": "Updated Group"}`
	req := httptest.NewRequest(http.MethodPut, "/songs/"+songID.String(), strings.NewReader(reqBody))
	req = mux.SetURLVars(req, map[string]string{"id": songID.String()})
	w := httptest.NewRecorder()

	mockService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	handler.Update(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "song updated successfully")
}

func TestHandler_Update_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout))

	handler := NewHandler(mockService, logger)
	songID := uuid.New()
	reqBody := `{"name": "Updated Song", "group": "Updated Group"}`
	req := httptest.NewRequest(http.MethodPut, "/songs/"+songID.String(), strings.NewReader(reqBody))
	req = mux.SetURLVars(req, map[string]string{"id": songID.String()})
	w := httptest.NewRecorder()

	mockService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrSongNotFound)

	handler.Update(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Contains(t, string(body), "song not found")
}
