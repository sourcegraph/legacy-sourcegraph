CREATE TABLE IF NOT EXISTS embedding_plugin_files (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL,
    contents bytea NOT NULL,
    embedding_plugin_id INT NOT NULL
);
