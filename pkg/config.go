package pkg

type Config struct {
	AdGuardURL           string
	LeasePath            string
	DryRun               bool
	Logger               Logger
	Username             string // Instead of B64Auth
	Password             string
	Scheme               string // Optional, defaults to "https"
	Timeout              int    // Optional, defaults to 10 seconds
	PreserveDeletedHosts bool   //  flag to control deletion behavior
	Debug                bool   // Flag to control debuggin info
}
