package kfserver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/K4rian/kfdsl/cmd"
	"github.com/K4rian/kfdsl/internal/utils"

	"github.com/spf13/viper"
)

type KFServer struct {
	rootDir string
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	execErr error
}

func NewKFServer(rootDir string, ctx context.Context) *KFServer {
	kfs := &KFServer{
		rootDir: rootDir,
	}
	kfs.ctx, kfs.cancel = context.WithCancel(ctx)
	return kfs
}

func (s *KFServer) RootDirectory() string {
	return s.rootDir
}

func (s *KFServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning() {
		return fmt.Errorf("already running")
	}

	if err := s.updateSystemSteamLibs(); err != nil {
		fmt.Printf("unable to update the Steam libraries: %v\n", err)
	}

	cmdLine := s.buildCommandLine()

	cmd := exec.CommandContext(s.ctx, cmdLine[0], cmdLine[1:]...)
	cmd.Dir = path.Join(s.rootDir, "System")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %v", err)
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
		fmt.Printf("failed to send SIGTERM: %v. Attempting to kill the process...\n", err)
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

func (s *KFServer) Restart(milliseconds ...int) error {
	if err := s.ctx.Err(); err != nil {
		// The context has already been cancelled
		return nil
	}

	if err := s.Stop(); err != nil {
		return err
	}

	delay := 100
	if len(milliseconds) > 0 {
		delay = milliseconds[0]
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	if err := s.Start(); err != nil {
		return err
	}
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
	_, err := os.Stat(s.rootDir)
	_, err2 := os.Stat(path.Join(s.rootDir, "System", "ucc-bin"))
	return err == nil && err2 == nil
}

func (s *KFServer) buildCommandLine() []string {
	executable := path.Join(s.rootDir, "System", "ucc-bin")
	settings := cmd.GetSettings()
	argsStr := fmt.Sprintf(
		"%s.rom?game=%s?VACSecured=%t?MaxPlayers=%s",
		settings.StartupMap,
		settings.GameMode,
		!settings.Unsecure.Value(),
		settings.MaxPlayers,
	)

	if settings.EnableMutLoader.Value() {
		argsStr += "?Mutator=MutLoader.MutLoader"
	} else if settings.Mutators.Value() != "" {
		argsStr += fmt.Sprintf("?Mutator=%s", settings.Mutators.Value())
	}

	args := append([]string{executable}, "server", argsStr, "-nohomedir")

	extraArgs := viper.GetStringSlice("KF_EXTRAARGS")
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}
	return args
}

func (s *KFServer) updateSystemSteamLibs() error {
	systemDir := path.Join(s.rootDir, "System")
	steamLibsDir := path.Join(viper.GetString("STEAMCMD_ROOT"), "linux32")

	libs := map[string]string{
		path.Join(steamLibsDir, "steamclient.so"):  path.Join(systemDir, "steamclient.so"),
		path.Join(steamLibsDir, "libtier0_s.so"):   path.Join(systemDir, "libtier0_s.so"),
		path.Join(steamLibsDir, "libvstdlib_s.so"): path.Join(systemDir, "libvstdlib_s.so"),
	}

	for srcFile, dstFile := range libs {
		identical, err := utils.SHA1Compare(srcFile, dstFile)
		if err != nil {
			return fmt.Errorf("error comparing files %s and %s: %v", srcFile, dstFile, err)
		}

		if !identical {
			if err := utils.CopyAndReplaceFile(srcFile, dstFile); err != nil {
				return err
			}
			fmt.Printf("KFServer: Updated Steam library %s with %s\n", dstFile, srcFile)
		}
	}
	return nil
}
