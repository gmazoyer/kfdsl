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
	"github.com/K4rian/kfdsl/internal/log"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/services/steamcmd"
	"github.com/K4rian/kfdsl/internal/settings"
	"github.com/K4rian/kfdsl/internal/utils"
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

	log.Logger.Info("Writing the KF Dedicated Server install script...", "scriptPath", installScript)
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
	log.Logger.Info("KF Dedicated Server install script was successfully written", "scriptPath", installScript)

	log.Logger.Info("Starting SteamCMD...", "rootDir", steamCMD.RootDirectory(), "appInstallRootDir", serverInstallDir)
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
	mutators := sett.Mutators.Value()
	if sett.EnableMutLoader.Value() {
		mutators = "MutLoader.MutLoader"
	}

	rootDir := viper.GetString("STEAMCMD_APPINSTALLDIR")
	extraArgs := viper.GetStringSlice("KF_EXTRAARGS")

	gameServer := kfserver.NewKFServer(
		rootDir,
		sett.StartupMap.Value(),
		sett.GameMode.Value(),
		sett.Unsecure.Value(),
		sett.MaxPlayers.Value(),
		mutators,
		extraArgs,
		ctx,
	)

	if !gameServer.IsPresent() {
		return nil, fmt.Errorf("unable to locate the KF Dedicated Server files in '%s', please install using SteamCMD", gameServer.RootDirectory())
	}

	configFilePath := sett.ConfigFile.Value()
	log.Logger.Info("Updating the KF Dedicated Server configuration file...", "configFile", configFilePath)
	if err := updateConfigFile(sett); err != nil {
		return nil, fmt.Errorf("failed to update the KF Dedicated Server configuration file '%s': %v", configFilePath, err)
	}
	log.Logger.Info("KF Dedicated Server configuration file successfully updated", "configFile", configFilePath)

	if sett.EnableKFPatcher.Value() {
		kfpConfigFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KFPatcherSettings.ini")
		log.Logger.Info("Updating the KFPatcher configuration file...", "configFile", kfpConfigFilePath)
		if err := updateKFPatcherConfigFile(sett); err != nil {
			return nil, fmt.Errorf("failed to update the KFPatcher configuration file '%s': %v", kfpConfigFilePath, err)
		}
		log.Logger.Info("KFPatcher configuration file successfully updated", "configFile", kfpConfigFilePath)
	}

	log.Logger.Info("Verifying KF Dedicated Server Steam libraries for updates...")
	updatedLibs, err := updateGameServerSteamLibs()
	if err == nil {
		if len(updatedLibs) > 0 {
			for _, lib := range updatedLibs {
				log.Logger.Info("Steam library successfully updated", "library", lib)
			}
		} else {
			log.Logger.Info("All KF Dedicated Server Steam libraries are up-to-date")
		}
	} else {
		log.Logger.Error("Unable to update the KF Dedicated Server Steam libraries", "err", err)
	}

	log.Logger.Info("Starting the KF Dedicated Server...", "rootDir", gameServer.RootDirectory(), "autoRestart", sett.AutoRestart.Value())
	if err := gameServer.Start(sett.AutoRestart.Value()); err != nil {
		return nil, fmt.Errorf("failed to start the KF Dedicated Server: %v", err)
	}
	return gameServer, nil
}

func createConfigFile(filePath string) error {
	defaultIniFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "Default.ini")
	if !utils.FileExists(defaultIniFilePath) {
		return fmt.Errorf("default configuration file '%s' not found", defaultIniFilePath)
	}

	if err := utils.CopyFile(defaultIniFilePath, filePath); err != nil {
		return fmt.Errorf("failed to copy '%s' file to '%s'", defaultIniFilePath, filePath)
	}
	return nil
}

func updateConfigFile(sett *settings.KFDSLSettings) error {
	kfiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", sett.ConfigFile.Value())

	// Duplicate the 'Default.ini' config file
	if !utils.FileExists(kfiFilePath) {
		if err := createConfigFile(kfiFilePath); err != nil {
			return err
		}
	}

	// Read the ini file
	kfi, err := config.NewKFIniFile(kfiFilePath)
	if err != nil {
		return err
	}

	// Generics
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

func updateKFPatcherConfigFile(sett *settings.KFDSLSettings) error {
	kfpiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KFPatcherSettings.ini")
	kfpi, err := config.NewKFPIniFile(kfpiFilePath)
	if err != nil {
		return err
	}

	cuList := []configUpdater[any]{
		newConfigUpdater(sett.KFPHidePerks.Name(), func() any { return kfpi.IsShowPerksEnabled() }, func(v any) bool { return kfpi.SetShowPerksEnabled(v.(bool)) }, !sett.KFPHidePerks.Value()),
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

func updateGameServerSteamLibs() ([]string, error) {
	ret := []string{}
	rootDir := viper.GetString("STEAMCMD_APPINSTALLDIR")
	systemDir := path.Join(rootDir, "System")
	libsDir := path.Join(viper.GetString("STEAMCMD_ROOT"), "linux32")

	libs := map[string]string{
		path.Join(libsDir, "steamclient.so"):  path.Join(systemDir, "steamclient.so"),
		path.Join(libsDir, "libtier0_s.so"):   path.Join(systemDir, "libtier0_s.so"),
		path.Join(libsDir, "libvstdlib_s.so"): path.Join(systemDir, "libvstdlib_s.so"),
	}

	for srcFile, dstFile := range libs {
		identical, err := utils.SHA1Compare(srcFile, dstFile)
		if err != nil {
			return ret, fmt.Errorf("error comparing files %s and %s: %v", srcFile, dstFile, err)
		}

		if !identical {
			if err := utils.CopyAndReplaceFile(srcFile, dstFile); err != nil {
				return ret, err
			}
			ret = append(ret, dstFile)
		}
	}
	return ret, nil
}

func readSteamCredentials(sett *settings.KFDSLSettings) error {
	var fromEnv bool

	defer func() {
		if fromEnv {
			_ = os.Unsetenv("STEAMACC_USERNAME")
			_ = os.Unsetenv("STEAMACC_PASSWORD")
		}
	}()

	// Try reading from Docker Secrets
	steamUsername, errUser := secrets.Read("STEAMACC_USERNAME")
	steamPassword, errPass := secrets.Read("STEAMACC_PASSWORD")

	// Fallback to environment variables if secrets are missing
	if errUser != nil {
		steamUsername = viper.GetString("STEAMACC_USERNAME")
		fromEnv = true
	}
	if errPass != nil {
		steamPassword = viper.GetString("STEAMACC_PASSWORD")
		fromEnv = true
	}

	// Ensure both credentials are present
	if steamUsername == "" || steamPassword == "" {
		return fmt.Errorf("incomplete credentials: both STEAMACC_USERNAME and STEAMACC_PASSWORD must be provided")
	}

	// Update the settings
	sett.SteamLogin = steamUsername
	sett.SteamPassword = steamPassword
	return nil
}
