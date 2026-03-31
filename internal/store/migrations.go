// internal/store/migrations.go
package store

var migrations = []string{
	// Migration 1: Initial schema
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS connections (
		name        TEXT PRIMARY KEY,
		site        TEXT NOT NULL,
		is_default  INTEGER DEFAULT 0,
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS resources (
		resource_id         TEXT NOT NULL,
		resource_type       TEXT NOT NULL,
		connection          TEXT NOT NULL,
		title               TEXT,
		remote_modified_at  TIMESTAMP,
		remote_modified_by  TEXT,
		remote_etag         TEXT,
		last_synced_at      TIMESTAMP,
		status              TEXT DEFAULT 'active',
		PRIMARY KEY (resource_id, resource_type, connection),
		FOREIGN KEY (connection) REFERENCES connections(name) ON UPDATE CASCADE ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS resource_versions (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_id     TEXT NOT NULL,
		resource_type   TEXT NOT NULL,
		connection       TEXT NOT NULL,
		version         INTEGER NOT NULL,
		content         TEXT NOT NULL,
		remote_snapshot TEXT,
		applied_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		applied_by      TEXT,
		message         TEXT,
		UNIQUE(resource_id, resource_type, connection, version),
		FOREIGN KEY (resource_id, resource_type, connection)
			REFERENCES resources(resource_id, resource_type, connection)
			ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_versions_lookup
		ON resource_versions(resource_id, resource_type, connection, version DESC);
	CREATE INDEX IF NOT EXISTS idx_versions_applied_at
		ON resource_versions(applied_at);
	CREATE INDEX IF NOT EXISTS idx_resources_connection
		ON resources(connection);`,
}
