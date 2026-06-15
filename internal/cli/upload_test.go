package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/machuz/eis/v2/internal/scorer"
)

// mapMemberSignals must carry scorer.Result onto the wire contract exactly —
// note the two renames the SaaS expects: Impact→total, Role→role_archetype.
func TestMapMemberSignals_FieldMapping(t *testing.T) {
	got := mapMemberSignals([]scorer.Result{{
		Author: "alice", Impact: 73.2, Catalysis: 68, Production: 100,
		Role: "Anchor", RoleConf: 0.7, TotalCommits: 12,
	}})
	if len(got) != 1 {
		t.Fatalf("want 1 signal, got %d", len(got))
	}
	s := got[0]
	if s.Author != "alice" || s.Total != 73.2 || s.Catalysis != 68 || s.Production != 100 {
		t.Fatalf("scalar mapping wrong: %+v", s)
	}
	if s.RoleArchetype != "Anchor" || s.RoleConfidence != 0.7 || s.TotalCommits != 12 {
		t.Fatalf("archetype mapping wrong: %+v", s)
	}
}

// The JSON body is the cross-repo contract with the SaaS UploadRequest. Assert
// the exact wire keys so a rename here fails loudly instead of silently
// dropping fields on the server.
func TestUploadRequest_WireContract(t *testing.T) {
	req := uploadRequest{
		GitSHA:        "abc123",
		SHAAuthorDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		AnalysisTime:  time.Date(2026, 6, 15, 1, 0, 0, 0, time.UTC),
		Domain:        "backend",
		Signals:       mapMemberSignals([]scorer.Result{{Author: "alice", Impact: 50}}),
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, key := range []string{
		`"git_sha":"abc123"`, `"sha_author_date":`, `"analysis_time":`,
		`"domain":"backend"`, `"signals":`, `"role_archetype":`, `"total":50`,
	} {
		if !strings.Contains(body, key) {
			t.Errorf("payload missing %s\nbody: %s", key, body)
		}
	}
}

func TestPostUpload_SendsBearerAndJSON(t *testing.T) {
	var gotAuth, gotCT, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	err := postUpload(context.Background(), srv.URL, "eis_tok", uploadRequest{GitSHA: "x"})
	if err != nil {
		t.Fatalf("postUpload: %v", err)
	}
	if gotAuth != "Bearer eis_tok" {
		t.Errorf("auth header = %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}
	if gotPath != "/api/v1/signals/upload" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPostUpload_SurfacesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("bad token"))
	}))
	defer srv.Close()

	err := postUpload(context.Background(), srv.URL, "eis_tok", uploadRequest{})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("want 401 error, got %v", err)
	}
}
