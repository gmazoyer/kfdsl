package config

import (
	"fmt"

	"github.com/K4rian/kfdsl/internal/config/ini"
	"github.com/K4rian/kfdsl/internal/settings"
)

type KFTGIniFile struct {
	*KFIniFile
	gameMode string
}

func NewKFTGIniFile(filePath string) (ServerIniFile, error) {
	iFile := &KFTGIniFile{
		KFIniFile: &KFIniFile{
			GenericIniFile: ini.NewGenericIniFile("KFTGIniFile"),
			filePath:       filePath,
		},
		gameMode: "KFCharPuppets.TOYGameInfo",
	}
	if err := iFile.Load(filePath); err != nil {
		return nil, err
	}
	return iFile, nil
}

func (kf *KFTGIniFile) GetGameLength() int {
	return kf.GetKeyInt(kf.gameMode, "KFGameLength", settings.DefaultInternalGameLength)
}

func (kf *KFTGIniFile) GetFriendlyFireRate() float64 {
	return kf.GetKeyFloat(kf.gameMode, "FriendlyFireScale", settings.DefaultFriendlyFire)
}

func (kf *KFTGIniFile) GetSpecimenType() string {
	return settings.DefaultSpecimenType
}

func (kf *KFTGIniFile) SetGameLength(length int) bool {
	// Short is the only supported length
	if length != 0 {
		kf.Logger.Warn("Invalid game length. Toy Master only supports a short game length (0). Ignoring value",
			"function", "SetGameLength", "gameLength", length)
	}

	if kf.GetGameLength() != 0 {
		kf.Logger.Warn("Invalid game length detected in ini file. Resetting to 0",
			"function", "SetGameLength")
		return kf.SetKeyInt(kf.gameMode, "KFGameLength", 0, true)
	}
	return true
}

func (kf *KFTGIniFile) SetFriendlyFireRate(rate float64) bool {
	return kf.SetKeyFloat(kf.gameMode, "FriendlyFireScale", rate, true)
}

func (kf *KFTGIniFile) SetMaxPlayers(players int) bool {
	// 6 is the maximum supported
	if players != 6 {
		kf.Logger.Warn("Invalid max players. Toy Master supports a maximum of 6 players. Ignoring value",
			"function", "SetMaxPlayers", "maxPlayers", players)
	}

	if kf.GetMaxPlayers() != 6 {
		kf.Logger.Warn("Invalid max players value detected in ini file. Resetting to 6",
			"function", "SetMaxPlayers")
		return kf.KFIniFile.SetMaxPlayers(6)
	}
	return true
}

func (kf *KFTGIniFile) SetSpecimenType(specimentype string) bool {
	if specimentype != settings.DefaultInternalSpecimenType {
		kf.Logger.Warn("Invalid specimen type. Toy Master only supports the default specimen type",
			"function", "SetSpecimenType", "specimenType", specimentype)
	}
	return true
}

func (kf *KFTGIniFile) SetMapVoteEnabled(enabled bool) error {
	if enabled {
		kf.Logger.Warn("Map voting is not needed for Toy Master, as only one map is available. Ignoring value",
			"function", "SetMapVoteEnabled", "enabled", enabled)
	}

	// Remove map list loader
	vhSection := "xVoting.xVotingHandler"
	if kf.HasKey(vhSection, "MapListLoaderType") && !kf.DeleteKey(vhSection, "MapListLoaderType") {
		return fmt.Errorf("unable to delete [%s].MapListLoaderType", vhSection)
	}
	return nil
}

func (kf *KFTGIniFile) SetMapVoteRepeatLimit(limit int) bool {
	return true
}

func (kf *KFTGIniFile) ClearMaplist(sectionName string) error {
	return nil
}

func (kf *KFTGIniFile) SetMaplist(sectionName string, maps []string) error {
	return nil
}
