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

	"github.com/K4rian/kfdsl/internal/settings"
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

func (s *KFServer) Start(autoRestart bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning() {
		return fmt.Errorf("already running")
	}

	if err := s.updateSystemSteamLibs(); err != nil {
		fmt.Printf("unable to update the Steam libraries: %v\n", err)
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

			s.mu.Lock()
			cmd := exec.CommandContext(s.ctx, cmdLine[0], cmdLine[1:]...)
			cmd.Dir = path.Join(s.rootDir, "System")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			s.cmd = cmd
			s.mu.Unlock()

			if err := cmd.Start(); err != nil {
				s.execErr = err
				return
			}

			if err := cmd.Wait(); err != nil {
				s.execErr = fmt.Errorf("process exited with error: %v", err)
			}

			s.mu.Lock()
			s.cmd = nil
			s.mu.Unlock()

			// Stop restarting if the context is canceled
			if s.ctx.Err() != nil {
				return
			}

			if !autoRestart {
				return
			}

			fmt.Printf("> Restarting the Killing Floor Server in %v...\n", restartDelay)
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
	settings := settings.Get()
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
