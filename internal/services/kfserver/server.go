package kfserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/K4rian/dslogger"
	"github.com/creack/pty"

	"github.com/K4rian/kfdsl/internal/log"
	"github.com/K4rian/kfdsl/internal/utils"
)

type KFServer struct {
	rootDir    string
	startupMap string
	gameMode   string
	unsecure   bool
	maxPlayers int
	mutators   string
	extraArgs  []string
	executable string
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	ptmx       *os.File
	logger     *dslogger.Logger
	done       chan struct{}
	execErr    error
}

func NewKFServer(
	rootDir string,
	startupMap string,
	gameMode string,
	unsecure bool,
	maxPlayers int,
	mutators string,
	extraArgs []string,
	ctx context.Context,
) *KFServer {
	kfs := &KFServer{
		rootDir:    rootDir,
		startupMap: startupMap,
		gameMode:   gameMode,
		unsecure:   unsecure,
		maxPlayers: maxPlayers,
		mutators:   mutators,
		extraArgs:  extraArgs,
		executable: path.Join(rootDir, "System", "ucc-bin"),
		logger:     log.Logger.WithService("KFServer"),
	}
	kfs.ctx, kfs.cancel = context.WithCancel(ctx)
	return kfs
}

func (s *KFServer) RootDirectory() string {
	return s.rootDir
}

func (s *KFServer) Start(autoRestart bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning() {
		return fmt.Errorf("already running")
	}

	// Build the command line
	cmdLine := s.buildCommandLine()

	// Create a done channel to signal when the process is finished
	s.done = make(chan struct{})

	go func() {
		defer close(s.done)

		restartDelay := time.Second
		maxDelay := 10 * time.Second
		maxFailures := 5
		failureCount := 0

		for {
			// Stop restarting if the context is canceled
			if s.ctx.Err() != nil {
				return
			}

			// Set-up the command
			s.mu.Lock()
			cmd := exec.CommandContext(s.ctx, cmdLine[0], cmdLine[1:]...)
			cmd.Dir = path.Join(s.rootDir, "System")
			s.cmd = cmd

			// Reset the execution error
			s.execErr = nil

			// Start the process with a pseudo-terminal
			ptmx, err := pty.Start(s.cmd)
			if err != nil {
				s.execErr = fmt.Errorf("failed to start pty: %v", err)
				s.mu.Unlock()
				return
			}
			s.ptmx = ptmx
			s.mu.Unlock()

			// Goroutine for real-time log capture
			go func() {
				defer func() {
					s.mu.Lock()
					if s.ptmx != nil {
						s.ptmx.Close()
						s.ptmx = nil
					}
					s.mu.Unlock()
				}()

				scanner := bufio.NewScanner(ptmx)
				for scanner.Scan() {
					s.logger.Info(scanner.Text())
				}

				if err := scanner.Err(); err != nil && !errors.Is(err, syscall.EIO) {
					s.execErr = fmt.Errorf("error reading from pty: %w", err)
				}
			}()

			// Wait for process to exit
			err = cmd.Wait()

			// Clean up
			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()

			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
					s.logger.Debug("KFServer exited normally")
				} else {
					s.execErr = fmt.Errorf("process exited with error: %v", err)
				}
			}

			// Stop restarting if the context is canceled
			if s.ctx.Err() != nil {
				return
			}

			// Handle auto-restart
			if !autoRestart {
				return
			}

			// If the process crashed quickly, count it as a failure
			if restartDelay < maxDelay {
				failureCount++
				if failureCount >= maxFailures {
					s.logger.Error("KFServer failed too many times, giving up auto-restart")
					return
				}
			} else {
				failureCount = 0 // Reset failure count
			}

			s.logger.Info("Restarting the Killing Floor Dedicated Server...", "delaySeconds", restartDelay)
			select {
			case <-time.After(restartDelay):
				// Continue to restart
			case <-s.ctx.Done():
				return
			}

			// Exponential backoff (capped at maxDelay)
			restartDelay *= 2
			if restartDelay > maxDelay {
				restartDelay = maxDelay
			}
		}
	}()
	return nil
}

func (s *KFServer) Stop() error {
	s.mu.Lock()
	s.cancel()
	defer s.mu.Unlock()

	// The process is already stopped or never started
	if s.cmd == nil {
		return nil
	}

	// Send CTRL+C to gracefully terminate KFServer inside the pty
	if s.ptmx != nil {
		s.logger.Info("Attempting to send SIGINT to KFServer...")
		if _, err := s.ptmx.Write([]byte{3}); err != nil {
			s.logger.Error("Failed to send SIGINT to pty", "err", err)
		}
	}

	// Wait briefly for graceful shutdown
	time.Sleep(2 * time.Second)

	// If still running, send SIGTERM
	if s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited() {
		s.logger.Warn("Process did not exit after SIGINT, attempting SIGTERM...")
		if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			s.logger.Error("Failed to send SIGTERM, attempting to kill the process...", "err", err)

			// If SIGTERM fails, use SIGKILL
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

func (s *KFServer) Wait() error {
	<-s.done

	if err := s.Stop(); err != nil {
		return err
	}
	return s.execErr
}

func (s *KFServer) IsRunning() bool {
	return s.cmd != nil && s.cmd.Process != nil && s.cmd.ProcessState == nil
}

func (s *KFServer) IsPresent() bool {
	return utils.FileExists(s.executable)
}

func (s *KFServer) buildCommandLine() []string {
	argsStr := fmt.Sprintf(
		"%s.rom?game=%s?VACSecured=%t?MaxPlayers=%d",
		s.startupMap,
		s.gameMode,
		!s.unsecure,
		s.maxPlayers,
	)

	if s.mutators != "" {
		argsStr += fmt.Sprintf("?Mutator=%s", s.mutators)
	}

	args := append([]string{s.executable}, "server", argsStr, "-nohomedir")

	if len(s.extraArgs) > 0 {
		args = append(args, s.extraArgs...)
	}
	return args
}
