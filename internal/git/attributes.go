package git

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// LinguistExcluded returns the tracked paths marked linguist-generated or
// linguist-vendored in .gitattributes — the generated / vendored files (lockfiles,
// *.pb.go, minified bundles, vendored deps) that must not inflate anyone's gravity.
//
// One batched `git check-attr` over the whole tracked file list (git ls-files),
// never a per-file call, so it stays cheap on large repos. Best-effort: any git
// error yields an empty set + the error, so an attribute-setup edge never drops
// real files. Deterministic — the attributes come from the repo's committed
// .gitattributes at the checked-out tree.
func LinguistExcluded(ctx context.Context, repoPath string) ([]string, error) {
	files, err := ListAllFiles(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	// `git check-attr -z --stdin <attrs>` reads NUL-separated paths on stdin and
	// writes NUL-separated <path>\0<attr>\0<info>\0 triples — robust to spaces or
	// newlines in paths. Each path yields one triple per requested attribute.
	var in bytes.Buffer
	for _, f := range files {
		in.WriteString(f)
		in.WriteByte(0)
	}
	cmd := exec.CommandContext(ctx, "git", "check-attr", "-z", "--stdin", "linguist-generated", "linguist-vendored")
	cmd.Dir = repoPath
	cmd.Stdin = &in
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fields := strings.Split(string(out), "\x00")
	seen := map[string]bool{}
	var excluded []string
	for i := 0; i+2 < len(fields); i += 3 {
		path, info := fields[i], fields[i+2]
		// linguist-generated / linguist-vendored are set via a bare attribute
		// ("set") or "=true"; treat only those as excluded, never unset/unspecified.
		if info == "set" || info == "true" {
			if !seen[path] {
				seen[path] = true
				excluded = append(excluded, path)
			}
		}
	}
	return excluded, nil
}
