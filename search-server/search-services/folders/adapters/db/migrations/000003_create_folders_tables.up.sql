CREATE SCHEMA IF NOT EXISTS folder;

CREATE TABLE IF NOT EXISTS folder.folders (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS folder.folder_comics (
    id BIGSERIAL PRIMARY KEY,
    comic_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    folder_id BIGINT NOT NULL REFERENCES folder.folders(id) ON DELETE CASCADE,
    UNIQUE (folder_id, comic_id)
);

CREATE INDEX IF NOT EXISTS idx_folders_user_id ON folder.folders(user_id);
CREATE INDEX IF NOT EXISTS idx_folder_comics_folder_id ON folder.folder_comics(folder_id);
