CREATE INDEX IF NOT EXISTS idx_songs_name ON songs (name);
CREATE INDEX IF NOT EXISTS idx_songs_group_name ON songs (group_name);
CREATE INDEX IF NOT EXISTS idx_songs_release_date ON songs (release_date);