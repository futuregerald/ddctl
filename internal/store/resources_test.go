// internal/store/resources_test.go
package store

import (
	"testing"
)

func seedConnection(t *testing.T, s *Store) {
	t.Helper()
	if err := s.AddConnection("prod", "datadoghq.com"); err != nil {
		t.Fatal(err)
	}
}

func TestTrackResource(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)

	err := s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")
	if err != nil {
		t.Fatal(err)
	}

	resources, err := s.ListResources("dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].ResourceID != "abc-123" {
		t.Errorf("expected id 'abc-123', got %q", resources[0].ResourceID)
	}
}

func TestSaveVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	err := s.SaveVersion("abc-123", "dashboard", "prod", "content here", "", "gerald", "initial version")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != 1 {
		t.Errorf("expected version 1, got %d", versions[0].Version)
	}
	if versions[0].Message != "initial version" {
		t.Errorf("expected message 'initial version', got %q", versions[0].Message)
	}
}

func TestSaveVersion_AutoIncrements(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v1", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v2", "", "gerald", "second")

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	// Versions returned newest first
	if versions[0].Version != 2 {
		t.Errorf("expected latest version 2, got %d", versions[0].Version)
	}
}

func TestGetLatestVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "old content", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "new content", "", "gerald", "second")

	v, err := s.GetLatestVersion("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if v.Content != "new content" {
		t.Errorf("expected 'new content', got %q", v.Content)
	}
}

func TestGetVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v1 content", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v2 content", "", "gerald", "second")

	v, err := s.GetVersion("abc-123", "dashboard", "prod", 1)
	if err != nil {
		t.Fatal(err)
	}
	if v.Content != "v1 content" {
		t.Errorf("expected 'v1 content', got %q", v.Content)
	}
}

func TestPruneVersions(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	for i := 0; i < 5; i++ {
		_ = s.SaveVersion("abc-123", "dashboard", "prod", "content", "", "gerald", "")
	}

	pruned, err := s.PruneVersions("abc-123", "dashboard", "prod", 3)
	if err != nil {
		t.Fatal(err)
	}
	if pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", pruned)
	}

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Errorf("expected 3 remaining, got %d", len(versions))
	}
}
