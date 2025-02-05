package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/K4rian/kfdsl/internal/arguments"
	"github.com/K4rian/kfdsl/internal/settings"
	appsettings "github.com/K4rian/kfdsl/internal/settings"
)

func BuildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "./kfdsl <arguments>",
		Short: "KF Dedicated Server Launcher",
		Long:  "A command-line tool to configure and run a Killing Floor dedicated server.",
		RunE:  runRootCommand,
	}

	var userHome, _ = os.UserHomeDir()

	var serverName, shortName, gameMode, startupMap, gameDifficulty, gameLength,
		password, adminName, adminMail, adminPassword, motd, specimenType, mutators,
		serverMutators, redirectURL, mapList, redirectHost, redirectDir, allTradersMessage string

	var gamePort, webadminPort, gamespyPort, maxPlayers, maxSpectators, region,
		mapVoteRepeatLimit, redirectPort, redirectMaxRequests, redirectBanTime int

	var friendlyFire float64

	var enableWebAdmin, enableMapVote, enableAdminPause, disableWeaponThrow,
		disableWeaponShake, enableThirdPerson, enableLowGore, uncap, unsecure, noSteam,
		disableValidation, enableAutoRestart, enableRedirectServer, enableMutloader,
		enableKFPatcher, disableZEDTime, enableBuyEverywhere, enableAllTraders bool

	flags := map[string]struct {
		Value   interface{}
		Desc    string
		Default interface{}
	}{
		"servername":                 {&serverName, "server name", appsettings.DefaultServerName},
		"shortname":                  {&shortName, "server short name", appsettings.DefaultShortName},
		"port":                       {&gamePort, "game UDP port", appsettings.DefaultGamePort},
		"webadminport":               {&webadminPort, "WebAdmin TCP port", appsettings.DefaultWebAdminPort},
		"gamespyport":                {&gamespyPort, "GameSpy UDP port", appsettings.DefaultGameSpyPort},
		"gamemode":                   {&gameMode, "game mode", appsettings.DefaultGameMode},
		"map":                        {&startupMap, "starting map", appsettings.DefaultStartupMap},
		"difficulty":                 {&gameDifficulty, "game difficulty (easy, normal, hard, suicidal, hell)", appsettings.DefaultGameDifficulty},
		"length":                     {&gameLength, "game length (waves) (short, medium, long)", appsettings.DefaultGameLength},
		"friendlyfire":               {&friendlyFire, "friendly fire rate (0.0-1.0)", appsettings.DefaultFriendlyFire},
		"maxplayers":                 {&maxPlayers, "maximum players", appsettings.DefaultMaxPlayers},
		"maxspectators":              {&maxSpectators, "maximum spectators", appsettings.DefaultMaxSpectators},
		"password":                   {&password, "server password", appsettings.DefaultPassword},
		"region":                     {&region, "server region", appsettings.DefaultRegion},
		"adminname":                  {&adminName, "server administrator name", appsettings.DefaultAdminName},
		"adminmail":                  {&adminMail, "server administrator email", appsettings.DefaultAdminMail},
		"adminpassword":              {&adminPassword, "server administrator password", appsettings.DefaultAdminPassword},
		"motd":                       {&motd, "message of the day", appsettings.DefaultMOTD},
		"specimentype":               {&specimenType, "specimen type (default, summer, halloween, christmas)", appsettings.DefaultSpecimenType},
		"mutators":                   {&mutators, "comma-separated mutators (command-line)", appsettings.DefaultMutators},
		"servermutators":             {&serverMutators, "comma-separated mutators (server actors)", appsettings.DefaultServerMutators},
		"redirecturl":                {&redirectURL, "redirect URL", appsettings.DefaultRedirectURL},
		"maplist":                    {&mapList, "comma-separated maps for the current game mode. Use 'all' to append all available map", appsettings.DefaultMaplist},
		"webadmin":                   {&enableWebAdmin, "enable WebAdmin panel", appsettings.DefaultEnableWebAdmin},
		"mapvote":                    {&enableMapVote, "enable map voting", appsettings.DefaultEnableMapVote},
		"mapvote-repeatlimit":        {&mapVoteRepeatLimit, "number of maps to be played before a map can repeat", appsettings.DefaultMapVoteRepeatLimit},
		"adminpause":                 {&enableAdminPause, "allow admin to pause game", appsettings.DefaultEnableAdminPause},
		"noweaponthrow":              {&disableWeaponThrow, "disable weapon throwing", appsettings.DefaultDisableWeaponThrow},
		"noweaponshake":              {&disableWeaponShake, "disable weapon-induced screen shake", appsettings.DefaultDisableWeaponShake},
		"thirdperson":                {&enableThirdPerson, "enable third-person view", appsettings.DefaultEnableThirdPerson},
		"lowgore":                    {&enableLowGore, "reduce gore", appsettings.DefaultEnableLowGore},
		"uncap":                      {&uncap, "uncap the frame rate", appsettings.DefaultUncap},
		"unsecure":                   {&unsecure, "disable VAC (Valve Anti-Cheat)", appsettings.DefaultUnsecure},
		"nosteam":                    {&noSteam, "start the server without calling SteamCMD", appsettings.DefaultNoSteam},
		"novalidate":                 {&disableValidation, "skip server files integrity check", appsettings.DefaultNoValidate},
		"autorestart":                {&enableAutoRestart, "restart server on crash", appsettings.DefaultAutoRestart},
		"redirectserver":             {&enableRedirectServer, "enable the HTTP Redirect Server", appsettings.DefaultEnableRedirectServer},
		"redirectserver-host":        {&redirectHost, "HTTP Redirect Server IP/Host", appsettings.DefaultRedirectServerHost},
		"redirectserver-port":        {&redirectPort, "HTTP Redirect Server TCP port", appsettings.DefaultRedirectServerPort},
		"redirectserver-dir":         {&redirectDir, "HTTP Redirect Server root directory", appsettings.DefaultRedirectServerDir},
		"redirectserver-maxrequests": {&redirectMaxRequests, "HTTP Redirect Server max requests per IP/minute", appsettings.DefaultRedirectServerMaxRequests},
		"redirectserver-bantime":     {&redirectBanTime, "HTTP Redirect Server ban duration (in minutes)", appsettings.DefaultRedirectServerBanTime},
		"mutloader":                  {&enableMutloader, "enable MutLoader (override inline mutators)", appsettings.DefaultEnableMutLoader},
		"kfpatcher":                  {&enableKFPatcher, "enable KFPatcher", appsettings.DefaultEnableKFPatcher},
		"nozedtime":                  {&disableZEDTime, "(KFPatcher) disable ZED time", appsettings.DefaultKFPDisableZedTime},
		"buyeverywhere":              {&enableBuyEverywhere, "(KFPatcher) allow players to shop whenever", appsettings.DefaultKFPBuyEverywhere},
		"alltraders":                 {&enableAllTraders, "(KFPatcher) make all trader's spots accessible", appsettings.DefaultKFPEnableAllTraders},
		"alltraders-message":         {&allTradersMessage, "(KFPatcher) All traders screen message", appsettings.DefaultKFPAllTradersMessage},
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
		viper.BindEnv(flag)
		viper.BindPFlag(flag, rootCmd.Flags().Lookup(flag))
	}

	viper.BindEnv("STEAMCMD_ROOT")
	viper.BindEnv("STEAMCMD_APPINSTALLDIR")
	viper.BindEnv("STEAMACC_USERNAME")
	viper.BindEnv("STEAMACC_PASSWORD")
	viper.BindEnv("KF_EXTRAARGS")

	viper.SetDefault("STEAMCMD_ROOT", path.Join(userHome, "steamcmd"))
	viper.SetDefault("STEAMCMD_APPINSTALLDIR", path.Join(userHome, "gameserver"))
	viper.SetDefault("STEAMACC_USERNAME", settings.DefaultSteamLogin)

	viper.SetEnvPrefix("KF")
	viper.AutomaticEnv()

	return rootCmd
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	settings := appsettings.Get()

	registerArguments(settings)

	if err := settings.Parse(); err != nil {
		return err
	}

	viper.SetDefault("KF_EXTRAARGS", args)
	settings.ExtraArgs = viper.GetStringSlice("KF_EXTRAARGS")
	return nil
}

