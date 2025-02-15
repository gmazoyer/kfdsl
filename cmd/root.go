package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/K4rian/kfdsl/internal/arguments"
	"github.com/K4rian/kfdsl/internal/settings"
)

func BuildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "./kfdsl",
		Short: "KF Dedicated Server Launcher",
		Long:  "A command-line tool to configure and run a Killing Floor dedicated server.",
		RunE:  runRootCommand,
	}

	var userHome, _ = os.UserHomeDir()

	var configFile, serverName, shortName, gameMode, startupMap, gameDifficulty, gameLength,
		password, adminName, adminMail, adminPassword, motd, specimenType, mutators,
		serverMutators, redirectURL, mapList, allTradersMessage, kfunflectURL, kfpatcherURL,
		logLevel, logFilePath, logFileFormat, steamRootDir, steamAppInstallDir string

	var gamePort, webadminPort, gamespyPort, maxPlayers, maxSpectators, region,
		mapVoteRepeatLimit, logMaxSize, logMaxBackups, logMaxAge int

	var friendlyFire float64

	var enableWebAdmin, enableMapVote, enableAdminPause, disableWeaponThrow,
		disableWeaponShake, enableThirdPerson, enableLowGore, uncap, unsecure, noSteam,
		disableValidation, enableAutoRestart, enableMutloader, enableKFPatcher, enableShowPerks,
		disableZEDTime, enableBuyEverywhere, enableAllTraders, enableFileLogging bool

	flags := map[string]struct {
		Value   interface{}
		Desc    string
		Default interface{}
	}{
		"config":                 {&configFile, "configuration file", settings.DefaultConfigFile},
		"servername":             {&serverName, "server name", settings.DefaultServerName},
		"shortname":              {&shortName, "server short name", settings.DefaultShortName},
		"port":                   {&gamePort, "game UDP port", settings.DefaultGamePort},
		"webadminport":           {&webadminPort, "WebAdmin TCP port", settings.DefaultWebAdminPort},
		"gamespyport":            {&gamespyPort, "GameSpy UDP port", settings.DefaultGameSpyPort},
		"gamemode":               {&gameMode, "game mode", settings.DefaultGameMode},
		"map":                    {&startupMap, "starting map", settings.DefaultStartupMap},
		"difficulty":             {&gameDifficulty, "game difficulty (easy, normal, hard, suicidal, hell)", settings.DefaultGameDifficulty},
		"length":                 {&gameLength, "game length (waves) (short, medium, long)", settings.DefaultGameLength},
		"friendlyfire":           {&friendlyFire, "friendly fire rate (0.0-1.0)", settings.DefaultFriendlyFire},
		"maxplayers":             {&maxPlayers, "maximum players", settings.DefaultMaxPlayers},
		"maxspectators":          {&maxSpectators, "maximum spectators", settings.DefaultMaxSpectators},
		"password":               {&password, "server password", settings.DefaultPassword},
		"region":                 {&region, "server region", settings.DefaultRegion},
		"adminname":              {&adminName, "server administrator name", settings.DefaultAdminName},
		"adminmail":              {&adminMail, "server administrator email", settings.DefaultAdminMail},
		"adminpassword":          {&adminPassword, "server administrator password", settings.DefaultAdminPassword},
		"motd":                   {&motd, "message of the day", settings.DefaultMOTD},
		"specimentype":           {&specimenType, "specimen type (default, summer, halloween, christmas)", settings.DefaultSpecimenType},
		"mutators":               {&mutators, "comma-separated mutators (command-line)", settings.DefaultMutators},
		"servermutators":         {&serverMutators, "comma-separated mutators (server actors)", settings.DefaultServerMutators},
		"redirecturl":            {&redirectURL, "redirect URL", settings.DefaultRedirectURL},
		"maplist":                {&mapList, "comma-separated maps for the current game mode. Use 'all' to append all available map", settings.DefaultMaplist},
		"webadmin":               {&enableWebAdmin, "enable WebAdmin panel", settings.DefaultEnableWebAdmin},
		"mapvote":                {&enableMapVote, "enable map voting", settings.DefaultEnableMapVote},
		"mapvote-repeatlimit":    {&mapVoteRepeatLimit, "number of maps to be played before a map can repeat", settings.DefaultMapVoteRepeatLimit},
		"adminpause":             {&enableAdminPause, "allow admin to pause game", settings.DefaultEnableAdminPause},
		"noweaponthrow":          {&disableWeaponThrow, "disable weapon throwing", settings.DefaultDisableWeaponThrow},
		"noweaponshake":          {&disableWeaponShake, "disable weapon-induced screen shake", settings.DefaultDisableWeaponShake},
		"thirdperson":            {&enableThirdPerson, "enable third-person view", settings.DefaultEnableThirdPerson},
		"lowgore":                {&enableLowGore, "reduce gore", settings.DefaultEnableLowGore},
		"uncap":                  {&uncap, "uncap the frame rate", settings.DefaultUncap},
		"unsecure":               {&unsecure, "disable VAC (Valve Anti-Cheat)", settings.DefaultUnsecure},
		"nosteam":                {&noSteam, "start the server without calling SteamCMD", settings.DefaultNoSteam},
		"novalidate":             {&disableValidation, "skip server files integrity check", settings.DefaultNoValidate},
		"autorestart":            {&enableAutoRestart, "restart server on crash", settings.DefaultAutoRestart},
		"mutloader":              {&enableMutloader, "enable MutLoader (override inline mutators)", settings.DefaultEnableMutLoader},
		"kfpatcher":              {&enableKFPatcher, "enable KFPatcher", settings.DefaultEnableKFPatcher},
		"hideperks":              {&enableShowPerks, "(KFPatcher) hide perks", settings.DefaultKFPHidePerks},
		"nozedtime":              {&disableZEDTime, "(KFPatcher) disable ZED time", settings.DefaultKFPDisableZedTime},
		"buyeverywhere":          {&enableBuyEverywhere, "(KFPatcher) allow players to shop whenever", settings.DefaultKFPBuyEverywhere},
		"alltraders":             {&enableAllTraders, "(KFPatcher) make all trader's spots accessible", settings.DefaultKFPEnableAllTraders},
		"alltraders-message":     {&allTradersMessage, "(KFPatcher) All traders screen message", settings.DefaultKFPAllTradersMessage},
		"kfunflect-url":          {&kfunflectURL, "(KFPatcher) KFUnflect URL", settings.DefaultKFUnflectURL},
		"kfpatcher-url":          {&kfpatcherURL, "(KFPatcher) archive URL", settings.DefaultKFPatcherURL},
		"log-to-file":            {&enableFileLogging, "enable file logging", settings.DefaultLogToFile},
		"log-level":              {&logLevel, "log level (info, debug, warn, error)", settings.DefaultLogLevel},
		"log-file":               {&logFilePath, "log file path", settings.DefaultLogFile},
		"log-file-format":        {&logFileFormat, "log format (text or json)", settings.DefaultLogFileFormat},
		"log-max-size":           {&logMaxSize, "max log file size (MB)", settings.DefaultLogMaxSize},
		"log-max-backups":        {&logMaxBackups, "max number of old log files to keep", settings.DefaultLogMaxBackups},
		"log-max-age":            {&logMaxAge, "max age of a log file (days)", settings.DefaultLogMaxAge},
		"steamcmd-root":          {&steamRootDir, "SteamCMD root directory", path.Join(userHome, "steamcmd")},
		"steamcmd-appinstalldir": {&steamAppInstallDir, "server installatation directory", path.Join(userHome, "gameserver")},
	}

	for flag, data := range flags {
		switch v := data.Default.(type) {
		case string:
			val := data.Value.(*string)
			rootCmd.Flags().StringVar(val, flag, v, data.Desc)
		case int:
			val := data.Value.(*int)
			rootCmd.Flags().IntVar(val, flag, v, data.Desc)
		case float64:
			val := data.Value.(*float64)
			rootCmd.Flags().Float64Var(val, flag, v, data.Desc)
		case bool:
			val := data.Value.(*bool)
			rootCmd.Flags().BoolVar(val, flag, v, data.Desc)
		}

		// SteamCMD-related configurations don't use the 'KF' prefix
		if strings.HasPrefix(flag, "steamcmd") {
			envName := strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
			viper.BindEnv(flag, envName)
		} else {
			viper.BindEnv(flag)
		}
		viper.BindPFlag(flag, rootCmd.Flags().Lookup(flag))
	}

	viper.BindEnv("STEAMACC_USERNAME")
	viper.BindEnv("STEAMACC_PASSWORD")
	viper.BindEnv("KF_EXTRAARGS")

	viper.SetDefault("STEAMACC_USERNAME", settings.DefaultSteamLogin)

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("KF")
	viper.AutomaticEnv()

	return rootCmd
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	sett := settings.Get()

	registerArguments(sett)

	if err := sett.Parse(); err != nil {
		return err
	}

	viper.SetDefault("KF_EXTRAARGS", args)
	sett.ExtraArgs = viper.GetStringSlice("KF_EXTRAARGS")
	return nil
}

