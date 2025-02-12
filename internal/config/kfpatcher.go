package config

import (
	"github.com/K4rian/kfdsl/internal/config/ini"
	"github.com/K4rian/kfdsl/internal/settings"
)

type KFPIniFile struct {
	*ini.GenericIniFile
	FilePath string
}

func NewKFPIniFile(filePath string) (*KFPIniFile, error) {
	KFPIniFile := &KFPIniFile{
		GenericIniFile: ini.NewGenericIniFile(),
		FilePath:       filePath,
	}
	if err := KFPIniFile.Load(filePath); err != nil {
		return nil, err
	}
	return KFPIniFile, nil
}

// sAlive

// sDead

// sSpectator

// sReady

// sNotReady

// sAwaiting

// sTagHP

// sTagKills

// fRefreshTime

func (kf *KFPIniFile) IsShowPerksEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bShowPerk", !settings.DefaultKFPHidePerks)
}

func (kf *KFPIniFile) IsZEDTimeEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bAllowZedTime", !settings.DefaultKFPDisableZedTime)
}

func (kf *KFPIniFile) IsAllTradersOpenEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bAllTradersOpen", settings.DefaultKFPEnableAllTraders)
}

func (kf *KFPIniFile) GetAllTradersMessage() string {
	return kf.GetKey("KFPatcher.Settings", "bAllTradersMessage", settings.DefaultKFPAllTradersMessage)
}

func (kf *KFPIniFile) IsBuyEverywhereEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bBuyEverywhere", settings.DefaultKFPBuyEverywhere)
}

// -----------------------

func (kf *KFPIniFile) SetShowPerksEnabled(enabled bool) bool {
	return kf.SetKeyBool("KFPatcher.Settings", "bShowPerk", enabled, true)
}

func (kf *KFPIniFile) SetZEDTimeEnabled(enabled bool) bool {
	return kf.SetKeyBool("KFPatcher.Settings", "bAllowZedTime", enabled, true)
}

func (kf *KFPIniFile) SetAllTradersOpenEnabled(enabled bool) bool {
	return kf.SetKeyBool("KFPatcher.Settings", "bAllTradersOpen", enabled, true)
}

func (kf *KFPIniFile) SetAllTradersMessage(message string) bool {
	return kf.SetKey("KFPatcher.Settings", "bAllTradersMessage", message, true)
}

func (kf *KFPIniFile) SetBuyEverywhereEnabled(enabled bool) bool {
	return kf.SetKeyBool("KFPatcher.Settings", "bBuyEverywhere", false, true)
}
