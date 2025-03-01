package config

import (
	"fmt"
	"strings"

	"github.com/K4rian/kfdsl/internal/config/ini"
	"github.com/K4rian/kfdsl/internal/settings"
)

type KFIniFile struct {
	*ini.GenericIniFile
	filePath string
	gameMode string
}

const (
	// Sections
	kfSectionGameEngine           = "Engine.GameEngine"
	kfSectionGameInfo             = "Engine.GameInfo"
	kfSectionGameReplication      = "Engine.GameReplicationInfo"
	kfSectionURL                  = "URL"
	kfSectionAccessControl        = "Engine.AccessControl"
	kfSectionTcpNetDriver         = "IpDrv.TcpNetDriver"
	kfSectionUdpGamespyQuery      = "IpDrv.UdpGamespyQuery"
	kfSectionWebServer            = "UWeb.WebServer"
	kfSectionHttpDownload         = "IpDrv.HTTPDownload"
	kfSectionVotingHandler        = "xVoting.xVotingHandler"
	kfSectionDefaultMapListLoader = "xVoting.DefaultMapListLoader"
	kfSectionKFGameType           = "KFmod.KFGameType"

	// Keys
	kfKeyServerName         = "ServerName"
	kfKeyShortName          = "ShortName"
	kfKeyGamePort           = "Port"
	kfKeyWebAdminPort       = "ListenPort"
	kfKeyGameSpyPort        = "OldQueryPortNumber"
	kfKeyGameDifficulty     = "GameDifficulty"
	kfKeyGameLength         = "KFGameLength"
	kfKeyFriendlyFireRate   = "FriendlyFireScale"
	kfKeyMaxPlayers         = "MaxPlayers"
	kfKeyMaxSpectators      = "MaxSpectators"
	kfKeyPassword           = "GamePassword"
	kfKeyRegion             = "ServerRegion"
	kfKeyAdminName          = "AdminName"
	kfKeyAdminMail          = "AdminEmail"
	kfKeyAdminPassword      = "AdminPassword"
	kfKeyMOTD               = "MessageOfTheDay"
	kfKeySpecimenType       = "SpecialEventType"
	kfKeyRedirectURL        = "RedirectToURL"
	kfKeyEnableWebAdmin     = "bEnabled"
	kfKeyEnableMapVote      = "bMapVote"
	kfKeyMapVoteRepeatLimit = "RepeatLimit"
	kfKeyEnableAdminPause   = "bAdminCanPause"
	kfKeyEnableWeaponThrow  = "bAllowWeaponThrowing"
	kfKeyWeaponShakeEffect  = "bWeaponShouldViewShake"
	kfKeyEnableThirdPerson  = "bAllowBehindView"
	kfKeyEnableLowGore      = "bLowGore"
	kfKeyMaxInternetRate    = "MaxInternetClientRate"

	// Mutators
	kfKeyServerActors = "ServerActors"

	// Voting
	kfKeyMapListLoaderType = "MapListLoaderType"
	kfKeyUseMapList        = "bUseMapList"
	kfKeyMapNamePrefixes   = "MapNamePrefixes"
	kfKeyMaps              = "Maps"
	kfKeyMapNum            = "MapNum"

	// Protected Actors (lowercase)
	kfBaseActorMasterServer = "ipdrv.masterserveruplink"
	kfBaseActorWebServer    = "uweb.webserver"
)

func NewKFIniFile(filePath string) (ServerIniFile, error) {
	iFile := &KFIniFile{
		GenericIniFile: ini.NewGenericIniFile("KFIniFile"),
		filePath:       filePath,
		gameMode:       "KFmod.KFGameType",
	}
	if err := iFile.Load(filePath); err != nil {
		return nil, err
	}
	return iFile, nil
}

func (kf *KFIniFile) FilePath() string {
	return kf.filePath
}

func (kf *KFIniFile) GetServerName() string {
	return kf.GetKey(kfSectionGameReplication, kfKeyServerName, settings.DefaultServerName)
}

func (kf *KFIniFile) GetShortName() string {
	return kf.GetKey(kfSectionGameReplication, kfKeyShortName, settings.DefaultShortName)
}

func (kf *KFIniFile) GetGamePort() int {
	return kf.GetKeyInt(kfSectionURL, kfKeyGamePort, settings.DefaultGamePort)
}

