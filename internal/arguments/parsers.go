package arguments

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strings"
)

func ParseNonEmptyStr(raw string, name string) func(r string) (string, error) {
	return func(r string) (string, error) {
		val := strings.TrimSpace(strings.ToLower(raw))
		if val == "" {
			return "", fmt.Errorf("invalid %s: undefined or empty", name)
		}
		return raw, nil
	}
}

func ParseUnsignedInt(raw int, name string) func(r int) (int, error) {
	return func(r int) (int, error) {
		if raw < 0 {
			return 0, fmt.Errorf("invalid %s (%d): value cannot be negative", name, raw)
		}
		return raw, nil
	}
}

// func ParsePositiveInt(raw int, name string) func(r int) (int, error) {
// 	return func(r int) (int, error) {
// 		if raw < 1 {
// 			return 0, fmt.Errorf("invalid %s (%d): value must be positive", name, raw)
// 		}
// 		return raw, nil
// 	}
// }

func ParseIntRange(raw int, min int, max int, name string) func(r int) (int, error) {
	return func(r int) (int, error) {
		if raw < min || raw > max {
			return 0, fmt.Errorf("invalid %s (%d): value must be between %d-%d", name, raw, min, max)
		}
		return raw, nil
	}
}

// -----------------------

func ParsePort(raw int, service string) func(r int) (int, error) {
	return func(r int) (int, error) {
		if raw < 1024 && raw > 65535 {
			return 0, fmt.Errorf("invalid %s port (%d): value must be in range 1025-65534", service, raw)
		}
		return raw, nil
	}
}

func ParsePassword(raw string, name string) func(r string) (string, error) {
	return func(r string) (string, error) {
		val := strings.TrimSpace(strings.ToLower(raw))
		if val != "" && len(val) > 16 {
			return "", fmt.Errorf("invalid %s (%s): value cannot exceed 16 characters", name, raw)
		}
		return strings.TrimSpace(raw), nil
	}
}

func ParseURL(raw string) (string, error) {
	val := strings.TrimSpace(raw)
	if val != "" {
		parsedURL, err := url.Parse(val)
		if err != nil || !strings.HasPrefix(parsedURL.Scheme, "http") {
			return "", fmt.Errorf("invalid URL '%s'", raw)
		}
	}
	return val, nil
}

func ParseIP(ip string) (string, error) {
	val := strings.TrimSpace(ip)
	if val == "" {
		return "", fmt.Errorf("IP address is empty")
	}

	parsedIP := net.ParseIP(val)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP address: '%s'", ip)
	}
	return val, nil
}

func ParseExistingDir(raw string) (string, error) {
	val := strings.TrimSpace(raw)
	if val != "" {
		info, err := os.Stat(val)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("directory does not exist: '%s'", raw)
			}
			return "", fmt.Errorf("error checking directory: '%s': %v", raw, err)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("path is not a directory: '%s'", raw)
		}
	}
	return val, nil
}

func ParseGameMode(raw string) (string, error) {
	validModes := map[string]string{
		"survival":  "KFmod.KFGameType",
		"objective": "KFStoryGame.KFstoryGameInfo",
		"toymaster": "KFCharPuppets.TOYGameInfo",
	}

	val, ok := validModes[strings.ToLower(raw)]
	if val == "" {
		val = validModes["survival"]
		ok = true
	}
	if ok {
		return val, nil
	}
	return raw, nil // Custom
}

func ParseGameDifficulty(raw string) (int, error) {
	difficulties := map[string]int{
		"easy":     1,
		"normal":   2,
		"hard":     4,
		"suicidal": 5,
		"hell":     7,
	}

	if val, ok := difficulties[strings.ToLower(raw)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("invalid Game Difficulty: %s", raw)
}

func ParseGameLength(raw string) (int, error) {
	lengths := map[string]int{
		"short":  0,
		"medium": 1,
		"long":   2,
	}

	if val, ok := lengths[strings.ToLower(raw)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("invalid Game Length: %s", raw)
}

func ParseFriendlyFireRate(raw float64) (float64, error) {
	if raw < 0.0 || raw > 1.0 {
		return 0, fmt.Errorf("invalid Friendly Fire Rate (%f): value must be between 0.0 and 1.0", raw)
	}
	return raw, nil
}

func ParseAdminMail(raw string) (string, error) {
	val := strings.TrimSpace(strings.ToLower(raw))
	_, err := mail.ParseAddress(val)
	if val != "" && err != nil {
		return "", fmt.Errorf("invalid Admin Email: %s", raw)
	}
	return raw, nil
}

func ParseSpecimentType(raw string) (string, error) {
	var specimenTypes = map[string]string{
		"default":   "ET_None",
		"summer":    "ET_SummerSideshow",
		"halloween": "ET_HillbillyHorror",
		"christmas": "ET_TwistedChristmas",
	}

	val, ok := specimenTypes[strings.ToLower(raw)]
	if val == "" {
		val = specimenTypes["default"]
		ok = true
	}
	if ok {
		return val, nil
	}
	return "", fmt.Errorf("invalid Specimen Type: %s", raw)
}
