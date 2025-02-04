package steamcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"syscall"
)

type SteamCMD struct {
	rootDir string
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	execErr error
}

func NewSteamCMD(rootDir string, ctx context.Context) *SteamCMD {
	scmd := &SteamCMD{
		rootDir: rootDir,
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
		[]string{
			filepath.Join(s.rootDir, "steamcmd.sh"),
		},
		args...,
	)

	cmd := exec.CommandContext(s.ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start SteamCMD: %v", err)
	}
	s.cmd = cmd

	go func() {
		defer s.cancel()

		if err := cmd.Wait(); err != nil {
			s.execErr = fmt.Errorf("process exited with error: %v", err)
		}
		s.mu.Lock()
		s.cmd = nil
		s.mu.Unlock()
	}()
	return nil
}

func (s *SteamCMD) Stop() error {
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
		fmt.Printf("failed to send SIGTERM: %v. Attempting to kill the process...\n", err)
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop SteamCMD process: %v", err)
		}
	}

	// Wait for the process to exit, if still running
	if s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited() {
		if err := s.cmd.Wait(); err != nil {
			// Skip the 'no child process' error (ECHILD)
			var syscallErr *os.SyscallError
			if errors.As(err, &syscallErr) && syscallErr.Err != syscall.ECHILD {
				return fmt.Errorf("failed to wait for SteamCMD to exit: %v", err)
			}
		}
	}
	s.cmd = nil
	return nil
}

func (s *SteamCMD) Wait() error {
	<-s.ctx.Done()
	if err := s.Stop(); err != nil {
		return err
	}
	if s.execErr != nil {
		return s.execErr
	}
	return nil
}

func (s *SteamCMD) RunScript(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
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
	_, err := os.Stat(s.rootDir)
	_, err2 := os.Stat(path.Join(s.rootDir, "steamcmd.sh"))
	return err == nil && err2 == nil
}
