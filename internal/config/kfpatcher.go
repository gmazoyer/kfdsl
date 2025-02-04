package config

import (
	"kfdsl/internal/config/ini"
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

// bShowPerk

func (kf *KFPIniFile) IsZEDTimeEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bAllowZedTime", true)
}

func (kf *KFPIniFile) IsAllTradersOpenEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bAllTradersOpen", false)
}

func (kf *KFPIniFile) GetAllTradersMessage() string {
	return kf.GetKey("KFPatcher.Settings", "bAllTradersMessage", "")
}

func (kf *KFPIniFile) IsBuyEverywhereEnabled() bool {
	return kf.GetKeyBool("KFPatcher.Settings", "bBuyEverywhere", false)
}

// -----------------------

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
