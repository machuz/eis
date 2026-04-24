package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/config"
	"github.com/machuz/eis/v2/internal/domain"
	"github.com/machuz/eis/v2/internal/git"
)

// --- resolveRepoDomain tests (config repo-pattern path, no git needed) ---

func TestResolveRepoDomain_LegacyConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"backend":  {Repos: []string{"api", "worker"}},
		"frontend": {Repos: []string{"web-app"}},
	}
	extMap := domain.BuildExtMap(nil, true)
	ctx := context.Background()

	tests := []struct {
		repoName string
		want     domain.Domain
	}{
		{"api", domain.Backend},
		{"worker", domain.Backend},
		{"web-app", domain.Frontend},
	}
	for _, tt := range tests {
		got := resolveRepoDomain(ctx, "/nonexistent", tt.repoName, cfg, extMap)
		if got != tt.want {
			t.Errorf("resolveRepoDomain(%q, legacy config) = %v, want %v", tt.repoName, got, tt.want)
		}
	}
}

func TestResolveRepoDomain_NewConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"mobile": {Repos: []string{"ios-app", "android-app"}, Extensions: []string{".swift", ".kt"}},
		"data":   {Repos: []string{"ml-pipeline"}, Extensions: []string{".ipynb"}},
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"mobile": {".swift", ".kt"},
		"data":   {".ipynb"},
	}, true)
	ctx := context.Background()

	tests := []struct {
		repoName string
		want     domain.Domain
	}{
		{"ios-app", domain.Domain("Mobile")},
		{"android-app", domain.Domain("Mobile")},
		{"ml-pipeline", domain.Domain("Data")},
	}
	for _, tt := range tests {
		got := resolveRepoDomain(ctx, "/nonexistent", tt.repoName, cfg, extMap)
		if got != tt.want {
			t.Errorf("resolveRepoDomain(%q, new config) = %v, want %v", tt.repoName, got, tt.want)
		}
	}
}

func TestResolveRepoDomain_MixedConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"backend": {Repos: []string{"api"}},                                     // legacy style
		"mobile":  {Repos: []string{"ios-app"}, Extensions: []string{".swift"}}, // new style
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"mobile": {".swift"},
	}, true)
	ctx := context.Background()

	got := resolveRepoDomain(ctx, "/nonexistent", "api", cfg, extMap)
	if got != domain.Backend {
		t.Errorf("legacy 'api' = %v, want Backend", got)
	}

	got = resolveRepoDomain(ctx, "/nonexistent", "ios-app", cfg, extMap)
	if got != domain.Domain("Mobile") {
		t.Errorf("new 'ios-app' = %v, want Mobile", got)
	}
}

func TestResolveRepoDomain_GlobPattern(t *testing.T) {
	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"backend": {Repos: []string{"*-service", "api-*"}},
	}
	extMap := domain.BuildExtMap(nil, true)
	ctx := context.Background()

	tests := []struct {
		repoName string
		want     domain.Domain
	}{
		{"auth-service", domain.Backend},
		{"api-gateway", domain.Backend},
	}
	for _, tt := range tests {
		got := resolveRepoDomain(ctx, "/nonexistent", tt.repoName, cfg, extMap)
		if got != tt.want {
			t.Errorf("resolveRepoDomain(%q, glob) = %v, want %v", tt.repoName, got, tt.want)
		}
	}
}

func TestResolveRepoDomain_BuiltInWithExtensions(t *testing.T) {
	// backend: { extensions: [.ts], repos: [api] }
	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"backend": {Repos: []string{"api"}, Extensions: []string{".ts"}},
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"backend": {".ts"},
	}, true)
	ctx := context.Background()

	// Repo pattern match
	got := resolveRepoDomain(ctx, "/nonexistent", "api", cfg, extMap)
	if got != domain.Backend {
		t.Errorf("repo pattern 'api' = %v, want Backend", got)
	}
}

// --- resolveRepoDomain with auto-detection (needs a temp git repo) ---

func TestResolveRepoDomain_AutoDetect_DefaultExtMap(t *testing.T) {
	repoPath := createTempGitRepo(t, map[string]string{
		"main.go":    "package main",
		"handler.go": "package main",
	})
	cfg := config.Default()
	extMap := domain.BuildExtMap(nil, true)

	got := resolveRepoDomain(context.Background(), repoPath, filepath.Base(repoPath), cfg, extMap)
	if got != domain.Backend {
		t.Errorf("auto-detect go files = %v, want Backend", got)
	}
}

