package tokens

import (
	"fmt"
	"strings"
)

// EnvExampleKeys is the ordered allowlist of keys written to .env.example.
var EnvExampleKeys = []string{
	EnvAppEnv,
	EnvHTTPAddr,
	EnvDatabaseURL,
	EnvPostgresUser,
	EnvPostgresPassword,
	EnvPostgresDB,
	EnvPostgresPort,
	EnvWorkerTick,
	EnvListingSearchProcessor,
	EnvAvitoHTTPBase,
	EnvRodUserDataDir,
	EnvRodBrowser,
	EnvRodFetchMode,
}

// EnvExampleValues maps each EnvExampleKeys entry to its local default.
func EnvExampleValues() map[string]string {
	return map[string]string{
		EnvAppEnv:           DefaultAppEnv,
		EnvHTTPAddr:         DefaultHTTPAddr(),
		EnvDatabaseURL:      DefaultDatabaseURL(),
		EnvPostgresUser:     DefaultPostgresUser,
		EnvPostgresPassword: DefaultPostgresPassword,
		EnvPostgresDB:       DefaultPostgresDB,
		EnvPostgresPort:     DefaultPostgresPort,
		EnvWorkerTick:             DefaultWorkerTick,
		EnvListingSearchProcessor: DefaultListingSearchProcessor,
		EnvAvitoHTTPBase:          AvitoHTTPSBase,
		EnvRodUserDataDir:         DefaultRodUserDataDir,
		EnvRodBrowser:             DefaultRodChromeBin,
		EnvRodFetchMode:           DefaultRodFetchMode,
	}
}

// RenderEnvExample returns the canonical .env.example body.
func RenderEnvExample() string {
	vals := EnvExampleValues()
	var b strings.Builder
	for _, k := range EnvExampleKeys {
		fmt.Fprintf(&b, "%s=%s\n", k, vals[k])
	}
	return b.String()
}
