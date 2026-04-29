package git

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// scannerMaxBuf is the per-line buffer cap for bufio.Scanner readers in the
// git package. Pumped to 64MB so single-line dumps (e.g. SQL bulk inserts
// committed as one line) cannot overflow the scanner and abandon the pipe
// in a half-read state — that combination, plus an undrained pipe, used to
// deadlock cmd.Wait() on git blame against huge files. See blame.go drain.
const scannerMaxBuf = 64 * 1024 * 1024

func RunLines(ctx context.Context, repoPath string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start git %v: %w", args, err)
	}

	var lines []string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), scannerMaxBuf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Drain anything the scanner left behind (e.g. on bufio.ErrTooLong) so
	// git's stdout doesn't stall its write side, which would block cmd.Wait
	// forever. On clean EOF this is a no-op.
	_, _ = io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		return lines, nil // partial results ok for blame
	}

	return lines, scanner.Err()
}

func RunStream(ctx context.Context, repoPath string, args ...string) (io.ReadCloser, *exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start git %v: %w", args, err)
	}

	return stdout, cmd, nil
}
