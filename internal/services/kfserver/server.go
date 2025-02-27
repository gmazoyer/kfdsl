package kfserver

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/K4rian/kfdsl/internal/services/base"
	"github.com/K4rian/kfdsl/internal/utils"
)

type KFServer struct {
	*base.BaseService
	configFileName string
	startupMap     string
	gameMode       string
	unsecure       bool
	maxPlayers     int
	mutators       string
	extraArgs      []string
	executable     string
}

func NewKFServer(
	rootDir string,
	configFileName string,
	startupMap string,
	gameMode string,
	unsecure bool,
	maxPlayers int,
	mutators string,
	extraArgs []string,
	ctx context.Context,
) *KFServer {
	kfs := &KFServer{
		BaseService:    base.NewBaseService("KFServer", rootDir, ctx),
		configFileName: configFileName,
		startupMap:     startupMap,
		gameMode:       gameMode,
		unsecure:       unsecure,
		maxPlayers:     maxPlayers,
		mutators:       mutators,
		extraArgs:      extraArgs,
		executable:     path.Join(rootDir, "System", "ucc-bin"),
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
	var argsBuilder strings.Builder

	// Base command
	argsBuilder.WriteString(s.startupMap)
	argsBuilder.WriteString(".rom?game=")
	argsBuilder.WriteString(s.gameMode)
	argsBuilder.WriteString("?VACSecured=")
	argsBuilder.WriteString(fmt.Sprintf("%t", !s.unsecure))
	argsBuilder.WriteString("?MaxPlayers=")
	argsBuilder.WriteString(fmt.Sprintf("%d", s.maxPlayers))

	// Append Mutator(s) if provided
	if s.mutators != "" {
		argsBuilder.WriteString("?Mutator=")
		argsBuilder.WriteString(s.mutators)
	}

	// Specify the configuration file to use
	iniFile := "ini=" + s.configFileName

	// Final command
	args := []string{s.executable, "server", argsBuilder.String(), iniFile, "-nohomedir"}

	// Append extra arguments if provided
	if len(s.extraArgs) > 0 {
		args = append(args, s.extraArgs...)
	}
	return args
}

func (s *KFServer) IsAvailable() bool {
	return utils.FileExists(s.executable)
}