func TestResolveRepoDomain_AutoDetect_CustomExtMap(t *testing.T) {
	repoPath := createTempGitRepo(t, map[string]string{
		"App.swift":   "import UIKit",
		"Model.swift": "struct Model {}",
	})
	cfg := config.Default()
	extMap := domain.BuildExtMap(map[string][]string{
		"mobile": {".swift"},
	}, true)

	got := resolveRepoDomain(context.Background(), repoPath, filepath.Base(repoPath), cfg, extMap)
	if got != domain.Domain("Mobile") {
		t.Errorf("auto-detect swift files with custom extMap = %v, want Mobile", got)
	}
}

func TestResolveRepoDomain_AutoDetect_OverrideBuiltIn(t *testing.T) {
	// .ts files default to Frontend, override to Backend
	repoPath := createTempGitRepo(t, map[string]string{
		"server.ts":  "console.log('hi')",
		"handler.ts": "export default {}",
	})
	cfg := config.Default()
	extMap := domain.BuildExtMap(map[string][]string{
		"backend": {".ts"},
	}, true)

	got := resolveRepoDomain(context.Background(), repoPath, filepath.Base(repoPath), cfg, extMap)
	if got != domain.Backend {
		t.Errorf("auto-detect .ts overridden to Backend = %v, want Backend", got)
	}
}

// --- default_domains: false ---

func TestResolveRepoDomain_NoDefaults_RepoPattern(t *testing.T) {
	falseVal := false
	cfg := config.Default()
	cfg.DefaultDomains = &falseVal
	cfg.Domains = config.DomainsConfig{
		"server": {Repos: []string{"api"}, Extensions: []string{".go", ".py"}},
		"client": {Repos: []string{"web"}, Extensions: []string{".ts", ".tsx"}},
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"server": {".go", ".py"},
		"client": {".ts", ".tsx"},
	}, false)
	ctx := context.Background()

	got := resolveRepoDomain(ctx, "/nonexistent", "api", cfg, extMap)
	if got != domain.Domain("Server") {
		t.Errorf("no-defaults repo pattern 'api' = %v, want Server", got)
	}

	got = resolveRepoDomain(ctx, "/nonexistent", "web", cfg, extMap)
	if got != domain.Domain("Client") {
		t.Errorf("no-defaults repo pattern 'web' = %v, want Client", got)
	}
}

func TestResolveRepoDomain_NoDefaults_AutoDetect(t *testing.T) {
	repoPath := createTempGitRepo(t, map[string]string{
		"main.go":    "package main",
		"handler.go": "package main",
	})
	falseVal := false
	cfg := config.Default()
	cfg.DefaultDomains = &falseVal
	// .go → Server (not Backend since defaults are off)
	extMap := domain.BuildExtMap(map[string][]string{
		"server": {".go"},
	}, false)

	got := resolveRepoDomain(context.Background(), repoPath, filepath.Base(repoPath), cfg, extMap)
	if got != domain.Domain("Server") {
		t.Errorf("no-defaults auto-detect .go = %v, want Server", got)
	}
}

func TestResolveRepoDomain_NoDefaults_UnrecognizedExt(t *testing.T) {
	repoPath := createTempGitRepo(t, map[string]string{
		"Main.java": "class Main {}",
	})
	falseVal := false
	cfg := config.Default()
	cfg.DefaultDomains = &falseVal
	// Only .go is mapped; .java is unrecognized
	extMap := domain.BuildExtMap(map[string][]string{
		"server": {".go"},
	}, false)

	got := resolveRepoDomain(context.Background(), repoPath, filepath.Base(repoPath), cfg, extMap)
	if got != domain.Unknown {
		t.Errorf("no-defaults unrecognized .java = %v, want Unknown", got)
	}
}

// --- --domain filter compatibility ---

