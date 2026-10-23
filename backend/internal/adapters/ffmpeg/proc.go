package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const (
	// maxStderr bounds how much of ffmpeg's error output is kept.
	maxStderr = 4096
	waitDelay = 2 * time.Second
)

// proc is a running ffmpeg whose stdout is read as a stream.
type proc struct {
	ctx    context.Context
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr *capped
	done   bool
	err    error
}

func start(ctx context.Context, bin string, args []string) (*proc, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	// Don't hang in Wait if a child process keeps the pipes open.
	cmd.WaitDelay = waitDelay
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout: %w", err)
	}
	p := &proc{
		ctx:    ctx,
		cmd:    cmd,
		stdout: stdout,
		stderr: &capped{max: maxStderr},
	}
	cmd.Stderr = p.stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting %s: %w", bin, err)
	}
	return p, nil
}

// Read reads ffmpeg's stdout. At the end of the stream it waits for the
// process and returns its failure, if any, instead of io.EOF.
func (p *proc) Read(b []byte) (int, error) {
	if p.done {
		return 0, p.endErr()
	}
	n, err := p.stdout.Read(b)
	if errors.Is(err, io.EOF) {
		p.wait()
		return n, p.endErr()
	}
	return n, err
}

// Close stops ffmpeg if the stream was not read to the end.
func (p *proc) Close() error {
	if !p.done {
		_ = p.cmd.Process.Kill()
		p.wait()
	}
	return nil
}

func (p *proc) wait() {
	p.done = true
	err := p.cmd.Wait()
	switch {
	case p.ctx.Err() != nil:
		p.err = p.ctx.Err()
	case err != nil:
		p.err = fmt.Errorf("ffmpeg: %w: %s", err, bytes.TrimSpace(p.stderr.buf.Bytes()))
	}
}

func (p *proc) endErr() error {
	if p.err != nil {
		return p.err
	}
	return io.EOF
}

// capped keeps the first max bytes written to it and drops the rest.
type capped struct {
	buf bytes.Buffer
	max int
}

func (c *capped) Write(b []byte) (int, error) {
	if room := c.max - c.buf.Len(); room > 0 {
		c.buf.Write(b[:min(len(b), room)])
	}
	return len(b), nil
}
