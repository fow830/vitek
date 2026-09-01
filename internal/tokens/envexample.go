package tokens

import (
	"fmt"
	"strings"
)

// EnvExampleKeys is the ordered allowlist of keys written to .env.example.
var EnvExampleKeys = []string{
	EnvAppEnv,
	EnvHTTPAddr,
	EnvLogLevel,
	EnvDatabaseURL,
	EnvRedisURL,
	EnvPostgresUser,
	EnvPostgresPassword,
	EnvPostgresDB,
	EnvPostgresPort,
	EnvRedisPort,
	EnvHTTPPort,
	EnvWorkerTick,
}

// EnvExampleValues maps each EnvExampleKeys entry to its local default.
func EnvExampleValues() map[string]string {
	return map[string]string{
		EnvAppEnv:           DefaultAppEnv,
		EnvHTTPAddr:         DefaultHTTPAddr(),
		EnvLogLevel:         DefaultLogLevel,
		EnvDatabaseURL:      DefaultDatabaseURL(),
		EnvRedisURL:         DefaultRedisURL(),
		EnvPostgresUser:     DefaultPostgresUser,
		EnvPostgresPassword: DefaultPostgresPassword,
		EnvPostgresDB:       DefaultPostgresDB,
		EnvPostgresPort:     DefaultPostgresPort,
		EnvRedisPort:        DefaultRedisPort,
		EnvHTTPPort:         DefaultHTTPPort,
		EnvWorkerTick:       DefaultWorkerTick,
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