func registerArguments(settings *appsettings.KFDSLSettings) {
	settings.ServerName = arguments.NewArgument[string](viper.GetString("servername"), "Server Name", arguments.ParseNonEmptyStr(viper.GetString("servername"), "Server Name"), nil, false)
	settings.ShortName = arguments.NewArgument[string](viper.GetString("shortname"), "Short Name", arguments.ParseNonEmptyStr(viper.GetString("shortname"), "Short Name"), nil, false)
	settings.GamePort = arguments.NewArgument[int](viper.GetInt("port"), "Game Port", arguments.ParsePort(viper.GetInt("port"), "Game Port"), nil, false)
	settings.WebAdminPort = arguments.NewArgument[int](viper.GetInt("webadminport"), "WebAdmin Port", arguments.ParsePort(viper.GetInt("webadminport"), "WebAdmin Port"), nil, false)
	settings.GameSpyPort = arguments.NewArgument[int](viper.GetInt("gamespyport"), "GameSpy Port", arguments.ParsePort(viper.GetInt("gamespyport"), "GameSpy Port"), nil, false)
	settings.GameMode = arguments.NewArgument[string](viper.GetString("gamemode"), "Game Mode", arguments.ParseGameMode, arguments.FormatGameMode, false)
	settings.StartupMap = arguments.NewArgument[string](viper.GetString("map"), "Startup Map", arguments.ParseNonEmptyStr(viper.GetString("map"), "Startup Map"), nil, false)
	settings.GameDifficulty = arguments.NewArgument[int](viper.GetString("difficulty"), "Game Difficulty", arguments.ParseGameDifficulty, arguments.FormatGameDifficulty, false)
	settings.GameLength = arguments.NewArgument[int](viper.GetString("length"), "Game Length", arguments.ParseGameLength, arguments.FormatGameLength, false)
	settings.FriendlyFire = arguments.NewArgument[float64](viper.GetFloat64("friendlyfire"), "Friendly Fire Rate", arguments.ParseFriendlyFireRate, arguments.FormatFriendlyFireRate, false)
	settings.MaxPlayers = arguments.NewArgument[int](viper.GetInt("maxplayers"), "Max Players", arguments.ParseIntRange(viper.GetInt("maxplayers"), 0, 32, "Max Players"), nil, false)
	settings.MaxSpectators = arguments.NewArgument[int](viper.GetInt("maxspectators"), "Max Spectators", arguments.ParseIntRange(viper.GetInt("maxspectators"), 0, 32, "Max Spectators"), nil, false)
	settings.Password = arguments.NewArgument[string](viper.GetString("password"), "Game Password", arguments.ParsePassword(viper.GetString("password"), "Game Password"), nil, true)
	settings.Region = arguments.NewArgument[int](viper.GetInt("region"), "Region", arguments.ParseUnsignedInt(viper.GetInt("region"), "Region"), nil, false)
	settings.AdminName = arguments.NewArgument[string](viper.GetString("adminname"), "Admin Name", nil, nil, false)
	settings.AdminMail = arguments.NewArgument[string](viper.GetString("adminmail"), "Admin Mail", arguments.ParseAdminMail, nil, true)
	settings.AdminPassword = arguments.NewArgument[string](viper.GetString("adminpassword"), "Admin Password", arguments.ParsePassword(viper.GetString("adminpassword"), "Admin Password"), nil, true)
	settings.MOTD = arguments.NewArgument[string](viper.GetString("motd"), "MOTD", nil, nil, false)
	settings.SpecimenType = arguments.NewArgument[string](viper.GetString("specimentype"), "Specimens Type", arguments.ParseSpecimentType, arguments.FormatSpecimentType, false)
	settings.Mutators = arguments.NewArgument[string](viper.GetString("mutators"), "Mutators", nil, nil, false)
	settings.ServerMutators = arguments.NewArgument[string](viper.GetString("servermutators"), "Server Mutators", nil, nil, false)
	settings.RedirectURL = arguments.NewArgument[string](viper.GetString("redirecturl"), "Redirect URL", arguments.ParseURL, nil, false)
	settings.Maplist = arguments.NewArgument[string](viper.GetString("maplist"), "Maplist", nil, nil, false)
	settings.EnableWebAdmin = arguments.NewArgument[bool](viper.GetBool("webadmin"), "Web Admin", nil, arguments.FormatBool, false)
	settings.EnableMapVote = arguments.NewArgument[bool](viper.GetBool("mapvote"), "Map Voting", nil, arguments.FormatBool, false)
	settings.MapVoteRepeatLimit = arguments.NewArgument[int](viper.GetInt("mapvote-repeatlimit"), "Map Vote Repeat Limit", arguments.ParseUnsignedInt(viper.GetInt("mapvote-repeatlimit"), "Map Vote Repeat Limit"), nil, false)
	settings.EnableAdminPause = arguments.NewArgument[bool](viper.GetBool("adminpause"), "Admin Pause", nil, arguments.FormatBool, false)
	settings.DisableWeaponThrow = arguments.NewArgument[bool](viper.GetBool("noweaponthrow"), "No Weapon Throw", nil, arguments.FormatBool, false)
	settings.DisableWeaponShake = arguments.NewArgument[bool](viper.GetBool("noweaponshake"), "No Weapon Shake", nil, arguments.FormatBool, false)
	settings.EnableThirdPerson = arguments.NewArgument[bool](viper.GetBool("thirdperson"), "Third Person View", nil, arguments.FormatBool, false)
	settings.EnableLowGore = arguments.NewArgument[bool](viper.GetBool("lowgore"), "Low Gore", nil, arguments.FormatBool, false)
	settings.Uncap = arguments.NewArgument[bool](viper.GetBool("uncap"), "Uncap Framerate", nil, arguments.FormatBool, false)
	settings.Unsecure = arguments.NewArgument[bool](viper.GetBool("unsecure"), "Unsecure (no VAC)", nil, arguments.FormatBool, false)
	settings.NoSteam = arguments.NewArgument[bool](viper.GetBool("nosteam"), "Skip SteamCMD", nil, arguments.FormatBool, false)
	settings.NoValidate = arguments.NewArgument[bool](viper.GetBool("novalidate"), "Files Validation", nil, arguments.FormatBool, false)
	settings.AutoRestart = arguments.NewArgument[bool](viper.GetBool("autorestart"), "Server Auto Restart", nil, arguments.FormatBool, false)
	settings.EnableRedirectServer = arguments.NewArgument[bool](viper.GetBool("redirectserver"), "Redirect Server", nil, arguments.FormatBool, false)
	settings.RedirectServerHost = arguments.NewArgument[string](viper.GetString("redirectserver-host"), "Redirect Server Host", arguments.ParseIP, nil, false)
	settings.RedirectServerPort = arguments.NewArgument[int](viper.GetInt("redirectserver-port"), "Redirect Server Port", arguments.ParsePort(viper.GetInt("redirectserver-port"), "Redirect Server Port"), nil, false)
	settings.RedirectServerDir = arguments.NewArgument[string](viper.GetString("redirectserver-dir"), "Redirect Server Root", arguments.ParseExistingDir, nil, false)
	settings.RedirectServerMaxRequests = arguments.NewArgument[int](viper.GetInt("redirectserver-maxrequests"), "Redirect Server Max Req.", arguments.ParseUnsignedInt(viper.GetInt("redirectserver-maxrequests"), "Redirect Server Max Req."), nil, false)
	settings.RedirectServerBanTime = arguments.NewArgument[int](viper.GetInt("redirectserver-bantime"), "Redirect Server Ban Time", arguments.ParseUnsignedInt(viper.GetInt("redirectserver-bantime"), "Redirect Server Ban Time"), nil, false)
	settings.EnableMutLoader = arguments.NewArgument[bool](viper.GetBool("mutloader"), "MutLoader", nil, arguments.FormatBool, false)
	settings.EnableKFPatcher = arguments.NewArgument[bool](viper.GetBool("kfpatcher"), "KFPatcher", nil, arguments.FormatBool, false)
	settings.KFPDisableZedTime = arguments.NewArgument[bool](viper.GetBool("nozedtime"), "KFP Disable ZED Time", nil, arguments.FormatBool, false)
	settings.KFPBuyEverywhere = arguments.NewArgument[bool](viper.GetBool("buyeverywhere"), "KFP Buy Everywhere", nil, arguments.FormatBool, false)
	settings.KFPEnableAllTraders = arguments.NewArgument[bool](viper.GetBool("alltraders"), "KFP All Traders", nil, arguments.FormatBool, false)
	settings.KFPAllTradersMessage = arguments.NewArgument[string](viper.GetString("alltraders-message"), "KFP All Traders Msg", nil, nil, false)
}
