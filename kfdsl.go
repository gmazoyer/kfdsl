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

	"github.com/K4rian/kfdsl/cmd"
	"github.com/K4rian/kfdsl/internal/config"
	"github.com/K4rian/kfdsl/internal/config/secrets"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/services/redirectserver"
	"github.com/K4rian/kfdsl/internal/services/steamcmd"
)

const (
	KF_APPID = 215360
)

func startSteamCMD(ctx context.Context) error {
	steamCMD := steamcmd.NewSteamCMD(viper.GetString("STEAMCMD_ROOT"), ctx)

	if !steamCMD.IsPresent() {
		return fmt.Errorf("unable to locate SteamCMD in '%s', please install it first", steamCMD.RootDirectory())
	}

	// Read Steam Account Credentials
	if err := readSteamCredentials(); err != nil {
		return err
	}

	settings := cmd.GetSettings()

	// Generate the Steam install script
	installScript := path.Join(viper.GetString("STEAMCMD_ROOT"), "kfds_install_script.txt")
	serverInstallDir := viper.GetString("STEAMCMD_APPINSTALLDIR")

	if err := steamCMD.WriteScript(
		installScript,
		settings.SteamLogin,
		settings.SteamPassword,
		serverInstallDir,
		KF_APPID,
		!settings.NoValidate.Value(),
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

func startGameServer(ctx context.Context) (*kfserver.KFServer, error) {
	gameServer := kfserver.NewKFServer(viper.GetString("STEAMCMD_APPINSTALLDIR"), ctx)

	if !gameServer.IsPresent() {
		return nil, fmt.Errorf("unable to locate the Killing Floor Dedicated Server files in '%s', please install using SteamCMD", gameServer.RootDirectory())
	}

	fmt.Printf("> Updating the Killing Floor configuration file...\n")
	if err := updateConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to update the KillingFloor.ini configuration file: %v", err)
	}

	settings := cmd.GetSettings()
	if settings.EnableKFPatcher.Value() {
		fmt.Printf("> Updating the KFPatcher configuration file...\n")
		if err := updateKFPatcherConfigFile(); err != nil {
			return nil, fmt.Errorf("failed to update the KFPatcherSettings.ini configuration file: %v", err)
		}
	}

	fmt.Printf("> Starting the Killing Floor Server...\n")
	if err := gameServer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start the Killing Floor Dedicated Server: %v", err)
	}
	return gameServer, nil
}

func startRedirectServer(ctx context.Context) (*redirectserver.HTTPRedirectServer, error) {
	settings := cmd.GetSettings()
	redirectServer := redirectserver.NewHTTPRedirectServer(
		settings.RedirectServerHost.Value(),
		settings.RedirectServerPort.Value(),
		settings.RedirectServerDir.Value(),
		settings.RedirectServerMaxRequests.Value(),
		settings.RedirectServerBanTime.Value(),
		ctx,
	)

	fmt.Printf("> Starting the HTTP Redirect Server...\n")
	if err := redirectServer.Listen(); err != nil {
		return nil, fmt.Errorf("failed to start the HTTP Redirect Server: %v", err)
	}
	return redirectServer, nil
}

func updateConfigFileServerMutators(iniFile *config.KFIniFile, settings *cmd.KFDSLSettings) error {
	mutatorsStr := settings.ServerMutators.Value()
	mutators := strings.FieldsFunc(mutatorsStr, func(r rune) bool { return r == ',' })

	if settings.EnableKFPatcher.Value() && !strings.Contains(mutatorsStr, "KFPatcher") {
		mutators = append(mutators, "KFPatcher.Mut")
	}

	if len(mutators) > 0 {
		if err := iniFile.SetServerMutators(mutators); err != nil {
			return fmt.Errorf("[ServerMutators]: %v", err)
		}
	} else {
		if err := iniFile.ClearServerMutators(); err != nil {
			return fmt.Errorf("[ServerMutators]: %v", err)
		}
	}
	return nil
}

