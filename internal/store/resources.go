// internal/store/resources.go
package store

import (
	"database/sql"
	"fmt"
	"os/user"
	"time"
)

type Resource struct {
	ResourceID       string
	ResourceType     string
	Connection       string
	Title            string
	RemoteModifiedAt *time.Time
	RemoteModifiedBy string
	RemoteEtag       string
	LastSyncedAt     *time.Time
	Status           string
}

type ResourceVersion struct {
	ID             int
	ResourceID     string
	ResourceType   string
	Connection     string
	Version        int
	Content        string
	RemoteSnapshot string
	AppliedAt      time.Time
	AppliedBy      string
	Message        string
}

func (s *Store) TrackResource(resourceID, resourceType, connection, title string) error {
	_, err := s.db.Exec(
		`INSERT INTO resources (resource_id, resource_type, connection, title)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(resource_id, resource_type, connection)
		 DO UPDATE SET title = excluded.title`,
		resourceID, resourceType, connection, title,
	)
	return err
}

func (s *Store) ListResources(resourceType, connection string) ([]Resource, error) {
	rows, err := s.db.Query(
		`SELECT resource_id, resource_type, connection, title, remote_modified_at,
		        remote_modified_by, last_synced_at, status
		 FROM resources WHERE resource_type = ? AND connection = ?
		 ORDER BY title`,
		resourceType, connection,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var r Resource
		var remoteModifiedBy sql.NullString
		if err := rows.Scan(&r.ResourceID, &r.ResourceType, &r.Connection,
			&r.Title, &r.RemoteModifiedAt, &remoteModifiedBy,
			&r.LastSyncedAt, &r.Status); err != nil {
			return nil, err
		}
		r.RemoteModifiedBy = remoteModifiedBy.String
		resources = append(resources, r)
	}
	return resources, rows.Err()
}

func (s *Store) SaveVersion(resourceID, resourceType, connection, content, remoteSnapshot, appliedBy, message string) error {
	if appliedBy == "" {
		if u, err := user.Current(); err == nil {
			appliedBy = u.Username
		}
	}

	// Get next version number
	var maxVersion int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?`,
		resourceID, resourceType, connection,
	).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO resource_versions (resource_id, resource_type, connection, version, content, remote_snapshot, applied_by, message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		resourceID, resourceType, connection, maxVersion+1, content, remoteSnapshot, appliedBy, message,
	)
	return err
}

func (s *Store) ListVersions(resourceID, resourceType, connection string) ([]ResourceVersion, error) {
	rows, err := s.db.Query(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		 ORDER BY version DESC`,
		resourceID, resourceType, connection,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []ResourceVersion
	for rows.Next() {
		var v ResourceVersion
		var remoteSnapshot sql.NullString
		var appliedBy sql.NullString
		var message sql.NullString
		if err := rows.Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
			&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &message); err != nil {
			return nil, err
		}
		v.RemoteSnapshot = remoteSnapshot.String
		v.AppliedBy = appliedBy.String
		v.Message = message.String
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *Store) GetLatestVersion(resourceID, resourceType, connection string) (*ResourceVersion, error) {
	v := &ResourceVersion{}
	var remoteSnapshot, appliedBy, message sql.NullString
	err := s.db.QueryRow(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		 ORDER BY version DESC LIMIT 1`,
		resourceID, resourceType, connection,
	).Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
		&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &message)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no versions found for %s/%s", resourceType, resourceID)
	}
	if err != nil {
		return nil, err
	}
	v.RemoteSnapshot = remoteSnapshot.String
	v.AppliedBy = appliedBy.String
	v.Message = message.String
	return v, nil
}

func (s *Store) GetVersion(resourceID, resourceType, connection string, version int) (*ResourceVersion, error) {
	v := &ResourceVersion{}
	var remoteSnapshot, appliedBy, msg sql.NullString
	err := s.db.QueryRow(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ? AND version = ?`,
		resourceID, resourceType, connection, version,
	).Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
		&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &msg)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version %d not found for %s/%s", version, resourceType, resourceID)
	}
	if err != nil {
		return nil, err
	}
	v.RemoteSnapshot = remoteSnapshot.String
	v.AppliedBy = appliedBy.String
	v.Message = msg.String
	return v, nil
}

func (s *Store) PruneVersions(resourceID, resourceType, connection string, keep int) (int, error) {
	result, err := s.db.Exec(
		`DELETE FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		   AND version NOT IN (
		     SELECT version FROM resource_versions
		     WHERE resource_id = ? AND resource_type = ? AND connection = ?
		     ORDER BY version DESC LIMIT ?
		   )`,
		resourceID, resourceType, connection,
		resourceID, resourceType, connection, keep,
	)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (s *Store) UpdateResourceSync(resourceID, resourceType, connection string, modifiedAt *time.Time, modifiedBy, etag string) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE resources SET remote_modified_at = ?, remote_modified_by = ?, remote_etag = ?, last_synced_at = ?
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?`,
		modifiedAt, modifiedBy, etag, now, resourceID, resourceType, connection,
	)
	return err
}
