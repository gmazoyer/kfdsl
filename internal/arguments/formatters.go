package arguments

import (
	"fmt"
	"math"
)

func FormatBool[T bool](a *Argument[bool]) string {
	if a.Value() {
		return "Enabled"
	}
	return "Disabled"
}

func FormatGameMode[T string](a *Argument[string]) string {
	val := a.Value()
	modes := map[string]string{
		"KFmod.KFGameType":            "Survival",
		"KFStoryGame.KFstoryGameInfo": "Objective",
		"KFCharPuppets.TOYGameInfo":   "Toy Master",
	}

	if _, ok := modes[val]; !ok {
		return "Custom"
	}
	return modes[val]
}

func FormatGameDifficulty[T int](a *Argument[int]) string {
	diff := map[int]string{
		1: "Easy",
		2: "Normal",
		4: "Hard",
		5: "Suicidal",
		7: "Hell on Earth",
	}
	return diff[a.Value()]
}

func FormatGameLength[T int](a *Argument[int]) string {
	lengths := map[int]string{
		0: "Short",
		1: "Medium",
		2: "Long",
	}
	return lengths[a.Value()]
}

func FormatFriendlyFireRate[T float64](a *Argument[float64]) string {
	return fmt.Sprintf("%.0f%%", math.Round(a.Value()*100))
}

func FormatSpecimenType[T string](a *Argument[string]) string {
	specimenTypes := map[string]string{
		"ET_None":             "Default",
		"ET_SummerSideshow":   "Summer (Summer Sideshow)",
		"ET_HillbillyHorror":  "Halloween (Hillbilly Horror)",
		"ET_TwistedChristmas": "Chrismas (Twisted Christmas)",
	}
	return specimenTypes[a.Value()]
}
