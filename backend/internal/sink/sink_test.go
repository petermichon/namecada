package sink

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/petermichon/domain-browser/backend/internal/engine"
	"github.com/petermichon/domain-browser/backend/internal/provider"
)

func openTest(t *testing.T) (*SQLite, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "checks.sqlite")
	s, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s, path
}

func result(id, platform, name, url string, status provider.Status, detail string) engine.Result {
	return engine.Result{ID: id, Platform: platform, Name: name, URL: url, Status: status, Detail: detail}
}

func count(t *testing.T, s *SQLite, query string) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func TestSQLiteWrite(t *testing.T) {
	s, _ := openTest(t)
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s.now = func() time.Time { return at }

	results := []engine.Result{
		result("lemmy", "lemmy", "acema", "https://lemmy.world/u/acema", provider.StatusTaken, "found"),
		result("peertube", "peertube", "acema", "https://framatube.org/a/acema", provider.StatusAvailable, ""),
	}
	if err := s.Write(context.Background(), results); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if got := count(t, s, "SELECT COUNT(*) FROM checks"); got != 2 {
		t.Fatalf("rows = %d, want 2", got)
	}

	var (
		platform, status, detail, url, checkedAt string
	)
	err := s.db.QueryRow(
		"SELECT platform, status, detail, url, checked_at FROM checks WHERE name = ? AND provider = ?",
		"acema", "lemmy",
	).Scan(&platform, &status, &detail, &url, &checkedAt)
	if err != nil {
		t.Fatalf("select lemmy row: %v", err)
	}
	if platform != "lemmy" || status != "taken" || detail != "found" || url != "https://lemmy.world/u/acema" {
		t.Errorf("row = %q/%q/%q/%q, unexpected", platform, status, detail, url)
	}
	if checkedAt != at.Format(time.RFC3339) {
		t.Errorf("checked_at = %q, want %q", checkedAt, at.Format(time.RFC3339))
	}
}

func TestSQLiteUpsertsOnRerun(t *testing.T) {
	s, _ := openTest(t)
	s.now = func() time.Time { return time.Unix(1_000, 0).UTC() }

	first := result("lemmy", "lemmy", "acema", "https://lemmy.world/u/acema", provider.StatusTaken, "found")
	if err := s.Write(context.Background(), []engine.Result{first}); err != nil {
		t.Fatalf("first Write: %v", err)
	}

	s.now = func() time.Time { return time.Unix(2_000, 0).UTC() }
	second := result("lemmy", "lemmy", "acema", "https://lemmy.world/u/acema", provider.StatusAvailable, "")
	if err := s.Write(context.Background(), []engine.Result{second}); err != nil {
		t.Fatalf("second Write: %v", err)
	}

	if got := count(t, s, "SELECT COUNT(*) FROM checks"); got != 1 {
		t.Fatalf("rows = %d, want 1 (upsert, not insert)", got)
	}

	var status, checkedAt string
	if err := s.db.QueryRow("SELECT status, checked_at FROM checks").Scan(&status, &checkedAt); err != nil {
		t.Fatal(err)
	}
	if status != "available" {
		t.Errorf("status = %q, want available", status)
	}
	if want := time.Unix(2_000, 0).UTC().Format(time.RFC3339); checkedAt != want {
		t.Errorf("checked_at = %q, want %q", checkedAt, want)
	}
}

func TestSQLiteKeyIsNameAndProvider(t *testing.T) {
	s, _ := openTest(t)
	s.now = func() time.Time { return time.Unix(0, 0).UTC() }

	// Same label, two providers, plus a second label for one provider.
	results := []engine.Result{
		result("lemmy", "lemmy", "acema", "u1", provider.StatusTaken, ""),
		result("peertube", "peertube", "acema", "u2", provider.StatusTaken, ""),
		result("lemmy", "lemmy", "other", "u3", provider.StatusTaken, ""),
	}
	if err := s.Write(context.Background(), results); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := count(t, s, "SELECT COUNT(*) FROM checks"); got != 3 {
		t.Fatalf("rows = %d, want 3", got)
	}
}

func TestSQLiteAvailableView(t *testing.T) {
	s, _ := openTest(t)
	s.now = func() time.Time { return time.Unix(0, 0).UTC() }

	results := []engine.Result{
		result("lemmy", "lemmy", "acema", "u1", provider.StatusTaken, ""),
		result("peertube", "peertube", "acema", "u2", provider.StatusAvailable, ""),
		result("github", "github", "acema", "u3", provider.StatusUnknown, ""),
		result("npm", "npm", "other", "u4", provider.StatusAvailable, ""),
	}
	if err := s.Write(context.Background(), results); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if got := count(t, s, "SELECT COUNT(*) FROM available"); got != 2 {
		t.Fatalf("available rows = %d, want 2", got)
	}

	rows, err := s.db.Query("SELECT name, provider FROM available ORDER BY name, provider")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name, p string
		if err := rows.Scan(&name, &p); err != nil {
			t.Fatal(err)
		}
		got = append(got, name+"/"+p)
	}
	if len(got) != 2 || got[0] != "acema/peertube" || got[1] != "other/npm" {
		t.Fatalf("available = %v, want [acema/peertube other/npm]", got)
	}
}

func TestSQLitePersistsAndReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checks.sqlite")

	s, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	s.now = func() time.Time { return time.Unix(0, 0).UTC() }
	if err := s.Write(context.Background(), []engine.Result{
		result("lemmy", "lemmy", "acema", "u1", provider.StatusTaken, ""),
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	if got := count(t, reopened, "SELECT COUNT(*) FROM checks"); got != 1 {
		t.Fatalf("rows after reopen = %d, want 1", got)
	}
}
