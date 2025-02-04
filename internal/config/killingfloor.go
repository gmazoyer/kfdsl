package config

import (
	"fmt"
	"strings"

	"kfdsl/internal/config/ini"
)

type KFIniFile struct {
	*ini.GenericIniFile
	FilePath string
}

func NewKFIniFile(filePath string) (*KFIniFile, error) {
	kfIniFile := &KFIniFile{
		GenericIniFile: ini.NewGenericIniFile(),
		FilePath:       filePath,
	}
	if err := kfIniFile.Load(filePath); err != nil {
		return nil, err
	}
	return kfIniFile, nil
}

func (kf *KFIniFile) GetServerName() string {
	return kf.GetKey("Engine.GameReplicationInfo", "ServerName", "Killing Floor Server")
}

func (kf *KFIniFile) GetShortName() string {
	return kf.GetKey("Engine.GameReplicationInfo", "ShortName", "KF Server")
}

func (kf *KFIniFile) GetGamePort() int {
	return kf.GetKeyInt("URL", "Port", 7707)
}

func (kf *KFIniFile) GetWebAdminPort() int {
	return kf.GetKeyInt("UWeb.WebServer", "ListenPort", 8075)
}

func (kf *KFIniFile) GetGameSpyPort() int {
	return kf.GetKeyInt("IpDrv.UdpGamespyQuery", "OldQueryPortNumber", 7717)
}

func (kf *KFIniFile) GetGameDifficulty() int {
	return kf.GetKeyInt("Engine.GameInfo", "GameDifficulty", 2)
}

func (kf *KFIniFile) GetGameLength() int {
	return kf.GetKeyInt("KFMod.KFGameType", "KFGameLength", 0)
}

func (kf *KFIniFile) GetFriendlyFireRate() float64 {
	return kf.GetKeyFloat("KFMod.KFGameType", "FriendlyFireScale", 0.0)
}

func (kf *KFIniFile) GetMaxPlayers() int {
	return kf.GetKeyInt("Engine.GameInfo", "MaxPlayers", 6)
}

func (kf *KFIniFile) GetMaxSpectators() int {
	return kf.GetKeyInt("Engine.GameInfo", "MaxSpectators", 0)
}

func (kf *KFIniFile) GetPassword() string {
	return kf.GetKey("Engine.AccessControl", "GamePassword", "")
}

func (kf *KFIniFile) GetRegion() int {
	return kf.GetKeyInt("Engine.GameReplicationInfo", "ServerRegion", 0)
}

func (kf *KFIniFile) GetAdminName() string {
	return kf.GetKey("Engine.GameReplicationInfo", "AdminName", "")
}

func (kf *KFIniFile) GetAdminMail() string {
	return kf.GetKey("Engine.GameReplicationInfo", "AdminEmail", "")
}

func (kf *KFIniFile) GetAdminPassword() string {
	return kf.GetKey("Engine.AccessControl", "AdminPassword", "")
}

func (kf *KFIniFile) GetMOTD() string {
	return kf.GetKey("Engine.GameReplicationInfo", "MessageOfTheDay", "")
}

func (kf *KFIniFile) GetSpecimenType() string {
	return kf.GetKey("KFMod.KFGameType", "SpecialEventType", "ET_None")
}

func (kf *KFIniFile) GetRedirectURL() string {
	return kf.GetKey("IpDrv.HTTPDownload", "RedirectToURL", "")
}

func (kf *KFIniFile) IsWebAdminEnabled() bool {
	return kf.GetKeyBool("UWeb.WebServer", "bEnabled", false)
}

func (kf *KFIniFile) IsMapVoteEnabled() bool {
	return kf.GetKeyBool("xVoting.xVotingHandler", "bMapVote", false)
}

func (kf *KFIniFile) GetMapVoteRepeatLimit() int {
	return kf.GetKeyInt("xVoting.xVotingHandler", "RepeatLimit", 1)
}

func (kf *KFIniFile) IsAdminPauseEnabled() bool {
	return kf.GetKeyBool("Engine.GameInfo", "bAdminCanPause", false)
}

func (kf *KFIniFile) IsWeaponThrowingEnabled() bool {
	return kf.GetKeyBool("Engine.GameInfo", "bAllowWeaponThrowing", true)
}