func (kf *KFIniFile) GetWebAdminPort() int {
	return kf.GetKeyInt(kfSectionWebServer, kfKeyWebAdminPort, settings.DefaultWebAdminPort)
}

func (kf *KFIniFile) GetGameSpyPort() int {
	return kf.GetKeyInt(kfSectionUdpGamespyQuery, kfKeyGameSpyPort, settings.DefaultGameSpyPort)
}

func (kf *KFIniFile) GetGameDifficulty() int {
	return kf.GetKeyInt(kfSectionGameInfo, kfKeyGameDifficulty, settings.DefaultInternalGameDifficulty)
}

func (kf *KFIniFile) GetGameLength() int {
	return kf.GetKeyInt(kf.gameMode, kfKeyGameLength, settings.DefaultInternalGameLength)
}

func (kf *KFIniFile) GetFriendlyFireRate() float64 {
	return kf.GetKeyFloat(kf.gameMode, kfKeyFriendlyFireRate, settings.DefaultFriendlyFire)
}

func (kf *KFIniFile) GetMaxPlayers() int {
	return kf.GetKeyInt(kfSectionGameInfo, kfKeyMaxPlayers, settings.DefaultMaxPlayers)
}

func (kf *KFIniFile) GetMaxSpectators() int {
	return kf.GetKeyInt(kfSectionGameInfo, kfKeyMaxSpectators, settings.DefaultMaxSpectators)
}

func (kf *KFIniFile) GetPassword() string {
	return kf.GetKey(kfSectionAccessControl, kfKeyPassword, settings.DefaultPassword)
}

func (kf *KFIniFile) GetRegion() int {
	return kf.GetKeyInt(kfSectionGameReplication, kfKeyRegion, settings.DefaultRegion)
}

func (kf *KFIniFile) GetAdminName() string {
	return kf.GetKey(kfSectionGameReplication, kfKeyAdminName, settings.DefaultAdminName)
}

func (kf *KFIniFile) GetAdminMail() string {
	return kf.GetKey(kfSectionGameReplication, kfKeyAdminMail, settings.DefaultAdminMail)
}

func (kf *KFIniFile) GetAdminPassword() string {
	return kf.GetKey(kfSectionAccessControl, kfKeyAdminPassword, settings.DefaultAdminPassword)
}

func (kf *KFIniFile) GetMOTD() string {
	return kf.GetKey(kfSectionGameReplication, kfKeyMOTD, settings.DefaultMOTD)
}

func (kf *KFIniFile) GetSpecimenType() string {
	return kf.GetKey(kf.gameMode, kfKeySpecimenType, settings.DefaultInternalSpecimenType)
}

func (kf *KFIniFile) GetRedirectURL() string {
	return kf.GetKey(kfSectionHttpDownload, kfKeyRedirectURL, settings.DefaultRedirectURL)
}

func (kf *KFIniFile) IsWebAdminEnabled() bool {
	return kf.GetKeyBool(kfSectionWebServer, kfKeyEnableWebAdmin, settings.DefaultEnableWebAdmin)
}

func (kf *KFIniFile) IsMapVoteEnabled() bool {
	return kf.GetKeyBool(kfSectionVotingHandler, kfKeyEnableMapVote, settings.DefaultEnableMapVote)
}

func (kf *KFIniFile) GetMapVoteRepeatLimit() int {
	return kf.GetKeyInt(kfSectionVotingHandler, kfKeyMapVoteRepeatLimit, settings.DefaultMapVoteRepeatLimit)
}

func (kf *KFIniFile) IsAdminPauseEnabled() bool {
	return kf.GetKeyBool(kfSectionGameInfo, kfKeyEnableAdminPause, settings.DefaultEnableAdminPause)
}

func (kf *KFIniFile) IsWeaponThrowingEnabled() bool {
	return kf.GetKeyBool(kfSectionGameInfo, kfKeyEnableWeaponThrow, !settings.DefaultDisableWeaponThrow)
}

func (kf *KFIniFile) IsWeaponShakeEffectEnabled() bool {
	return kf.GetKeyBool(kfSectionGameInfo, kfKeyWeaponShakeEffect, settings.DefaultDisableWeaponShake)
}

func (kf *KFIniFile) IsThirdPersonEnabled() bool {
	return kf.GetKeyBool(kfSectionGameInfo, kfKeyEnableThirdPerson, settings.DefaultEnableThirdPerson)
}

