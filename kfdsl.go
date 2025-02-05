package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/K4rian/kfdsl/internal/config"
	"github.com/K4rian/kfdsl/internal/config/secrets"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/services/redirectserver"
	"github.com/K4rian/kfdsl/internal/services/steamcmd"
	"github.com/K4rian/kfdsl/internal/settings"
)

type configUpdater[T any] struct {
	name string       // Name
	gv   func() T     // Getter
	sv   func(T) bool // Setter
	nv   any          // Value to set
}

func newConfigUpdater[T any](name string, gv func() T, sv func(T) bool, nv any) configUpdater[T] {
	return configUpdater[T]{
		name: name,
		gv:   gv,
		sv:   sv,
		nv:   nv,
	}
}

const (
	KF_APPID = 215360
)

func startSteamCMD(sett *settings.KFDSLSettings, ctx context.Context) error {
	steamCMD := steamcmd.NewSteamCMD(viper.GetString("STEAMCMD_ROOT"), ctx)

	if !steamCMD.IsPresent() {
		return fmt.Errorf("unable to locate SteamCMD in '%s', please install it first", steamCMD.RootDirectory())
	}

	// Read Steam Account Credentials
	if err := readSteamCredentials(sett); err != nil {
		return fmt.Errorf("credentials error: %v", err)
	}

	// Generate the Steam install script
	installScript := path.Join(viper.GetString("STEAMCMD_ROOT"), "kfds_install_script.txt")
	serverInstallDir := viper.GetString("STEAMCMD_APPINSTALLDIR")

	if err := steamCMD.WriteScript(
		installScript,
		sett.SteamLogin,
		sett.SteamPassword,
		serverInstallDir,
		KF_APPID,
		!sett.NoValidate.Value(),
	); err != nil {
		return err
	}

	fmt.Printf("> Starting SteamCMD...\n")
	if err := steamCMD.RunScript(installScript); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	// Block until SteamCMD finishes
	if err := steamCMD.Wait(); err != nil {
		return err
	}
	return nil
}

func startGameServer(sett *settings.KFDSLSettings, ctx context.Context) (*kfserver.KFServer, error) {
	gameServer := kfserver.NewKFServer(viper.GetString("STEAMCMD_APPINSTALLDIR"), ctx)

	if !gameServer.IsPresent() {
		return nil, fmt.Errorf("unable to locate the Killing Floor Dedicated Server files in '%s', please install using SteamCMD", gameServer.RootDirectory())
	}

	fmt.Printf("> Updating the Killing Floor configuration file...\n")
	if err := updateConfigFile(sett); err != nil {
		return nil, fmt.Errorf("failed to update the KillingFloor.ini configuration file: %v", err)
	}

	if sett.EnableKFPatcher.Value() {
		fmt.Printf("> Updating the KFPatcher configuration file...\n")
		if err := updateKFPatcherConfigFile(sett); err != nil {
			return nil, fmt.Errorf("failed to update the KFPatcherSettings.ini configuration file: %v", err)
		}
	}

	fmt.Printf("> Starting the Killing Floor Server...\n")
	if err := gameServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start the Killing Floor Dedicated Server: %v", err)
	}
	return gameServer, nil
}

func startRedirectServer(sett *settings.KFDSLSettings, ctx context.Context) (*redirectserver.HTTPRedirectServer, error) {
	redirectServer := redirectserver.NewHTTPRedirectServer(
		sett.RedirectServerHost.Value(),
		sett.RedirectServerPort.Value(),
		sett.RedirectServerDir.Value(),
		sett.RedirectServerMaxRequests.Value(),
		sett.RedirectServerBanTime.Value(),
		ctx,
	)

	fmt.Printf("> Starting the HTTP Redirect Server...\n")
	if err := redirectServer.Listen(); err != nil {
		return nil, fmt.Errorf("failed to start the HTTP Redirect Server: %v", err)
	}
	return redirectServer, nil
}

func updateConfigFileServerMutators(iniFile *config.KFIniFile, sett *settings.KFDSLSettings) error {
	mutatorsStr := sett.ServerMutators.Value()
	mutators := strings.FieldsFunc(mutatorsStr, func(r rune) bool { return r == ',' })

	if sett.EnableKFPatcher.Value() && !strings.Contains(mutatorsStr, "KFPatcher") {
		mutators = append(mutators, "KFPatcher.Mut")
	}

	if len(mutators) > 0 {
		if err := iniFile.SetServerMutators(mutators); err != nil {
			return err
		}
	} else {
		if err := iniFile.ClearServerMutators(); err != nil {
			return err
		}
	}
	return nil
}

