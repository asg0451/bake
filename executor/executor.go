package executor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.coldcutz.net/bake/bakefile"
	"go.coldcutz.net/bake/options"

	"github.com/bmatcuk/doublestar/v4"
)

type Executor struct {
	config bakefile.Config
	target bakefile.Target
	opts   options.Opts
	log    *slog.Logger
}

func New(config bakefile.Config, targetName string, opts options.Opts, log *slog.Logger) (*Executor, error) {
	target, ok := config.Targets[targetName]
	if !ok {
		return nil, fmt.Errorf("unknown target %q", targetName)
	}

	return &Executor{
		config: config,
		target: target,
		opts:   opts,
		log:    log,
	}, nil
}

func (e *Executor) Exec(ctx context.Context) error {
	allDeps, err := globAll(e.target.Deps)
	if err != nil {
		return fmt.Errorf("globbing deps: %w", err)
	}
	allArtifacts, err := globAll(e.target.Artifacts)
	if err != nil {
		return fmt.Errorf("globbing artifacts: %w", err)
	}
	if len(allDeps) == 0 {
		return fmt.Errorf("no deps found")
	}

	artifactModTime, err := minMtime(allArtifacts)
	if err != nil {
		return fmt.Errorf("getting artifact mtime: %w", err)
	}
	depsModTime, err := minMtime(allDeps)
	if err != nil {
		return fmt.Errorf("getting deps mtime: %w", err)
	}
	e.log.Debug("processed deps and aftifacts", "deps", allDeps, "artifacts", allArtifacts, "mtime", artifactModTime)

	if !artifactModTime.IsZero() && artifactModTime.Before(depsModTime) {
		e.log.Debug("artifacts are up to date", "artifact_mtime", artifactModTime, "deps_mtime", depsModTime)
		return nil
	}

	if err := e.runCmd(ctx); err != nil {
		return fmt.Errorf("creating command: %w", err)
	}
	e.log.Debug("command finished successfully")
	return nil
}

func (e *Executor) runCmd(ctx context.Context) error {
	tf, err := os.CreateTemp("", "bakefile-*.sh")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tf.Name())

	// write the command to the temp file
	_, err = fmt.Fprintf(tf, `
#!/bin/bash
set -euo pipefail
%s
	`, strings.TrimSpace(e.target.Command))
	if err != nil {
		return fmt.Errorf("writing command to temp file: %w", err)
	}

	bashFlags := "-l"
	if e.opts.Verbose {
		bashFlags += "x"
	}
	cmd := exec.CommandContext(ctx, "/bin/bash", bashFlags, tf.Name())
	cmd.Env = append(os.Environ(), "BAKEFILE=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

func minMtime(files []string) (time.Time, error) {
	var mtime time.Time
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return time.Time{}, fmt.Errorf("statting %q: %w", file, err)
		}
		if mtime.IsZero() || stat.ModTime().Before(mtime) {
			mtime = stat.ModTime()
		}
	}
	return mtime, nil
}

func globAll(pats []string) ([]string, error) {
	var all []string
	for _, pat := range pats {
		matches, err := doublestar.FilepathGlob(pat, doublestar.WithFailOnIOErrors(), doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("globbing %q: %w", pat, err)
		}
		all = append(all, matches...)
	}
	return all, nil
}
