package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"songLibrary/internal/domain"
	"songLibrary/pkg/logger/logger/sl"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type SongRequest struct {
	Name  string `json:"name" validate:"required,name"`
	Group string `json:"group" validate:"required,group"`
}

type SongResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	ReleaseDate time.Time `json:"release_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaginatedTextResponse struct {
	Text []string `json:"text"`
}

type Service interface {
	Add(ctx context.Context, song *domain.SongInfo) error
	Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo) error
	Delete(ctx context.Context, song *domain.SongInfo) error

	GetAllWithFilter(ctx context.Context, song *domain.Song, page, pageSize int) ([]*domain.Song, error)
	GetPaginatedText(ctx context.Context, song *domain.SongInfo) ([]string, error)
}

type Handler struct {
	Service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{
		Service: service,
		log:     log,
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/songs", func(r chi.Router) {
		r.Post("/", h.Add)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Get("/", h.GetAllWithFilter)
		r.Get("/{id}/text", h.GetPaginatedText)
	})

	return r
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Add"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req SongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid request"))
		return
	}

	if req.Name == "" || req.Group == "" {
		log.Info("name or group is missing in request")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("name and group are required"))
		return
	}

	songInfo := &domain.SongInfo{
		Name:  req.Name,
		Group: req.Group,
	}

	if err := h.Service.Add(r.Context(), songInfo); err != nil {
		log.Error("failed to add song", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	log.Info("song successfully added", slog.String("song_name", songInfo.Name))
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, OkResp("song added successfully"))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Update"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error("invalid song id", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid song id"))
		return
	}

	var req SongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid request"))
		return
	}

	songInfo := &domain.SongInfo{
		ID:    id,
		Name:  req.Name,
		Group: req.Group,
	}

	if err := h.Service.Update(r.Context(), songInfo); err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Info("song not found during update", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrResp("song not found"))
			return
		}
		log.Error("failed to update song", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	log.Info("song successfully updated", slog.String("song_name", songInfo.Name))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, OkResp("song updated successfully"))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Delete"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error("invalid song id", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid song id"))
		return
	}

	songInfo := &domain.SongInfo{ID: id}

	if err := h.Service.Delete(r.Context(), songInfo); err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Info("song not found during deletion", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrResp("song not found"))
			return
		}
		log.Error("failed to delete song", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	log.Info("song successfully deleted", slog.String("song_id", id.String()))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, OkResp("song deleted successfully"))
}

func (h *Handler) GetAllWithFilter(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetAllWithFilter"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	group := r.URL.Query().Get("group")
	name := r.URL.Query().Get("name")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10
	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			log.Warn("invalid page parameter", slog.String("page", pageStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrResp("invalid page parameter"))
			return
		}
	}
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			log.Warn("invalid page_size parameter", slog.String("page_size", pageSizeStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrResp("invalid page_size parameter"))
			return
		}
	}

	songSearch := &domain.SongSearch{
		Name:  name,
		Group: group,
	}

	log.Info("attempting to fetch songs with filters", slog.String("group", group), slog.String("name", name), slog.Int("page", page), slog.Int("page_size", pageSize))

	songs, err := h.Service.GetAllWithFilter(r.Context(), songSearch, page, pageSize)
	if err != nil {
		log.Error("failed to fetch songs with filter", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	var songsResponse []SongResponse
	for _, song := range songs {
		songsResponse = append(songsResponse, *MustConvertSongToResponse(song))
	}

	log.Info("songs successfully fetched", slog.Int("count", len(songsResponse)))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, songsResponse)
}

func (h *Handler) GetPaginatedText(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetPaginatedText"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error("invalid song id", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid song id"))
		return
	}

	songInfo := &domain.SongInfo{ID: id}

	verses, err := h.Service.GetPaginatedText(r.Context(), songInfo)
	if err != nil {
		if errors.Is(err, domain.ErrSongNotFound) {
			log.Info("song not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrResp("song not found"))
			return
		}
		log.Error("failed to paginate song text", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	log.Info("song text successfully paginated", slog.String("song_id", id.String()))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, PaginatedTextResponse{Text: verses})
}

func ConvertSongToResponse(song *domain.Song) (*SongResponse, error) {
	if song.ID == uuid.Nil {
		return nil, domain.ErrInvalidSongID
	}

	if song.Name == "" {
		return nil, domain.ErrInvalidSongName
	}

	if song.Group == "" {
		return nil, domain.ErrInvalidSongGroup
	}

	if song.Text == "" {
		return nil, domain.ErrInvalidSongText
	}

	response := &SongResponse{
		ID:          song.ID,
		Name:        song.Name,
		Group:       song.Group,
		Text:        song.Text,
		Link:        song.Link,
		ReleaseDate: song.ReleaseDate,
		CreatedAt:   song.CreatedAt,
		UpdatedAt:   song.UpdatedAt,
	}

	return response, nil
}

func MustConvertSongToResponse(song *domain.Song) *SongResponse {
	songResponse, _ := ConvertSongToResponse(song)
	return songResponse
}

func ErrResp(err string) map[string]string {
	return map[string]string{"error": err}
}

func OkResp(msg string) map[string]string {
	return map[string]string{"message": msg}
}

// func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
// 	const op = "Handler.Get"

// 	log := h.log.With(
// 		slog.String("op", op),
// 		slog.String("request_id", middleware.GetReqID(r.Context())),
// 	)

// 	idParam := chi.URLParam(r, "id")
// 	id, err := uuid.Parse(idParam)
// 	if err != nil {
// 		log.Error("invalid song id", sl.Err(err))
// 		render.JSON(w, r, resp.Error("invalid song id"))
// 		return
// 	}

// 	songInfo := &domain.SongInfo{ID: id}
// 	song, err := h.Service.Get(r.Context(), songInfo)
// 	if errors.Is(err, domain.ErrSongNotFound) {
// 		log.Info("song not found", slog.String("id", id.String()))
// 		render.JSON(w, r, resp.Error("song not found"))
// 		return
// 	}
// 	if err != nil {
// 		log.Error("failed to get song", sl.Err(err))
// 		render.JSON(w, r, resp.Error("internal error"))
// 		return
// 	}

// 	convSong, err := ConvertSongToResponse(song)
// 	if err != nil {
// 		log.Error("failed to convert song into response", sl.Err(err))
// 		render.JSON(w, r, resp.Error("conversion error"))
// 	}

// 	log.Info("song successfully fetched", slog.String("song_name", song.Name))

// 	render.Status(r, http.StatusOK)
// 	render.JSON(w, r, convSong)
// }