func TestDomainFilter_BuiltIn(t *testing.T) {
	tests := []struct {
		filter     string
		repoDomain domain.Domain
		shouldSkip bool
	}{
		{"Backend", domain.Backend, false}, // alias → BE
		{"backend", domain.Backend, false}, // case-insensitive alias
		{"BACKEND", domain.Backend, false}, // all caps alias
		{"BE", domain.Backend, false},      // short form
		{"be", domain.Backend, false},      // short form lowercase
		{"Frontend", domain.Backend, true}, // different domain
		{"", domain.Backend, false},        // no filter = include all
	}
	for _, tt := range tests {
		skipped := tt.filter != "" && domain.NormalizeName(tt.filter) != tt.repoDomain
		if skipped != tt.shouldSkip {
			t.Errorf("filter=%q domain=%v: skipped=%v, want %v", tt.filter, tt.repoDomain, skipped, tt.shouldSkip)
		}
	}
}

func TestDomainFilter_CustomDomain(t *testing.T) {
	tests := []struct {
		filter     string
		repoDomain domain.Domain
		shouldSkip bool
	}{
		{"Mobile", domain.Domain("Mobile"), false},
		{"mobile", domain.Domain("Mobile"), false}, // case-insensitive
		{"Backend", domain.Domain("Mobile"), true},  // different domain
		{"Data", domain.Domain("Data"), false},
	}
	for _, tt := range tests {
		skipped := tt.filter != "" && domain.NormalizeName(tt.filter) != tt.repoDomain
		if skipped != tt.shouldSkip {
			t.Errorf("filter=%q domain=%v: skipped=%v, want %v", tt.filter, tt.repoDomain, skipped, tt.shouldSkip)
		}
	}
}

// --- timeline domain grouping (shared resolveRepoDomain + SortDomains) ---

func TestTimeline_CustomDomainGrouping(t *testing.T) {
	// Simulate what timeline.go does: group repos by domain, then iterate in SortDomains order
	goRepo := createTempGitRepo(t, map[string]string{"main.go": "package main"})
	swiftRepo := createTempGitRepo(t, map[string]string{"App.swift": "import UIKit"})
	tsRepo := createTempGitRepo(t, map[string]string{"app.ts": "console.log()"})

	cfg := config.Default()
	cfg.Domains = config.DomainsConfig{
		"mobile": {Extensions: []string{".swift"}},
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"mobile": {".swift"},
	}, true)
	ctx := context.Background()

	type repoInfo struct {
		path   string
		domain domain.Domain
	}
	var repos []repoInfo
	for _, rp := range []string{goRepo, swiftRepo, tsRepo} {
		d := resolveRepoDomain(ctx, rp, filepath.Base(rp), cfg, extMap)
		repos = append(repos, repoInfo{path: rp, domain: d})
	}

	// Group by domain
	domainRepos := make(map[domain.Domain][]repoInfo)
	for _, r := range repos {
		domainRepos[r.domain] = append(domainRepos[r.domain], r)
	}

	// Should have Backend (.go), Frontend (.ts), and Mobile (.swift)
	if _, ok := domainRepos[domain.Backend]; !ok {
		t.Error("expected Backend domain for .go repo")
	}
	if _, ok := domainRepos[domain.Frontend]; !ok {
		t.Error("expected Frontend domain for .ts repo")
	}
	if _, ok := domainRepos[domain.Domain("Mobile")]; !ok {
		t.Error("expected Mobile domain for .swift repo")
	}

	// SortDomains should order: Backend, Frontend, Mobile
	var keys []domain.Domain
	for d := range domainRepos {
		keys = append(keys, d)
	}
	sorted := domain.SortDomains(keys)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 domains, got %d", len(sorted))
	}
	if sorted[0] != domain.Backend {
		t.Errorf("sorted[0] = %v, want Backend", sorted[0])
	}
	if sorted[1] != domain.Frontend {
		t.Errorf("sorted[1] = %v, want Frontend", sorted[1])
	}
	if sorted[2] != domain.Domain("Mobile") {
		t.Errorf("sorted[2] = %v, want Mobile", sorted[2])
	}
}

