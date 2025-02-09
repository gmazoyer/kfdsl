package kfserver

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetInstalledMaps(dir, prefix string) ([]string, error) {
	var filteredFiles []string

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := file.Name()
		if !file.IsDir() && strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, ".rom") {
			filteredFiles = append(filteredFiles, strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		}
	}
	return filteredFiles, nil
}

func GetGameModeMapPrefix(gamemode string) string {
	modes := map[string]string{
		"survival":  "KF-",
		"objective": "KFO-",
		"toymaster": "TOY-",
	}
	return modes[strings.ToLower(gamemode)]
}

func GetGameModeMaplistName(gamemode string) string {
	mlist := map[string]string{
		"survival":  "KFMod.KFMaplist",
		"objective": "KFStoryGame.KFOMapList",
		"toymaster": "KFCharPuppets.TOYMapList",
	}
	return mlist[strings.ToLower(gamemode)]
}

func GetSeasonalSpecimenType() string {
	currentMonth := time.Now().Month()

	switch currentMonth {
	case time.June, time.July, time.August:
		return "ET_SummerSideshow"
	case time.October:
		return "ET_HillbillyHorror"
	case time.December:
		return "ET_TwistedChristmas"
	default:
		return "ET_None"
	}
}
