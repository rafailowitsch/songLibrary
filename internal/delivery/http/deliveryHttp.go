package deliveryHttp

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	mwLogger "songLibrary/internal/delivery/http/middleware/logger"
	"songLibrary/internal/domain"
	"songLibrary/internal/dto"
	"songLibrary/pkg/logger/sl"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Service interface {
	Add(ctx context.Context, song *domain.SongInfo) error
	Get(ctx context.Context, song *domain.SongInfo) (*domain.Song, error)
	Update(ctx context.Context, song *domain.SongInfo, updatedSong *domain.Song) error
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

	// r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(mwLogger.New(h.log))
	r.Use(middleware.Recoverer)

	r.Route("/songs", func(r chi.Router) {
		r.Post("/", h.Add)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Get("/", h.GetAllWithFilter)
		r.Get("/{id}/text", h.GetPaginatedText)
	})

	r.Get("/ping", h.Ping)

	return r
}

// @Summary Add a new song
// @Description Add a new song to the library
// @Tags songs
// @Accept  json
// @Produce  json
// @Param song body dto.AddSongRequest true "Add song request"
// @Success 201 {object} map[string]string "song added successfully"
// @Failure 400 {object} map[string]string "invalid request"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs [post]
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Add"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req dto.AddSongRequest
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

// @Summary Get a song
// @Description Get song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id path string true "Song ID"
// @Success 200 {object} dto.SongResponse
// @Failure 400 {object} map[string]string "invalid song id"
// @Failure 404 {object} map[string]string "song not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Get"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error("invalid song id", sl.Err(err))
		render.JSON(w, r, ErrResp("invalid song id"))
		return
	}

	songInfo := &domain.SongInfo{ID: id}
	song, err := h.Service.Get(r.Context(), songInfo)
	if errors.Is(err, domain.ErrSongNotFound) {
		log.Info("song not found", slog.String("id", id.String()))
		render.JSON(w, r, ErrResp("song not found"))
		return
	}
	if err != nil {
		log.Error("failed to get song", sl.Err(err))
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	convSong, err := ConvertSongToResponse(song)
	if err != nil {
		log.Error("failed to convert song into response", sl.Err(err))
		render.JSON(w, r, ErrResp("conversion error"))
	}

	log.Info("song successfully fetched", slog.String("song_name", song.Name))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, convSong)
}

// @Summary Update a song
// @Description Update a song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id path string true "Song ID"
// @Param song body dto.UpdateSongRequest true "Update song request"
// @Success 200 {object} map[string]string "song updated successfully"
// @Failure 400 {object} map[string]string "invalid request or invalid song id"
// @Failure 404 {object} map[string]string "song not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs/{id} [put]
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

	var req dto.UpdateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrResp("invalid request"))
		return
	}

	songInfo := &domain.SongInfo{ID: id}
	song := &domain.Song{
		Name:  req.Name,
		Group: req.Group,
		Text:  req.Text,
		Link:  req.Link,
	}

	if err := h.Service.Update(r.Context(), songInfo, song); err != nil {
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

// @Summary Delete a song
// @Description Delete a song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id path string true "Song ID"
// @Success 200 {object} map[string]string "song deleted successfully"
// @Failure 400 {object} map[string]string "invalid song id"
// @Failure 404 {object} map[string]string "song not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs/{id} [delete]
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

// @Summary Get all songs with filters
// @Description Get a list of songs with optional filters for group, name, and release date, with pagination
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string false "Filter by group"
// @Param song query string false "Filter by song name"
// @Param release_date query string false "Filter by release date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param page_size query int false "Number of songs per page"
// @Success 200 {array} dto.SongResponse
// @Failure 400 {object} map[string]string "invalid page or page_size parameter"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs [get]
func (h *Handler) GetAllWithFilter(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetAllWithFilter"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	group := r.URL.Query().Get("group")
	name := r.URL.Query().Get("song")
	releaseDateStr := r.URL.Query().Get("release_date")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 0
	pageSize := 0
	var err error

	// Обработка параметра page
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			log.Warn("invalid page parameter", slog.String("page", pageStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrResp("invalid page parameter"))
			return
		}
	}

	// Обработка параметра page_size
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			log.Warn("invalid page_size parameter", slog.String("page_size", pageSizeStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrResp("invalid page_size parameter"))
			return
		}
	}

	// Обработка параметра release_date (дата релиза)
	var releaseDate time.Time
	if releaseDateStr != "" {
		releaseDate, err = time.Parse("2006-01-02", releaseDateStr) // Используем формат YYYY-MM-DD
		if err != nil {
			log.Warn("invalid release_date parameter", slog.String("release_date", releaseDateStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrResp("invalid release_date parameter"))
			return
		}
	}

	songSearch := &domain.Song{
		Name:        name,
		Group:       group,
		ReleaseDate: releaseDate, // Передаем дату релиза в объект поиска
	}

	log.Info("attempting to fetch songs with filters",
		slog.String("group", group),
		slog.String("name", name),
		slog.String("release_date", releaseDateStr),
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
	)

	songs, err := h.Service.GetAllWithFilter(r.Context(), songSearch, page, pageSize)
	if err != nil {
		log.Error("failed to fetch songs with filter", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrResp("internal error"))
		return
	}

	var songsResponse []dto.SongResponse
	for _, song := range songs {
		songsResponse = append(songsResponse, *MustConvertSongToResponse(song))
	}

	log.Info("songs successfully fetched", slog.Int("count", len(songsResponse)))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, songsResponse)
}

// @Summary Get paginated text of a song
// @Description Get paginated text of the song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param id path string true "Song ID"
// @Success 200 {object} dto.PaginatedTextResponse
// @Failure 400 {object} map[string]string "invalid song id"
// @Failure 404 {object} map[string]string "song not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /songs/{id}/text [get]
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
	render.JSON(w, r, dto.PaginatedTextResponse{Text: verses})
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Ping"

	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	log.Info("ping sent")
	render.Status(r, http.StatusOK)
	render.JSON(w, r, "pong")
}

func ConvertSongToResponse(song *domain.Song) (*dto.SongResponse, error) {
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

	response := &dto.SongResponse{
		ID:          song.ID.String(),
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

func MustConvertSongToResponse(song *domain.Song) *dto.SongResponse {
	songResponse, _ := ConvertSongToResponse(song)
	return songResponse
}

func ErrResp(err string) map[string]string {
	return map[string]string{"error": err}
}

func OkResp(msg string) map[string]string {
	return map[string]string{"message": msg}
}
