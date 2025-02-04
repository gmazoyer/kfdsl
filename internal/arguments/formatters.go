package arguments

import (
	"fmt"
	"math"
)

func FormatBool(raw bool) string {
	if raw {
		return "Enabled"
	}
	return "Disabled"
}

func FormatGameMode(value string) string {
	modes := map[string]string{
		"KFmod.KFGameType":            "Survival",
		"KFStoryGame.KFstoryGameInfo": "Objective",
		"KFCharPuppets.TOYGameInfo":   "Toy Master",
	}

	if _, ok := modes[value]; !ok {
		return "Custom"
	}
	return modes[value]
}

func FormatGameDifficulty(value int) string {
	diff := map[int]string{
		1: "Easy",
		2: "Normal",
		4: "Hard",
		5: "Suicidal",
		7: "Hell on Earth",
	}
	return diff[value]
}

func FormatGameLength(value int) string {
	lengths := map[int]string{
		0: "Short",
		1: "Medium",
		2: "Long",
	}
	return lengths[value]
}

func FormatFriendlyFireRate(value float64) string {
	percent := math.Round(value * 100)
	return fmt.Sprintf("%.0f%%", percent)
}

func FormatSpecimentType(value string) string {
	var specimenTypes = map[string]string{
		"ET_None":             "Default",
		"ET_SummerSideshow":   "Summer (Summer Sideshow)",
		"ET_HillbillyHorror":  "Halloween (Hillbilly Horror)",
		"ET_TwistedChristmas": "Chrismas (Twisted Christmas)",
	}
	return specimenTypes[value]
}