func (kf *KFIniFile) IsWeaponShakeEffectEnabled() bool {
	return kf.GetKeyBool("Engine.GameInfo", "bWeaponShouldViewShake", true)
}

func (kf *KFIniFile) IsThirdPersonEnabled() bool {
	return kf.GetKeyBool("Engine.GameInfo", "bAllowBehindView", false)
}

func (kf *KFIniFile) IsLowGoreEnabled() bool {
	return kf.GetKeyBool("Engine.GameInfo", "bLowGore", false)
}

func (kf *KFIniFile) GetMaxInternetClientRate() int {
	return kf.GetKeyInt("IpDrv.TcpNetDriver", "MaxInternetClientRate", 10000)
}

func (kf *KFIniFile) IsFPSUncapped() bool {
	return kf.GetMaxInternetClientRate() > 10000
}

// -----------------------

func (kf *KFIniFile) SetServerName(servername string) bool {
	return kf.SetKey("Engine.GameReplicationInfo", "ServerName", servername, true)
}

func (kf *KFIniFile) SetShortName(shortname string) bool {
	return kf.SetKey("Engine.GameReplicationInfo", "ShortName", shortname, true)
}

func (kf *KFIniFile) SetGamePort(port int) bool {
	return kf.SetKeyInt("URL", "Port", port, true)
}

func (kf *KFIniFile) SetWebAdminPort(port int) bool {
	return kf.SetKeyInt("UWeb.WebServer", "ListenPort", port, true)
}

func (kf *KFIniFile) SetGameSpyPort(port int) bool {
	return kf.SetKeyInt("IpDrv.UdpGamespyQuery", "OldQueryPortNumber", port, true)
}

func (kf *KFIniFile) SetGameDifficulty(difficulty int) bool {
	return kf.SetKeyInt("Engine.GameInfo", "GameDifficulty", difficulty, true)
}

func (kf *KFIniFile) SetGameLength(length int) bool {
	return kf.SetKeyInt("KFMod.KFGameType", "KFGameLength", length, true)
}

func (kf *KFIniFile) SetFriendlyFireRate(rate float64) bool {
	return kf.SetKeyFloat("KFMod.KFGameType", "FriendlyFireScale", rate, true)
}

func (kf *KFIniFile) SetMaxPlayers(players int) bool {
	return kf.SetKeyInt("Engine.GameInfo", "MaxPlayers", players, true)
}

func (kf *KFIniFile) SetMaxSpectators(spectators int) bool {
	return kf.SetKeyInt("Engine.GameInfo", "MaxSpectators", spectators, true)
}

func (kf *KFIniFile) SetPassword(password string) bool {
	return kf.SetKey("Engine.AccessControl", "GamePassword", password, true)
}

func (kf *KFIniFile) SetRegion(region int) bool {
	return kf.SetKeyInt("Engine.GameReplicationInfo", "ServerRegion", region, true)
}

func (kf *KFIniFile) SetAdminName(adminame string) bool {
	return kf.SetKey("Engine.GameReplicationInfo", "AdminName", adminame, true)
}

func (kf *KFIniFile) SetAdminMail(adminmail string) bool {
	return kf.SetKey("Engine.GameReplicationInfo", "AdminEmail", adminmail, true)
}

func (kf *KFIniFile) SetAdminPassword(adminpassword string) bool {
	return kf.SetKey("Engine.AccessControl", "AdminPassword", adminpassword, true)
}

func (kf *KFIniFile) SetMOTD(motd string) bool {
	return kf.SetKey("Engine.GameReplicationInfo", "MessageOfTheDay", motd, true)
}

func (kf *KFIniFile) SetSpecimenType(specimentype string) bool {
	return kf.SetKey("KFMod.KFGameType", "SpecialEventType", specimentype, true)
}

func (kf *KFIniFile) SetRedirectURL(url string) bool {
	return kf.SetKey("IpDrv.HTTPDownload", "RedirectToURL", url, true)
}

func (kf *KFIniFile) SetWebAdminEnabled(enabled bool) bool {
	return kf.SetKeyBool("UWeb.WebServer", "bEnabled", enabled, true)
}