func (kf *KFIniFile) IsLowGoreEnabled() bool {
	return kf.GetKeyBool(kfSectionGameInfo, kfKeyEnableLowGore, settings.DefaultEnableLowGore)
}

func (kf *KFIniFile) GetMaxInternetClientRate() int {
	return kf.GetKeyInt(kfSectionTcpNetDriver, kfKeyMaxInternetRate, settings.DefaultMaxInternetClientRate)
}

func (kf *KFIniFile) SetServerName(servername string) bool {
	return kf.SetKey(kfSectionGameReplication, kfKeyServerName, servername, true)
}

func (kf *KFIniFile) SetShortName(shortname string) bool {
	return kf.SetKey(kfSectionGameReplication, kfKeyShortName, shortname, true)
}

func (kf *KFIniFile) SetGamePort(port int) bool {
	return kf.SetKeyInt(kfSectionURL, kfKeyGamePort, port, true)
}

func (kf *KFIniFile) SetWebAdminPort(port int) bool {
	return kf.SetKeyInt(kfSectionWebServer, kfKeyWebAdminPort, port, true)
}

func (kf *KFIniFile) SetGameSpyPort(port int) bool {
	return kf.SetKeyInt(kfSectionUdpGamespyQuery, kfKeyGameSpyPort, port, true)
}

func (kf *KFIniFile) SetGameDifficulty(difficulty int) bool {
	return kf.SetKeyInt(kfSectionGameInfo, kfKeyGameDifficulty, difficulty, true)
}

func (kf *KFIniFile) SetGameLength(length int) bool {
	return kf.SetKeyInt(kf.gameMode, kfKeyGameLength, length, true)
}

func (kf *KFIniFile) SetFriendlyFireRate(rate float64) bool {
	return kf.SetKeyFloat(kf.gameMode, kfKeyFriendlyFireRate, rate, true)
}

func (kf *KFIniFile) SetMaxPlayers(players int) bool {
	return kf.SetKeyInt(kfSectionGameInfo, kfKeyMaxPlayers, players, true)
}

func (kf *KFIniFile) SetMaxSpectators(spectators int) bool {
	return kf.SetKeyInt(kfSectionGameInfo, kfKeyMaxSpectators, spectators, true)
}

func (kf *KFIniFile) SetPassword(password string) bool {
	return kf.SetKey(kfSectionAccessControl, kfKeyPassword, password, true)
}

func (kf *KFIniFile) SetRegion(region int) bool {
	return kf.SetKeyInt(kfSectionGameReplication, kfKeyRegion, region, true)
}

func (kf *KFIniFile) SetAdminName(adminame string) bool {
	return kf.SetKey(kfSectionGameReplication, kfKeyAdminName, adminame, true)
}

func (kf *KFIniFile) SetAdminMail(adminmail string) bool {
	return kf.SetKey(kfSectionGameReplication, kfKeyAdminMail, adminmail, true)
}

func (kf *KFIniFile) SetAdminPassword(adminpassword string) bool {
	return kf.SetKey(kfSectionAccessControl, kfKeyAdminPassword, adminpassword, true)
}

func (kf *KFIniFile) SetMOTD(motd string) bool {
	return kf.SetKey(kfSectionGameReplication, kfKeyMOTD, motd, true)
}

func (kf *KFIniFile) SetSpecimenType(specimentype string) bool {
	return kf.SetKey(kf.gameMode, kfKeySpecimenType, specimentype, true)
}

func (kf *KFIniFile) SetRedirectURL(url string) bool {
	return kf.SetKey(kfSectionHttpDownload, kfKeyRedirectURL, url, true)
}

func (kf *KFIniFile) SetWebAdminEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionWebServer, kfKeyEnableWebAdmin, enabled, true)
}

func (kf *KFIniFile) SetMapVoteEnabled(enabled bool) error {
	if !kf.SetKeyBool(kfSectionVotingHandler, kfKeyEnableMapVote, enabled, true) {
		return fmt.Errorf("unable to set %s.%s to %t", kfSectionVotingHandler, kfKeyEnableMapVote, enabled)
	}

	if enabled {
		if !kf.SetKey(kfSectionVotingHandler, kfKeyMapListLoaderType, kfSectionDefaultMapListLoader, true) {
			return fmt.Errorf("unable to set %s.%s to %s", kfSectionVotingHandler, kfKeyMapListLoaderType, kfSectionDefaultMapListLoader)
		}
		if !kf.SetKeyBool(kfSectionDefaultMapListLoader, kfKeyUseMapList, true, true) {
			return fmt.Errorf("unable to set %s.%s to true", kfSectionDefaultMapListLoader, kfKeyUseMapList)
		}
		if !kf.SetKey(kfSectionDefaultMapListLoader, kfKeyMapNamePrefixes, "", true) {
			return fmt.Errorf("unable to clear %s.%s", kfSectionDefaultMapListLoader, kfKeyMapNamePrefixes)
		}
	} else {
		if !kf.SetKey(kfSectionVotingHandler, kfKeyMapListLoaderType, "", true) {
			return fmt.Errorf("unable to clear %s.%s", kfSectionVotingHandler, kfKeyMapListLoaderType)
		}
	}
	return nil
}

