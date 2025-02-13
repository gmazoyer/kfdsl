package steamcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/K4rian/dslogger"
	"github.com/creack/pty"

	"github.com/K4rian/kfdsl/internal/log"
	"github.com/K4rian/kfdsl/internal/utils"
)

type SteamCMD struct {
	rootDir string
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	ptmx    *os.File
	logger  *dslogger.Logger
	done    chan struct{}
	execErr error
}

func NewSteamCMD(rootDir string, ctx context.Context) *SteamCMD {
	scmd := &SteamCMD{
		rootDir: rootDir,
		logger:  log.Logger.WithService("SteamCMD"),
	}
	scmd.ctx, scmd.cancel = context.WithCancel(ctx)
	return scmd
}

func (s *SteamCMD) RootDirectory() string {
	return s.rootDir
}

func (s *SteamCMD) Run(args ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning() {
		return fmt.Errorf("already running")
	}

	args = append(
		[]string{filepath.Join(s.rootDir, "steamcmd.sh")},
		args...,
	)

	// Set up the command
	cmd := exec.CommandContext(s.ctx, args[0], args[1:]...)
	s.cmd = cmd

	// Reset the execution error
	s.execErr = nil

	// Start the process with a pseudo-terminal
	ptmx, err := pty.Start(s.cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %v", err)
	}
	s.ptmx = ptmx

	// Create a done channel to signal when the process is finished
	s.done = make(chan struct{})

	// Goroutine for real-time log capture and wait for process exit
	go func() {
		defer func() {
			s.mu.Lock()
			if s.ptmx != nil {
				s.ptmx.Close()
				s.ptmx = nil
			}
			s.cmd = nil
			s.mu.Unlock()
			close(s.done)
		}()

		scanner := bufio.NewScanner(ptmx)
		for scanner.Scan() {
			s.logger.Info(scanner.Text())
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, syscall.EIO) {
			s.execErr = fmt.Errorf("error reading from pty: %w", err)
		}

		// Wait for the process to exit
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
				s.logger.Debug("SteamCMD exited normally")
			} else {
				s.execErr = fmt.Errorf("process exited with error: %v", err)
			}
		}
	}()

	// Goroutine to monitor the cancellation context
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)

		<-signalChan
		signal.Stop(signalChan)

		s.mu.Lock()
		defer s.mu.Unlock()

		if s.ptmx != nil && s.IsRunning() {
			s.cancel()

			s.logger.Info("Shutting down...")
			if _, err := s.ptmx.Write([]byte{3}); err != nil {
				s.logger.Error("Failed to send SIGINT to pty", "err", err)
			}
			s.execErr = fmt.Errorf("process cancelled")
		}
	}()
	return nil
}

func (s *SteamCMD) Stop() error {
	s.mu.Lock()
	s.cancel()
	defer s.mu.Unlock()

	// The process is already stopped or never started
	if s.cmd == nil {
		return nil
	}

	// Send CTRL+C to gracefully terminate SteamCMD inside the pty
	s.logger.Info("Attempting to send SIGINT to SteamCMD...")
	if _, err := s.ptmx.Write([]byte{3}); err != nil {
		s.logger.Error("Failed to send SIGINT to pty", "err", err)
	}

	// Give some time for the process to terminate
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited() {
		s.logger.Warn("Process did not exit after SIGINT, attempting SIGTERM...")

		if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			s.logger.Error("Failed to send SIGTERM, attempting to kill the process...", "err", err)

			// If SIGTERM fails use SIGKILL
			if err := s.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to force kill process: %v", err)
			}
			s.logger.Info("Process forcefully killed")
		}
	} else {
		s.logger.Info("Process exited gracefully")
	}

	// Clean up the pseudo-terminal
	if s.ptmx != nil {
		s.ptmx.Close()
		s.ptmx = nil
	}

	s.cmd = nil
	return nil
}

func (s *SteamCMD) Wait() error {
	<-s.done

	if s.execErr != nil {
		return s.execErr
	}
	return nil
}

func (s *SteamCMD) RunScript(fileName string) error {
	if !utils.FileExists(fileName) {
		return fmt.Errorf("script file %s not found", fileName)
	}

	args := []string{
		"+runscript", fileName,
		"+quit",
	}
	return s.Run(args...)
}

func (s *SteamCMD) WriteScript(fileName string, loginUser string, loginPassword string, installDir string, appID int, validate bool) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("cannot create script file %s: %v", fileName, err)
	}
	defer file.Close()

	validateStr := "validate"
	if !validate {
		validateStr = ""
	}

	content := fmt.Sprintf(
		"force_install_dir %s\nlogin %s %s\napp_update %d %s\nquit",
		installDir,
		loginUser,
		loginPassword,
		appID,
		validateStr,
	)

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("cannot write script file %s: %v", fileName, err)
	}
	return nil
}

func (s *SteamCMD) IsRunning() bool {
	return s.cmd != nil && s.cmd.Process != nil && s.cmd.ProcessState == nil
}

func (s *SteamCMD) IsPresent() bool {
	return utils.FileExists(path.Join(s.rootDir, "steamcmd.sh"))
}