func (kf *KFIniFile) SetMapVoteEnabled(enabled bool) error {
	votingSection := "xVoting.xVotingHandler"

	if !kf.SetKeyBool(votingSection, "bMapVote", enabled, true) {
		return fmt.Errorf("unable to set %s.bMapVote to %t", votingSection, enabled)
	}

	if enabled {
		loaderSection := "xVoting.DefaultMapListLoader"

		if !kf.SetKey(votingSection, "MapListLoaderType", loaderSection, true) {
			return fmt.Errorf("unable to set %s.MapListLoaderType to %s", votingSection, loaderSection)
		}
		if !kf.SetKeyBool(loaderSection, "bUseMapList", true, true) {
			return fmt.Errorf("unable to set %s.bUseMapList to true", loaderSection)
		}
		if !kf.SetKey(loaderSection, "MapNamePrefixes", "", true) {
			return fmt.Errorf("unable to clear %s.MapNamePrefixes", loaderSection)
		}
	} else {
		if !kf.SetKey(votingSection, "MapListLoaderType", "", true) {
			return fmt.Errorf("unable to clear %s.MapListLoaderType", votingSection)
		}
	}
	return nil
}

func (kf *KFIniFile) SetMapVoteRepeatLimit(limit int) bool {
	return kf.SetKeyInt("xVoting.xVotingHandler", "RepeatLimit", limit, true)
}

func (kf *KFIniFile) SetAdminPauseEnabled(enabled bool) bool {
	return kf.SetKeyBool("Engine.GameInfo", "bAdminCanPause", enabled, true)
}

func (kf *KFIniFile) SetWeaponThrowingEnabled(enabled bool) bool {
	return kf.SetKeyBool("Engine.GameInfo", "bAllowWeaponThrowing", enabled, true)
}

func (kf *KFIniFile) SetWeaponShakeEffectEnabled(enabled bool) bool {
	return kf.SetKeyBool("Engine.GameInfo", "bWeaponShouldViewShake", enabled, true)
}

func (kf *KFIniFile) SetThirdPersonEnabled(enabled bool) bool {
	return kf.SetKeyBool("Engine.GameInfo", "bAllowBehindView", enabled, true)
}

func (kf *KFIniFile) SetLowGoreEnabled(enabled bool) bool {
	return kf.SetKeyBool("Engine.GameInfo", "bLowGore", enabled, true)
}

func (kf *KFIniFile) SetMaxInternetClientRate(rate int) bool {
	return kf.SetKeyInt("IpDrv.TcpNetDriver", "MaxInternetClientRate", rate, true)
}

// ----------------------

func (kf *KFIniFile) ClearServerMutators() error {
	section, key := "Engine.GameEngine", "ServerActors"
	actors := kf.GetAllKeys(section, key)
	for index, actor := range actors {
		act := strings.TrimSpace(actor)
		if act != "IpDrv.MasterServerUplink" && act != "UWeb.WebServer" { // Ignore Base Actors
			if deleted := kf.DeleteUniqueKey(section, key, &actor, &index); !deleted {
				return fmt.Errorf("unable to delete ServerActor: %s", actor)
			}
		}
	}
	return nil
}

func (kf *KFIniFile) SetServerMutators(mutators []string) error {
	section, key := "Engine.GameEngine", "ServerActors"
	if len(mutators) > 0 {
		for _, mutator := range mutators {
			if added := kf.SetKey(section, key, mutator, false); !added {
				return fmt.Errorf("unable to add mutator as ServerActor: %s", mutator)
			}
		}
	}
	return nil
}

func (kf *KFIniFile) ClearMaplist(sectionName string) error {
	section := kf.GetSection(sectionName)
	if section != nil {
		section.DeleteKey("Maps")
		if len(section.GetAllKeys("Maps")) > 0 {
			return fmt.Errorf("unable to clear the maplist: %s", sectionName)
		}
	}
	return nil
}

func (kf *KFIniFile) SetMaplist(sectionName string, maps []string) error {
	// Create the section if it doesn't exist
	section := kf.GetSection(sectionName)
	if section == nil {
		if added := kf.SetKeyInt(sectionName, "MapNum", 0, true); !added {
			return fmt.Errorf("unable to create the maplist section '%s'", sectionName)
		}
		section = kf.GetSection(sectionName)
	} else {
		// Clear all existing maps
		section.DeleteKey("Maps")
	}

	for _, m := range maps {
		section.SetKey("Maps", m)
	}
	return nil
}
