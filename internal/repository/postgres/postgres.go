package postgres

import (
	"context"
	"errors"
	"fmt"
	"songLibrary/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type SongDB struct {
	db *pgx.Conn
}

func NewSongDB(db *pgx.Conn) *SongDB {
	return &SongDB{
		db: db,
	}
}

func (s *SongDB) Create(ctx context.Context, song *domain.Song) error {
	const op = "repository.SongDB.Create"

	song.ID = uuid.New()
	song.CreatedAt = time.Now()
	song.UpdatedAt = time.Now()

	query := `INSERT INTO songs (id, name, group_name, text, link, release_date, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(ctx, query, song.ID, song.Name, song.Group, song.Text, song.Link, song.ReleaseDate, song.CreatedAt, song.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Код ошибки для дубликатов
				return fmt.Errorf("%s: %w", op, domain.ErrSongExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *SongDB) Read(ctx context.Context, song *domain.SongSearch) (*domain.Song, error) {
	const op = "repository.SongDB.Read"

	query := `SELECT id, name, group_name, text, link, release_date, created_at, updated_at
              FROM songs WHERE id = $1`
	row := s.db.QueryRow(ctx, query, song.ID)

	var targetSong domain.Song
	err := row.Scan(&targetSong.ID, &targetSong.Name, &targetSong.Group, &targetSong.Text, &targetSong.Link, &targetSong.ReleaseDate, &targetSong.CreatedAt, &targetSong.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrSongNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &targetSong, nil
}

func (s *SongDB) ReadAllWithFilter(ctx context.Context, song *domain.Song, limit, offset int) ([]*domain.Song, error) {
	const op = "repository.SongDB.ReadAllWithFilter"

	// Базовый запрос
	query := `SELECT id, name, group_name, text, link, release_date, created_at, updated_at FROM songs`
	var conditions []string
	var params []interface{}
	var paramIndex = 1

	// Проверяем поля фильтра и добавляем условия в запрос
	if song.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", paramIndex))
		params = append(params, "%"+song.Name+"%")
		paramIndex++
	}
	if song.Group != "" {
		conditions = append(conditions, fmt.Sprintf("group_name ILIKE $%d", paramIndex))
		params = append(params, "%"+song.Group+"%")
		paramIndex++
	}
	if !song.ReleaseDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("release_date = $%d", paramIndex))
		params = append(params, song.ReleaseDate)
		paramIndex++
	}

	// Добавляем условия к запросу, если они есть
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if limit != 0 {
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
		params = append(params, limit, offset)
	}

	// Выполняем запрос
	rows, err := s.db.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	// Обрабатываем результаты
	var songs []*domain.Song
	for rows.Next() {
		var song domain.Song
		err := rows.Scan(&song.ID, &song.Name, &song.Group, &song.Text, &song.Link, &song.ReleaseDate, &song.CreatedAt, &song.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		songs = append(songs, &song)
	}

	return songs, nil
}

func (s *SongDB) Update(ctx context.Context, songSearch *domain.SongSearch, updatedSong *domain.Song) error {
	const op = "repository.SongDB.Update"

	updatedSong.UpdatedAt = time.Now()

	query := `UPDATE songs SET name = $1, group_name = $2, text = $3, link = $4, release_date = $5, updated_at = $6 
              WHERE id = $7`

	result, err := s.db.Exec(ctx, query, updatedSong.Name, updatedSong.Group, updatedSong.Text, updatedSong.Link, updatedSong.ReleaseDate, updatedSong.UpdatedAt, songSearch.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrSongNotFound)
	}

	return nil
}

func (s *SongDB) Delete(ctx context.Context, songSearch *domain.SongSearch) error {
	const op = "repository.SongDB.Delete"

	query := `DELETE FROM songs WHERE id = $1`
	result, err := s.db.Exec(ctx, query, songSearch.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrSongNotFound)
	}

	return nil
}