func updateConfigFileMaplist(iniFile *config.KFIniFile, sett *settings.KFDSLSettings) error {
	gameMode := sett.GameMode.RawValue()

	sectionName := kfserver.GetGameModeMaplistName(gameMode)
	if sectionName == "" {
		return fmt.Errorf("undefined section name for game mode: %s", gameMode)
	}

	mapList := strings.FieldsFunc(sett.Maplist.Value(), func(r rune) bool { return r == ',' })

	if len(mapList) > 0 {
		if mapList[0] == "all" {
			gameServerRoot := viper.GetString("STEAMCMD_APPINSTALLDIR")
			gameModePrefix := kfserver.GetGameModeMapPrefix(gameMode)
			installedMaps, err := kfserver.GetInstalledMaps(
				path.Join(gameServerRoot, "Maps"),
				gameModePrefix,
			)
			if err != nil {
				return fmt.Errorf("unable to fetch available maps for game mode '%s': %v", gameMode, err)
			}
			mapList = installedMaps
		}

		if err := iniFile.SetMaplist(sectionName, mapList); err != nil {
			return err
		}
	} else {
		if err := iniFile.ClearMaplist(sectionName); err != nil {
			return err
		}
	}
	return nil
}

func updateConfigFile(sett *settings.KFDSLSettings) error {
	kfiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KillingFloor.ini")
	kfi, err := config.NewKFIniFile(kfiFilePath)
	if err != nil {
		return err
	}

	// Generic config
	cuList := []configUpdater[any]{
		newConfigUpdater(sett.ServerName.Name(), func() any { return kfi.GetServerName() }, func(v any) bool { return kfi.SetServerName(v.(string)) }, sett.ServerName.Value()),
		newConfigUpdater(sett.ShortName.Name(), func() any { return kfi.GetShortName() }, func(v any) bool { return kfi.SetShortName(v.(string)) }, sett.ShortName.Value()),
		newConfigUpdater(sett.GamePort.Name(), func() any { return kfi.GetGamePort() }, func(v any) bool { return kfi.SetGamePort(v.(int)) }, sett.GamePort.Value()),
		newConfigUpdater(sett.WebAdminPort.Name(), func() any { return kfi.GetWebAdminPort() }, func(v any) bool { return kfi.SetWebAdminPort(v.(int)) }, sett.WebAdminPort.Value()),
		newConfigUpdater(sett.GameSpyPort.Name(), func() any { return kfi.GetGameSpyPort() }, func(v any) bool { return kfi.SetGameSpyPort(v.(int)) }, sett.GameSpyPort.Value()),
		newConfigUpdater(sett.GameDifficulty.Name(), func() any { return kfi.GetGameDifficulty() }, func(v any) bool { return kfi.SetGameDifficulty(v.(int)) }, sett.GameDifficulty.Value()),
		newConfigUpdater(sett.GameLength.Name(), func() any { return kfi.GetGameLength() }, func(v any) bool { return kfi.SetGameLength(v.(int)) }, sett.GameLength.Value()),
		newConfigUpdater(sett.FriendlyFire.Name(), func() any { return kfi.GetFriendlyFireRate() }, func(v any) bool { return kfi.SetFriendlyFireRate(v.(float64)) }, sett.FriendlyFire.Value()),
		newConfigUpdater(sett.MaxPlayers.Name(), func() any { return kfi.GetMaxPlayers() }, func(v any) bool { return kfi.SetMaxPlayers(v.(int)) }, sett.MaxPlayers.Value()),
		newConfigUpdater(sett.Password.Name(), func() any { return kfi.GetPassword() }, func(v any) bool { return kfi.SetPassword(v.(string)) }, sett.Password.Value()),
		newConfigUpdater(sett.Region.Name(), func() any { return kfi.GetRegion() }, func(v any) bool { return kfi.SetRegion(v.(int)) }, sett.Region.Value()),
		newConfigUpdater(sett.AdminName.Name(), func() any { return kfi.GetAdminName() }, func(v any) bool { return kfi.SetAdminName(v.(string)) }, sett.AdminName.Value()),
		newConfigUpdater(sett.AdminMail.Name(), func() any { return kfi.GetAdminMail() }, func(v any) bool { return kfi.SetAdminMail(v.(string)) }, sett.AdminMail.Value()),
		newConfigUpdater(sett.AdminPassword.Name(), func() any { return kfi.GetAdminPassword() }, func(v any) bool { return kfi.SetAdminPassword(v.(string)) }, sett.AdminPassword.Value()),
		newConfigUpdater(sett.MOTD.Value(), func() any { return kfi.GetMOTD() }, func(v any) bool { return kfi.SetMOTD(v.(string)) }, sett.MOTD.Value()),
		newConfigUpdater(sett.SpecimenType.Name(), func() any { return kfi.GetSpecimenType() }, func(v any) bool { return kfi.SetSpecimenType(v.(string)) }, sett.SpecimenType.Value()),
		newConfigUpdater(sett.RedirectURL.Name(), func() any { return kfi.GetRedirectURL() }, func(v any) bool { return kfi.SetRedirectURL(v.(string)) }, sett.RedirectURL.Value()),
		newConfigUpdater(sett.EnableWebAdmin.Name(), func() any { return kfi.IsWebAdminEnabled() }, func(v any) bool { return kfi.SetWebAdminEnabled(v.(bool)) }, sett.EnableWebAdmin.Value()),
		newConfigUpdater(sett.EnableMapVote.Name(), func() any { return kfi.IsMapVoteEnabled() }, func(v any) bool { return kfi.SetMapVoteEnabled(v.(bool)) == nil }, sett.EnableMapVote.Value()),
		newConfigUpdater(sett.MapVoteRepeatLimit.Name(), func() any { return kfi.GetMapVoteRepeatLimit() }, func(v any) bool { return kfi.SetMapVoteRepeatLimit(v.(int)) }, sett.MapVoteRepeatLimit.Value()),
	}
	for _, conf := range cuList {
		currentValue := conf.gv()
		if currentValue != conf.nv && !conf.sv(conf.nv) {
			return fmt.Errorf("[%s]: failed to set the new value: %v", conf.name, conf.nv)
		}
	}

	// Special cases
	clientRate := settings.DefaultMaxInternetClientRate
	if sett.Uncap.Value() {
		clientRate = 15000
	}
	if kfi.GetMaxInternetClientRate() != clientRate &&
		!kfi.SetMaxInternetClientRate(clientRate) {
		return fmt.Errorf("[Max Internet Client Rate]: failed to set the new value: %d", clientRate)
	}

	if err := updateConfigFileServerMutators(kfi, sett); err != nil {
		return fmt.Errorf("[Server Mutators]: %v", err)
	}

	if err := updateConfigFileMaplist(kfi, sett); err != nil {
		return fmt.Errorf("[Maplist]: %v", err)
	}
	return kfi.Save(kfiFilePath)
}

