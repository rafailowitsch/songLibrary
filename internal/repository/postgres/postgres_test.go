package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"songLibrary/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Helper function to setup PostgreSQL container for songs
func setupPostgresForSongs(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dsn := "postgres://user:password@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	poolConfig, err := pgxpool.ParseConfig(dsn)
	assert.NoError(t, err)
	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	assert.NoError(t, err)

	_, err = conn.Exec(ctx, `
		CREATE TABLE songs (
			id UUID PRIMARY KEY,
			name VARCHAR(100),
			group_name VARCHAR(100),
			text TEXT,
			link TEXT,
			release_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	assert.NoError(t, err)

	teardown := func() {
		conn.Close()
		postgresContainer.Terminate(ctx)
	}

	return conn, teardown
}

func TestSongDB_Create(t *testing.T) {
	conn, teardown := setupPostgresForSongs(t)
	defer teardown()

	songDB := NewPostgres(conn)

	song := &domain.Song{
		Name:        "Hysteria",
		Group:       "Muse",
		Text:        "It's bugging me...",
		Link:        "https://link-to-song.com",
		ReleaseDate: time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC),
	}

	err := songDB.Create(context.Background(), song)
	assert.NoError(t, err)

	// Verify the song was inserted
	var insertedSong domain.Song
	err = conn.QueryRow(context.Background(), `SELECT id, name, group_name, text, link, release_date, created_at, updated_at FROM songs WHERE id = $1`, song.ID).Scan(
		&insertedSong.ID,
		&insertedSong.Name,
		&insertedSong.Group,
		&insertedSong.Text,
		&insertedSong.Link,
		&insertedSong.ReleaseDate,
		&insertedSong.CreatedAt,
		&insertedSong.UpdatedAt,
	)
	assert.NoError(t, err)
	assert.Equal(t, song.Name, insertedSong.Name)
	assert.Equal(t, song.Group, insertedSong.Group)
	assert.Equal(t, song.Text, insertedSong.Text)
	assert.Equal(t, song.Link, insertedSong.Link)
}

func TestSongDB_Read(t *testing.T) {
	conn, teardown := setupPostgresForSongs(t)
	defer teardown()

	// Insert a song for testing
	songID := uuid.New()
	_, err := conn.Exec(context.Background(), `INSERT INTO songs (id, name, group_name, text, link, release_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		songID, "Hysteria", "Muse", "It's bugging me...", "https://link-to-song.com", time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC), time.Now(), time.Now())
	assert.NoError(t, err)

	songDB := NewPostgres(conn)

	// Read the inserted song
	songSearch := &domain.SongInfo{ID: songID}
	song, err := songDB.Read(context.Background(), songSearch)
	assert.NoError(t, err)
	assert.NotNil(t, song)
	assert.Equal(t, songID, song.ID)
	assert.Equal(t, "Hysteria", song.Name)
	assert.Equal(t, "Muse", song.Group)
	assert.Equal(t, "It's bugging me...", song.Text)
}

func TestSongDB_ReadAllWithFilter(t *testing.T) {
	conn, teardown := setupPostgresForSongs(t)
	defer teardown()

	// Insert multiple songs for testing
	_, err := conn.Exec(context.Background(), `
		INSERT INTO songs (id, name, group_name, text, link, release_date, created_at, updated_at) VALUES
		($1, $2, $3, $4, $5, $6, $7, $8),
		($9, $10, $11, $12, $13, $14, $15, $16)`,
		uuid.New(), "Hysteria", "Muse", "It's bugging me...", "https://link-to-song1.com", time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC), time.Now(), time.Now(),
		uuid.New(), "Time is Running Out", "Muse", "I think I'm drowning...", "https://link-to-song2.com", time.Date(2003, 9, 15, 0, 0, 0, 0, time.UTC), time.Now(), time.Now(),
	)
	assert.NoError(t, err)

	songDB := NewPostgres(conn)

	song := &domain.Song{
		Group: "Muse",
	}
	songs, err := songDB.ReadAllWithFilter(context.Background(), song, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, songs, 2)

	song = &domain.Song{
		Name: "Time is Running Out",
	}
	songs, err = songDB.ReadAllWithFilter(context.Background(), song, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, songs, 1)

	songs, err = songDB.ReadAllWithFilter(context.Background(), &domain.Song{}, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, songs, 2)
}

func TestSongDB_Update(t *testing.T) {
	conn, teardown := setupPostgresForSongs(t)
	defer teardown()

	// Insert a song for testing
	songID := uuid.New()
	_, err := conn.Exec(context.Background(), `INSERT INTO songs (id, name, group_name, text, link, release_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		songID, "Hysteria", "Muse", "It's bugging me...", "https://link-to-song.com", time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC), time.Now(), time.Now())
	assert.NoError(t, err)

	songDB := NewPostgres(conn)

	// Initialize search and updatedSong
	songSearch := &domain.SongInfo{
		ID: songID,
	}

	updatedSong := &domain.Song{
		ID:          songID,
		Name:        "Hysteria (Updated)",
		Group:       "Muse",
		Text:        "It's bugging me... (Updated)",
		Link:        "https://link-to-song-updated.com",
		ReleaseDate: time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Now(),
	}

	err = songDB.Update(context.Background(), songSearch, updatedSong)
	assert.NoError(t, err)

	// Verify the song was updated
	var song domain.Song
	err = conn.QueryRow(context.Background(), `SELECT id, name, group_name, text, link, release_date, created_at, updated_at FROM songs WHERE id = $1`, songID).Scan(
		&song.ID,
		&song.Name,
		&song.Group,
		&song.Text,
		&song.Link,
		&song.ReleaseDate,
		&song.CreatedAt,
		&song.UpdatedAt,
	)
	assert.NoError(t, err)
	assert.Equal(t, updatedSong.Name, song.Name)
	assert.Equal(t, updatedSong.Text, song.Text)
	assert.Equal(t, updatedSong.Link, song.Link)
}

func TestSongDB_Delete(t *testing.T) {
	conn, teardown := setupPostgresForSongs(t)
	defer teardown()

	// Insert a song for testing
	songID := uuid.New()
	_, err := conn.Exec(context.Background(), `INSERT INTO songs (id, name, group_name, text, link, release_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		songID, "Hysteria", "Muse", "It's bugging me...", "https://link-to-song.com", time.Date(2003, 12, 1, 0, 0, 0, 0, time.UTC), time.Now(), time.Now())
	assert.NoError(t, err)

	songDB := NewPostgres(conn)

	// Create a search struct
	songSearch := &domain.SongInfo{ID: songID}

	// Delete the song
	err = songDB.Delete(context.Background(), songSearch)
	assert.NoError(t, err)

	// Verify the song was deleted
	var song domain.Song
	err = conn.QueryRow(context.Background(), `SELECT id, name, group_name, text, link, release_date, created_at, updated_at FROM songs WHERE id = $1`, songID).Scan(
		&song.ID,
		&song.Name,
		&song.Group,
		&song.Text,
		&song.Link,
		&song.ReleaseDate,
		&song.CreatedAt,
		&song.UpdatedAt,
	)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pgx.ErrNoRows))
}
