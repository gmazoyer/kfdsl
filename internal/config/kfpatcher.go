package config

import (
	"github.com/K4rian/kfdsl/internal/config/ini"
	"github.com/K4rian/kfdsl/internal/settings"
)

type KFPIniFile struct {
	*ini.GenericIniFile
	filePath string
}

const (
	// Sections
	kfpRootSection = "KFPatcher.Settings"

	// Keys
	kfpKeyShowPerk          = "bShowPerk"
	kfpKeyAllowZedTime      = "bAllowZedTime"
	kfpKeyAllTradersOpen    = "bAllTradersOpen"
	kfpKeyAllTradersMessage = "bAllTradersMessage"
	kfpKeyBuyEverywhere     = "bBuyEverywhere"
)

func NewKFPIniFile(filePath string) (*KFPIniFile, error) {
	kfpIniFile := &KFPIniFile{
		GenericIniFile: ini.NewGenericIniFile("KFPIniFile"),
		filePath:       filePath,
	}
	if err := kfpIniFile.Load(filePath); err != nil {
		return nil, err
	}
	return kfpIniFile, nil
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
	return kf.GetKeyBool(kfpRootSection, kfpKeyShowPerk, !settings.DefaultKFPHidePerks)
}

func (kf *KFPIniFile) IsZEDTimeEnabled() bool {
	return kf.GetKeyBool(kfpRootSection, kfpKeyAllowZedTime, !settings.DefaultKFPDisableZedTime)
}

func (kf *KFPIniFile) IsAllTradersOpenEnabled() bool {
	return kf.GetKeyBool(kfpRootSection, kfpKeyAllTradersOpen, settings.DefaultKFPEnableAllTraders)
}

func (kf *KFPIniFile) GetAllTradersMessage() string {
	return kf.GetKey(kfpRootSection, kfpKeyAllTradersMessage, settings.DefaultKFPAllTradersMessage)
}

func (kf *KFPIniFile) IsBuyEverywhereEnabled() bool {
	return kf.GetKeyBool(kfpRootSection, kfpKeyBuyEverywhere, settings.DefaultKFPBuyEverywhere)
}

func (kf *KFPIniFile) SetShowPerksEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfpRootSection, kfpKeyShowPerk, enabled, true)
}

func (kf *KFPIniFile) SetZEDTimeEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfpRootSection, kfpKeyAllowZedTime, enabled, true)
}

func (kf *KFPIniFile) SetAllTradersOpenEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfpRootSection, kfpKeyAllTradersOpen, enabled, true)
}

func (kf *KFPIniFile) SetAllTradersMessage(message string) bool {
	return kf.SetKey(kfpRootSection, kfpKeyAllTradersMessage, message, true)
}

func (kf *KFPIniFile) SetBuyEverywhereEnabled(enabled bool) bool {
	return kf.SetKeyBool(kfpRootSection, kfpKeyBuyEverywhere, enabled, true)
}