func TestTimeline_NoDefaultDomains_Grouping(t *testing.T) {
	goRepo := createTempGitRepo(t, map[string]string{"main.go": "package main"})
	tsRepo := createTempGitRepo(t, map[string]string{"app.ts": "console.log()"})

	falseVal := false
	cfg := config.Default()
	cfg.DefaultDomains = &falseVal
	cfg.Domains = config.DomainsConfig{
		"server": {Extensions: []string{".go"}},
		"client": {Extensions: []string{".ts"}},
	}
	extMap := domain.BuildExtMap(map[string][]string{
		"server": {".go"},
		"client": {".ts"},
	}, false)
	ctx := context.Background()

	domainRepos := make(map[domain.Domain]int)
	for _, rp := range []string{goRepo, tsRepo} {
		d := resolveRepoDomain(ctx, rp, filepath.Base(rp), cfg, extMap)
		domainRepos[d]++
	}

	// With defaults off, .go → Server, .ts → Client (not Backend/Frontend)
	if _, ok := domainRepos[domain.Domain("Server")]; !ok {
		t.Error("expected Server domain for .go repo (defaults off)")
	}
	if _, ok := domainRepos[domain.Domain("Client")]; !ok {
		t.Error("expected Client domain for .ts repo (defaults off)")
	}
	if _, ok := domainRepos[domain.Backend]; ok {
		t.Error("Backend should not exist when defaults are off")
	}
	if _, ok := domainRepos[domain.Frontend]; ok {
		t.Error("Frontend should not exist when defaults are off")
	}
}

// --- revert filtering tests ---

func TestFilterRevertedCommits(t *testing.T) {
	commits := []git.Commit{
		{Hash: "aaa", Author: "alice", Subject: "feat: add feature"},
		{Hash: "bbb", Author: "bob", Subject: "fix: bug"},
		{Hash: "ccc", Author: "alice", Subject: "Revert \"feat: add feature\""},
	}

	reverted := map[string]bool{"aaa": true, "ccc": true}
	filtered := filterRevertedCommits(commits, reverted)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 commit after filtering, got %d", len(filtered))
	}
	if filtered[0].Hash != "bbb" {
		t.Errorf("expected remaining commit bbb, got %s", filtered[0].Hash)
	}
}

func TestFilterRevertedCommits_NilMap(t *testing.T) {
	commits := []git.Commit{
		{Hash: "aaa", Author: "alice"},
	}
	filtered := filterRevertedCommits(commits, nil)
	if len(filtered) != 1 {
		t.Fatalf("nil map should return all commits, got %d", len(filtered))
	}
}

func TestFindRevertedCommits_NonMergeRevert(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	// Create a repo with a commit, then revert it
	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "test")

	os.WriteFile(filepath.Join(dir, "file.go"), []byte("package main\nfunc hello() {}"), 0644)
	run("git", "add", "-A")
	run("git", "commit", "-m", "feat: add hello")
	originalHash := run("git", "rev-parse", "HEAD")

	// Revert the commit
	run("git", "revert", "HEAD", "--no-edit")

	ctx := context.Background()
	excluded, err := git.FindRevertedCommits(ctx, dir)
	if err != nil {
		t.Fatalf("FindRevertedCommits failed: %v", err)
	}
	if len(excluded) == 0 {
		t.Fatal("expected excluded commits, got none")
	}
	if !excluded[originalHash] {
		t.Errorf("original commit %s should be excluded", originalHash)
	}
}

func TestFindRevertedCommits_RevertOfRevert(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "test")

	os.WriteFile(filepath.Join(dir, "file.go"), []byte("package main"), 0644)
	run("git", "add", "-A")
	run("git", "commit", "-m", "feat: initial")
	originalHash := run("git", "rev-parse", "HEAD")

	// Revert the commit
	run("git", "revert", "HEAD", "--no-edit")

	// Revert the revert (reinstate)
	run("git", "revert", "HEAD", "--no-edit")

	ctx := context.Background()
	excluded, err := git.FindRevertedCommits(ctx, dir)
	if err != nil {
		t.Fatalf("FindRevertedCommits failed: %v", err)
	}

	// Original should NOT be excluded (revert-of-revert cancels out)
	if excluded[originalHash] {
		t.Errorf("original commit %s should NOT be excluded after revert-of-revert", originalHash)
	}
}

// --- helpers ---

// createTempGitRepo creates a temp directory with git init and the given files committed.
func createTempGitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "test")

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	run("git", "add", "-A")
	run("git", "commit", "-m", "init")

	return dir
}
