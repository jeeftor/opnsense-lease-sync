package pkg

type Config struct {
	AdGuardURL string
	LeasePath  string
	DryRun     bool
	Logger     Logger
	B64Auth    string // Added this field
}