func registerArguments(sett *settings.KFDSLSettings) {
	sett.ConfigFile = arguments.NewArgument[string](viper.GetString("config"), "Config File", nil, nil, false)
	sett.ServerName = arguments.NewArgument[string](viper.GetString("servername"), "Server Name", arguments.ParseNonEmptyStr(viper.GetString("servername"), "Server Name"), nil, false)
	sett.ShortName = arguments.NewArgument[string](viper.GetString("shortname"), "Short Name", arguments.ParseNonEmptyStr(viper.GetString("shortname"), "Short Name"), nil, false)
	sett.GamePort = arguments.NewArgument[int](viper.GetInt("port"), "Game Port", arguments.ParsePort(viper.GetInt("port"), "Game Port"), nil, false)
	sett.WebAdminPort = arguments.NewArgument[int](viper.GetInt("webadminport"), "WebAdmin Port", arguments.ParsePort(viper.GetInt("webadminport"), "WebAdmin Port"), nil, false)
	sett.GameSpyPort = arguments.NewArgument[int](viper.GetInt("gamespyport"), "GameSpy Port", arguments.ParsePort(viper.GetInt("gamespyport"), "GameSpy Port"), nil, false)
	sett.GameMode = arguments.NewArgument[string](viper.GetString("gamemode"), "Game Mode", arguments.ParseGameMode, arguments.FormatGameMode, false)
	sett.StartupMap = arguments.NewArgument[string](viper.GetString("map"), "Startup Map", arguments.ParseNonEmptyStr(viper.GetString("map"), "Startup Map"), nil, false)
	sett.GameDifficulty = arguments.NewArgument[int](viper.GetString("difficulty"), "Game Difficulty", arguments.ParseGameDifficulty, arguments.FormatGameDifficulty, false)
	sett.GameLength = arguments.NewArgument[int](viper.GetString("length"), "Game Length", arguments.ParseGameLength, arguments.FormatGameLength, false)
	sett.FriendlyFire = arguments.NewArgument[float64](viper.GetFloat64("friendlyfire"), "Friendly Fire Rate", arguments.ParseFriendlyFireRate, arguments.FormatFriendlyFireRate, false)
	sett.MaxPlayers = arguments.NewArgument[int](viper.GetInt("maxplayers"), "Max Players", arguments.ParseIntRange(viper.GetInt("maxplayers"), 0, 32, "Max Players"), nil, false)
	sett.MaxSpectators = arguments.NewArgument[int](viper.GetInt("maxspectators"), "Max Spectators", arguments.ParseIntRange(viper.GetInt("maxspectators"), 0, 32, "Max Spectators"), nil, false)
	sett.Password = arguments.NewArgument[string](viper.GetString("password"), "Game Password", arguments.ParsePassword(viper.GetString("password"), "Game Password"), nil, true)
	sett.Region = arguments.NewArgument[int](viper.GetInt("region"), "Region", arguments.ParseUnsignedInt(viper.GetInt("region"), "Region"), nil, false)
	sett.AdminName = arguments.NewArgument[string](viper.GetString("adminname"), "Admin Name", nil, nil, false)
	sett.AdminMail = arguments.NewArgument[string](viper.GetString("adminmail"), "Admin Mail", arguments.ParseAdminMail, nil, true)
	sett.AdminPassword = arguments.NewArgument[string](viper.GetString("adminpassword"), "Admin Password", arguments.ParsePassword(viper.GetString("adminpassword"), "Admin Password"), nil, true)
	sett.MOTD = arguments.NewArgument[string](viper.GetString("motd"), "MOTD", nil, nil, false)
	sett.SpecimenType = arguments.NewArgument[string](viper.GetString("specimentype"), "Specimens Type", arguments.ParseSpecimentType, arguments.FormatSpecimentType, false)
	sett.Mutators = arguments.NewArgument[string](viper.GetString("mutators"), "Mutators", nil, nil, false)
	sett.ServerMutators = arguments.NewArgument[string](viper.GetString("servermutators"), "Server Mutators", nil, nil, false)
	sett.RedirectURL = arguments.NewArgument[string](viper.GetString("redirecturl"), "Redirect URL", arguments.ParseURL, nil, false)
	sett.Maplist = arguments.NewArgument[string](viper.GetString("maplist"), "Maplist", nil, nil, false)
	sett.EnableWebAdmin = arguments.NewArgument[bool](viper.GetBool("webadmin"), "Web Admin", nil, arguments.FormatBool, false)
	sett.EnableMapVote = arguments.NewArgument[bool](viper.GetBool("mapvote"), "Map Voting", nil, arguments.FormatBool, false)
	sett.MapVoteRepeatLimit = arguments.NewArgument[int](viper.GetInt("mapvote-repeatlimit"), "Map Vote Repeat Limit", arguments.ParseUnsignedInt(viper.GetInt("mapvote-repeatlimit"), "Map Vote Repeat Limit"), nil, false)
	sett.EnableAdminPause = arguments.NewArgument[bool](viper.GetBool("adminpause"), "Admin Pause", nil, arguments.FormatBool, false)
	sett.DisableWeaponThrow = arguments.NewArgument[bool](viper.GetBool("noweaponthrow"), "No Weapon Throw", nil, arguments.FormatBool, false)
	sett.DisableWeaponShake = arguments.NewArgument[bool](viper.GetBool("noweaponshake"), "No Weapon Shake", nil, arguments.FormatBool, false)
	sett.EnableThirdPerson = arguments.NewArgument[bool](viper.GetBool("thirdperson"), "Third Person View", nil, arguments.FormatBool, false)
	sett.EnableLowGore = arguments.NewArgument[bool](viper.GetBool("lowgore"), "Low Gore", nil, arguments.FormatBool, false)
	sett.Uncap = arguments.NewArgument[bool](viper.GetBool("uncap"), "Uncap Framerate", nil, arguments.FormatBool, false)
	sett.Unsecure = arguments.NewArgument[bool](viper.GetBool("unsecure"), "Unsecure (no VAC)", nil, arguments.FormatBool, false)
	sett.NoSteam = arguments.NewArgument[bool](viper.GetBool("nosteam"), "Skip SteamCMD", nil, arguments.FormatBool, false)
	sett.NoValidate = arguments.NewArgument[bool](viper.GetBool("novalidate"), "Files Validation", nil, arguments.FormatBool, false)
	sett.AutoRestart = arguments.NewArgument[bool](viper.GetBool("autorestart"), "Server Auto Restart", nil, arguments.FormatBool, false)
	sett.EnableMutLoader = arguments.NewArgument[bool](viper.GetBool("mutloader"), "MutLoader", nil, arguments.FormatBool, false)
	sett.EnableKFPatcher = arguments.NewArgument[bool](viper.GetBool("kfpatcher"), "KFPatcher", nil, arguments.FormatBool, false)
	sett.KFPHidePerks = arguments.NewArgument[bool](viper.GetBool("hideperks"), "KFP Hide Perks", nil, arguments.FormatBool, false)
	sett.KFPDisableZedTime = arguments.NewArgument[bool](viper.GetBool("nozedtime"), "KFP Disable ZED Time", nil, arguments.FormatBool, false)
	sett.KFPBuyEverywhere = arguments.NewArgument[bool](viper.GetBool("buyeverywhere"), "KFP Buy Everywhere", nil, arguments.FormatBool, false)
	sett.KFPEnableAllTraders = arguments.NewArgument[bool](viper.GetBool("alltraders"), "KFP All Traders", nil, arguments.FormatBool, false)
	sett.KFPAllTradersMessage = arguments.NewArgument[string](viper.GetString("alltraders-message"), "KFP All Traders Msg", nil, nil, false)
	sett.KFPatcherURL = arguments.NewArgument[string](viper.GetString("kfpatcher-url"), "KFPatcher URL", arguments.ParseURL, nil, false)
	sett.KFUnflectURL = arguments.NewArgument[string](viper.GetString("kfunflect-url"), "KFUnflect URL", arguments.ParseURL, nil, false)
	sett.LogToFile = arguments.NewArgument[bool](viper.GetBool("log-to-file"), "Log to File", nil, arguments.FormatBool, false)
	sett.LogLevel = arguments.NewArgument[string](viper.GetString("log-level"), "Log Level", arguments.ParseLogLevel, nil, false)
	sett.LogFile = arguments.NewArgument[string](viper.GetString("log-file"), "Log File", nil, nil, false)
	sett.LogFileFormat = arguments.NewArgument[string](viper.GetString("log-file-format"), "Log File Format", arguments.ParseLogFileFormat, nil, false)
	sett.LogMaxSize = arguments.NewArgument[int](viper.GetInt("log-max-size"), "Log Max Size (MB)", arguments.ParsePositiveInt(viper.GetInt("log-max-size"), "Log Max Size"), nil, false)
	sett.LogMaxBackups = arguments.NewArgument[int](viper.GetInt("log-max-backups"), "Log Max Backups", arguments.ParsePositiveInt(viper.GetInt("log-max-backups"), "Log Max Backups"), nil, false)
	sett.LogMaxAge = arguments.NewArgument[int](viper.GetInt("log-max-age"), "Log Max Age (days)", arguments.ParsePositiveInt(viper.GetInt("log-max-age"), "Log Max Age"), nil, false)
}