func updateConfigFileMaplist(iniFile *config.KFIniFile, settings *cmd.KFDSLSettings) error {
	gameMode := settings.GameMode.RawValue()

	sectionName := kfserver.GetGameModeMaplistName(gameMode)
	if sectionName == "" {
		return fmt.Errorf("[Maplist]: undefined section name for game mode: %s", gameMode)
	}

	mapList := strings.FieldsFunc(settings.Maplist.Value(), func(r rune) bool { return r == ',' })

	if len(mapList) > 0 {
		if mapList[0] == "all" {
			gameServerRoot := viper.GetString("STEAMCMD_APPINSTALLDIR")
			gameModePrefix := kfserver.GetGameModeMapPrefix(gameMode)
			installedMaps, err := kfserver.GetInstalledMaps(
				path.Join(gameServerRoot, "Maps"),
				gameModePrefix,
			)
			if err != nil {
				return fmt.Errorf("[Maplist]: unable to fetch available maps for game mode '%s': %v", gameMode, err)
			}
			mapList = installedMaps
		}

		if err := iniFile.SetMaplist(sectionName, mapList); err != nil {
			return fmt.Errorf("[Maplist]: %v", err)
		}
	} else {
		if err := iniFile.ClearMaplist(sectionName); err != nil {
			return fmt.Errorf("[Maplist]: %v", err)
		}
	}
	return nil
}

func updateConfigFile() error {
	// TODO: Check returned values
	kfiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KillingFloor.ini")
	kfi, err := config.NewKFIniFile(kfiFilePath)
	if err != nil {
		return err
	}

	settings := cmd.GetSettings()

	if kfi.GetServerName() != settings.ServerName.Value() {
		kfi.SetServerName(settings.ServerName.Value())
	}

	if kfi.GetShortName() != settings.ShortName.Value() {
		kfi.SetShortName(settings.ShortName.Value())
	}

	if kfi.GetGamePort() != settings.GamePort.Value() {
		kfi.SetGamePort(settings.GamePort.Value())
	}

	if kfi.GetWebAdminPort() != settings.WebAdminPort.Value() {
		kfi.SetWebAdminPort(settings.WebAdminPort.Value())
	}

	if kfi.GetGameSpyPort() != settings.GameSpyPort.Value() {
		kfi.SetGameSpyPort(settings.GameSpyPort.Value())
	}

	if kfi.GetGameDifficulty() != settings.GameDifficulty.Value() {
		kfi.SetGameDifficulty(settings.GameDifficulty.Value())
	}

	if kfi.GetGameLength() != settings.GameLength.Value() {
		kfi.SetGameLength(settings.GameLength.Value())
	}

	if kfi.GetFriendlyFireRate() != settings.FriendlyFire.Value() {
		kfi.SetFriendlyFireRate(settings.FriendlyFire.Value())
	}

	if kfi.GetMaxPlayers() != settings.MaxPlayers.Value() {
		kfi.SetMaxPlayers(settings.MaxPlayers.Value())
	}

	if kfi.GetMaxSpectators() != settings.MaxSpectators.Value() {
		kfi.SetMaxSpectators(settings.MaxSpectators.Value())
	}

	if kfi.GetPassword() != settings.Password.Value() {
		kfi.SetPassword(settings.Password.Value())
	}

	if kfi.GetRegion() != settings.Region.Value() {
		kfi.SetRegion(settings.Region.Value())
	}

	if kfi.GetAdminName() != settings.AdminName.Value() {
		kfi.SetAdminName(settings.AdminName.Value())
	}

	if kfi.GetAdminMail() != settings.AdminMail.Value() {
		kfi.SetAdminMail(settings.AdminMail.Value())
	}

	if kfi.GetAdminPassword() != settings.AdminPassword.Value() {
		kfi.SetAdminPassword(settings.AdminPassword.Value())
	}

	if kfi.GetMOTD() != settings.MOTD.Value() {
		kfi.SetMOTD(settings.MOTD.Value())
	}

	if kfi.GetSpecimenType() != settings.SpecimenType.Value() {
		kfi.SetSpecimenType(settings.SpecimenType.Value())
	}

	if kfi.GetRedirectURL() != settings.RedirectURL.Value() {
		kfi.SetRedirectURL(settings.RedirectURL.Value())
	}

	if kfi.IsWebAdminEnabled() != settings.EnableWebAdmin.Value() {
		kfi.SetWebAdminEnabled(settings.EnableWebAdmin.Value())
	}

	if kfi.IsMapVoteEnabled() != settings.EnableMapVote.Value() {
		kfi.SetMapVoteEnabled(settings.EnableMapVote.Value())
	}

	if kfi.GetMapVoteRepeatLimit() != settings.MapVoteRepeatLimit.Value() {
		kfi.SetMapVoteRepeatLimit(settings.MapVoteRepeatLimit.Value())
	}

	if kfi.IsAdminPauseEnabled() != settings.EnableAdminPause.Value() {
		kfi.SetAdminPauseEnabled(settings.EnableAdminPause.Value())
	}

	if kfi.IsWeaponThrowingEnabled() != !settings.DisableWeaponThrow.Value() {
		kfi.SetWeaponThrowingEnabled(!settings.DisableWeaponThrow.Value())
	}

	if kfi.IsWeaponShakeEffectEnabled() != !settings.DisableWeaponShake.Value() {
		kfi.SetWeaponShakeEffectEnabled(!settings.DisableWeaponShake.Value())
	}

	if kfi.IsThirdPersonEnabled() != settings.EnableThirdPerson.Value() {
		kfi.SetThirdPersonEnabled(settings.EnableThirdPerson.Value())
	}

	if kfi.IsLowGoreEnabled() != settings.EnableLowGore.Value() {
		kfi.SetLowGoreEnabled(settings.EnableLowGore.Value())
	}

	clientRate := 10000 // Default
	if settings.Uncap.Value() {
		clientRate = 15000
	}
	if kfi.GetMaxInternetClientRate() != clientRate {
		kfi.SetMaxInternetClientRate(clientRate)
	}

	if err := updateConfigFileServerMutators(kfi, settings); err != nil {
		return err
	}

	if err := updateConfigFileMaplist(kfi, settings); err != nil {
		return err
	}
	return kfi.Save(kfiFilePath)
}

