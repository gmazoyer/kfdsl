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
	logger     *dslogger.Logger
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

	cmdLine := s.buildCommandLine()

	go func() {
		restartDelay := time.Second
		maxDelay := 10 * time.Second

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

			// Start the process with a pseudo-terminal
			ptmx, err := pty.Start(s.cmd)
			if err != nil {
				s.execErr = fmt.Errorf("failed to start pty: %v", err)
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()

			// Goroutine for real-time log capture
			go func() {
				defer ptmx.Close()

				scanner := bufio.NewScanner(ptmx)
				for scanner.Scan() {
					s.logger.Info(scanner.Text())
				}

				if err := scanner.Err(); err != nil {
					s.execErr = fmt.Errorf("error reading from pty: %v", err)
				}
			}()

			// Wait for process to exit
			if err := cmd.Wait(); err != nil {
				s.execErr = fmt.Errorf("process exited with error: %v", err)
			}

			// Clean up
			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()

			// Stop restarting if the context is canceled
			if s.ctx.Err() != nil {
				return
			}

			// Handle auto-restart
			if !autoRestart {
				return
			}

			s.logger.Info("Restarting the Killing Floor Dedicated Server...", "delaySeconds", restartDelay)
			time.Sleep(restartDelay)

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
	defer s.mu.Unlock()

	// Already stopped
	if s.cmd != nil && s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
		s.cmd = nil
		return nil
	}

	// Already stopped or never started
	if s.cmd == nil {
		s.cmd = nil
		return nil
	}

	// The context has already been cancelled
	if err := s.ctx.Err(); err != nil {
		return nil
	}

	// Send a SIGTERM signal to the server process
	// If the server process refuses to exit, kill it
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		s.logger.Error("failed to send SIGTERM. Attempting to kill the process...", "err", err)
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop server process: %v", err)
		}
	}

	// Wait for the process to exit, if still running
	if s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited() {
		if err := s.cmd.Wait(); err != nil {
			// Skip the 'no child process' error (ECHILD)
			var syscallErr *os.SyscallError
			if errors.As(err, &syscallErr) && syscallErr.Err != syscall.ECHILD {
				return fmt.Errorf("failed to wait for server to exit: %v", err)
			}
		}
	}
	s.cmd = nil
	return nil
}

func (s *KFServer) Wait() error {
	<-s.ctx.Done()
	if err := s.Stop(); err != nil {
		return err
	}
	if s.execErr != nil {
		return s.execErr
	}
	return nil
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