func (kf *KFIniFile) SetMapVoteRepeatLimit(limit int) bool {
	return kf.SetKeyInt(kfSectionVotingHandler, kfKeyMapVoteRepeatLimit, limit, true)
}

func (kf *KFIniFile) SetAdminPauseEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionGameInfo, kfKeyEnableAdminPause, enabled, true)
}

func (kf *KFIniFile) SetWeaponThrowingEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionGameInfo, kfKeyEnableWeaponThrow, enabled, true)
}

func (kf *KFIniFile) SetWeaponShakeEffectEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionGameInfo, kfKeyWeaponShakeEffect, enabled, true)
}

func (kf *KFIniFile) SetThirdPersonEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionGameInfo, kfKeyEnableThirdPerson, enabled, true)
}

func (kf *KFIniFile) SetLowGoreEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfSectionGameInfo, kfKeyEnableLowGore, enabled, true)
}

func (kf *KFIniFile) SetMaxInternetClientRate(rate int) bool {
	return kf.SetKeyInt(kfSectionTcpNetDriver, kfKeyMaxInternetRate, rate, true)
}

func (kf *KFIniFile) ServerMutatorExists(mutator string) bool {
	mutator = strings.ToLower(strings.TrimSpace(mutator))
	actors := kf.GetKeys(kfSectionGameEngine, kfKeyServerActors)
	for _, actor := range actors {
		if strings.EqualFold(strings.ToLower(strings.TrimSpace(actor)), mutator) {
			return true
		}
	}
	return false
}

func (kf *KFIniFile) ClearServerMutators() error {
	// Get server actors
	actors := kf.GetKeys(kfSectionGameEngine, kfKeyServerActors)

	// Base actors that must NOT be removed
	baseActors := map[string]struct{}{
		kfBaseActorMasterServer: {},
		kfBaseActorWebServer:    {},
	}

	for _, actor := range actors {
		act := strings.TrimSpace(actor)

		if _, exists := baseActors[strings.ToLower(act)]; !exists {
			if !kf.DeleteUniqueKey(kfSectionGameEngine, kfKeyServerActors, &actor, nil) {
				return fmt.Errorf("unable to delete ServerActor: %s", actor)
			}
		}
	}
	return nil
}

func (kf *KFIniFile) SetServerMutators(mutators []string) error {
	if len(mutators) > 0 {
		for _, mutator := range mutators {
			// Don't add the same mutator twice
			if kf.ServerMutatorExists(mutator) {
				continue
			}

			if added := kf.SetKey(kfSectionGameEngine, kfKeyServerActors, mutator, false); !added {
				return fmt.Errorf("unable to add mutator as ServerActor: %s", mutator)
			}
		}
	}
	return nil
}

func (kf *KFIniFile) ClearMaplist(sectionName string) error {
	if section := kf.GetSection(sectionName); section != nil {
		section.DeleteKey(kfKeyMaps)

		if len(section.GetKeys(kfKeyMaps)) > 0 {
			return fmt.Errorf("unable to clear the maplist: %s", sectionName)
		}
	}
	return nil
}

func (kf *KFIniFile) SetMaplist(sectionName string, maps []string) error {
	// Create the section if it doesn't exist
	section := kf.GetSection(sectionName)
	if section == nil {
		if added := kf.SetKeyInt(sectionName, kfKeyMapNum, 0, true); !added {
			return fmt.Errorf("unable to create the maplist section '%s'", sectionName)
		}
		section = kf.GetSection(sectionName)
	} else {
		// Clear all existing maps
		section.DeleteKey(kfKeyMaps)
	}

	for _, m := range maps {
		section.SetKey(kfKeyMaps, m)
	}
	return nil
}