func updateKFPatcherConfigFile(sett *settings.KFDSLSettings) error {
	kfpiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KFPatcherSettings.ini")
	kfpi, err := config.NewKFPIniFile(kfpiFilePath)
	if err != nil {
		return err
	}

	cuList := []configUpdater[any]{
		newConfigUpdater(sett.KFPDisableZedTime.Name(), func() any { return kfpi.IsZEDTimeEnabled() }, func(v any) bool { return kfpi.SetZEDTimeEnabled(v.(bool)) }, !sett.KFPDisableZedTime.Value()),
		newConfigUpdater(sett.KFPEnableAllTraders.Name(), func() any { return kfpi.IsAllTradersOpenEnabled() }, func(v any) bool { return kfpi.SetZEDTimeEnabled(v.(bool)) }, sett.KFPEnableAllTraders.Value()),
		newConfigUpdater(sett.KFPAllTradersMessage.Name(), func() any { return kfpi.GetAllTradersMessage() }, func(v any) bool { return kfpi.SetAllTradersMessage(v.(string)) }, sett.KFPAllTradersMessage.Value()),
		newConfigUpdater(sett.KFPBuyEverywhere.Name(), func() any { return kfpi.IsBuyEverywhereEnabled() }, func(v any) bool { return kfpi.SetBuyEverywhereEnabled(v.(bool)) }, sett.KFPBuyEverywhere.Value()),
	}
	for _, conf := range cuList {
		currentValue := conf.gv()
		if currentValue != conf.nv && !conf.sv(conf.nv) {
			return fmt.Errorf("[%s]: failed to set the new value: %v", conf.name, conf.nv)
		}
	}
	return kfpi.Save(kfpiFilePath)
}

func readSteamCredentials(sett *settings.KFDSLSettings) error {
	defer func() {
		_ = os.Unsetenv("STEAMACC_USERNAME")
		_ = os.Unsetenv("STEAMACC_PASSWORD")
	}()

	// Read from Docker Secrets
	secrets, err := secrets.Read("kfds")
	if err == nil {
		username, ok := secrets["STEAMACC_USERNAME"]
		if !ok {
			return fmt.Errorf("steam account username (STEAMACC_USERNAME) not found in Docker secret")
		}
		password, ok := secrets["STEAMACC_PASSWORD"]
		if !ok {
			return fmt.Errorf("steam account password (STEAMACC_PASSWORD) not found in Docker secret")
		}
		sett.SteamLogin = username
		sett.SteamPassword = password
	} else {
		// Read from environment variables
		sett.SteamLogin = viper.GetString("STEAMACC_USERNAME")
		sett.SteamPassword = viper.GetString("STEAMACC_PASSWORD")
	}

	if sett.SteamLogin == "" || sett.SteamPassword == "" {
		return fmt.Errorf("incomplete credentials: both STEAMACC_USERNAME and STEAMACC_PASSWORD must be provided")
	}
	return nil
}