func updateKFPatcherConfigFile() error {
	// TODO: Check returned values
	kfpiFilePath := filepath.Join(viper.GetString("STEAMCMD_APPINSTALLDIR"), "System", "KFPatcherSettings.ini")
	kfpi, err := config.NewKFPIniFile(kfpiFilePath)
	if err != nil {
		return err
	}

	settings := cmd.GetSettings()

	if kfpi.IsZEDTimeEnabled() != !settings.KFPDisableZedTime.Value() {
		kfpi.SetZEDTimeEnabled(!settings.KFPDisableZedTime.Value())
	}

	if kfpi.IsAllTradersOpenEnabled() != settings.KFPEnableAllTraders.Value() {
		kfpi.SetAllTradersOpenEnabled(settings.KFPEnableAllTraders.Value())
	}

	if kfpi.GetAllTradersMessage() != settings.KFPAllTradersMessage.Value() {
		kfpi.SetAllTradersMessage(settings.KFPAllTradersMessage.Value())
	}

	if kfpi.IsBuyEverywhereEnabled() != settings.KFPBuyEverywhere.Value() {
		kfpi.SetBuyEverywhereEnabled(settings.KFPBuyEverywhere.Value())
	}
	return kfpi.Save(kfpiFilePath)
}

func readSteamCredentials() error {
	settings := cmd.GetSettings()

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
		settings.SteamLogin = username
		settings.SteamPassword = password
	} else {
		// Read from environment variables
		settings.SteamLogin = viper.GetString("STEAMACC_USERNAME")
		settings.SteamPassword = viper.GetString("STEAMACC_PASSWORD")
	}

	if settings.SteamLogin == "" || settings.SteamPassword == "" {
		return fmt.Errorf("incomplete credentials: both STEAMACC_USERNAME and STEAMACC_PASSWORD must be provided")
	}

	os.Unsetenv("STEAMACC_USERNAME")
	os.Unsetenv("STEAMACC_PASSWORD")
	return nil
}
