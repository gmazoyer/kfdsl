package config

type ServerIniFile interface {
	FilePath() string
	Load(filePath string) error
	Save(filePath string) error

	GetServerName() string
	GetShortName() string
	GetGamePort() int
	GetWebAdminPort() int
	GetGameSpyPort() int
	GetGameDifficulty() int
	GetGameLength() int
	GetFriendlyFireRate() float64
	GetMaxPlayers() int
	GetMaxSpectators() int
	GetPassword() string
	GetRegion() int
	GetAdminName() string
	GetAdminMail() string
	GetAdminPassword() string
	GetMOTD() string
	GetSpecimenType() string
	GetRedirectURL() string
	IsWebAdminEnabled() bool
	IsMapVoteEnabled() bool
	GetMapVoteRepeatLimit() int
	IsAdminPauseEnabled() bool
	IsWeaponThrowingEnabled() bool
	IsWeaponShakeEffectEnabled() bool
	IsThirdPersonEnabled() bool
	IsLowGoreEnabled() bool
	GetMaxInternetClientRate() int

	SetServerName(servername string) bool
	SetShortName(shortname string) bool
	SetGamePort(port int) bool
	SetWebAdminPort(port int) bool
	SetGameSpyPort(port int) bool
	SetGameDifficulty(difficulty int) bool
	SetGameLength(length int) bool
	SetFriendlyFireRate(rate float64) bool
	SetMaxPlayers(players int) bool
	SetMaxSpectators(spectators int) bool
	SetPassword(password string) bool
	SetRegion(region int) bool
	SetAdminName(adminame string) bool
	SetAdminMail(adminmail string) bool
	SetAdminPassword(adminpassword string) bool
	SetMOTD(motd string) bool
	SetSpecimenType(specimentype string) bool
	SetRedirectURL(url string) bool
	SetWebAdminEnabled(enabled bool) bool
	SetMapVoteEnabled(enabled bool) error
	SetMapVoteRepeatLimit(limit int) bool
	SetAdminPauseEnabled(enabled bool) bool
	SetWeaponThrowingEnabled(enabled bool) bool
	SetWeaponShakeEffectEnabled(enabled bool) bool
	SetThirdPersonEnabled(enabled bool) bool
	SetLowGoreEnabled(enabled bool) bool
	SetMaxInternetClientRate(rate int) bool

	ServerMutatorExists(mutator string) bool
	ClearServerMutators() error
	SetServerMutators(mutators []string) error

	ClearMaplist(sectionName string) error
	SetMaplist(sectionName string, maps []string) error
}
