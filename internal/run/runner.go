package run

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	ModeSmart          Mode = "Smart"
	ModeClean          Mode = "Clean"
	ModeValidationOnly Mode = "ValidationOnly"
)

type Mode string

type RunSpec struct {
	Name    string
	Root    string
	Mode    Mode
	Command string
	Args    []string
}

type Result struct {
	Status   string
	Stdout   string
	Stderr   string
	ExitCode int
}

func Execute(ctx context.Context, spec RunSpec) Result {
	logDir := filepath.Join(spec.Root, "tmp", "augury-node-tui")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return Result{Status: "error", ExitCode: -1}
	}
	logPath := filepath.Join(logDir, spec.Name+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return Result{Status: "error", ExitCode: -1}
	}
	defer logFile.Close()

	cmd := exec.CommandContext(ctx, spec.Command, spec.Args...)
	cmd.Dir = spec.Root

	if spec.Mode == ModeClean {
		cmd.Env = append(os.Environ(), "CLEAN=1")
	} else {
		cmd.Env = os.Environ()
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return Result{Status: "error", ExitCode: -1}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return Result{Status: "error", ExitCode: -1}
	}

	if err := cmd.Start(); err != nil {
		if ctx.Err() != nil {
			return Result{Status: "cancelled"}
		}
		return Result{Status: "error", ExitCode: -1}
	}

	var stdoutBuf, stderrBuf strings.Builder
	var wg sync.WaitGroup
	tee := func(r io.Reader, buf *strings.Builder, log io.Writer) {
		defer wg.Done()
		mw := io.MultiWriter(buf, log)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			mw.Write([]byte(line))
		}
	}
	wg.Add(2)
	go tee(stdoutPipe, &stdoutBuf, logFile)
	go tee(stderrPipe, &stderrBuf, logFile)

	err = cmd.Wait()
	wg.Wait()

	if ctx.Err() != nil {
		return Result{
			Status: "cancelled",
			Stdout: stdoutBuf.String(),
			Stderr: stderrBuf.String(),
		}
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Result{
				Status:   "error",
				Stdout:   stdoutBuf.String(),
				Stderr:   stderrBuf.String(),
				ExitCode: exitErr.ExitCode(),
			}
		}
		return Result{Status: "error", ExitCode: -1}
	}

	return Result{
		Status:   "success",
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: 0,
	}
}
