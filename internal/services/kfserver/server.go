package kfserver

import (
	"context"
	"fmt"
	"path"

	"github.com/K4rian/kfdsl/internal/services/base"
	"github.com/K4rian/kfdsl/internal/utils"
)

type KFServer struct {
	*base.BaseService
	startupMap string
	gameMode   string
	unsecure   bool
	maxPlayers int
	mutators   string
	extraArgs  []string
	executable string
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
		BaseService: base.NewBaseService("KFServer", rootDir, ctx),
		startupMap:  startupMap,
		gameMode:    gameMode,
		unsecure:    unsecure,
		maxPlayers:  maxPlayers,
		mutators:    mutators,
		extraArgs:   extraArgs,
		executable:  path.Join(rootDir, "System", "ucc-bin"),
	}
	return kfs
}

func (s *KFServer) Start(autoRestart bool) error {
	args := s.buildCommandLine()
	err := s.BaseService.Start(args, autoRestart)
	if err != nil {
		return err
	}
	return nil
}

func (s *KFServer) buildCommandLine() []string {
	argsStr := fmt.Sprintf(
		"%s.rom?game=%s?VACSecured=%t?MaxPlayers=%d",
		s.startupMap, s.gameMode, !s.unsecure, s.maxPlayers)

	if s.mutators != "" {
		argsStr += fmt.Sprintf("?Mutator=%s", s.mutators)
	}

	args := append([]string{s.executable}, "server", argsStr, "-nohomedir")
	if len(s.extraArgs) > 0 {
		args = append(args, s.extraArgs...)
	}
	return args
}

func (s *KFServer) IsAvailable() bool {
	return utils.FileExists(s.executable)
}
