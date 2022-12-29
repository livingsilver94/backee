package installer

import (
	"os"
	"regexp"
	"strings"
)

const (
	envVarPrefix  = "${"
	envVarPattern = "[a-zA-Z_]\\w+"
	envVarSuffix  = "}"
)

var envVarRegex *regexp.Regexp

func init() {
	envVarRegex = regexp.MustCompile(regexp.QuoteMeta(envVarPrefix) + envVarPattern + regexp.QuoteMeta(envVarSuffix))
}

func ReplaceEnvVars(src string) string {
	replacer := func(match string) string {
		varName := strings.TrimPrefix(match, envVarPrefix)
		varName = strings.TrimSuffix(varName, envVarSuffix)
		return os.Getenv(varName)
	}
	return envVarRegex.ReplaceAllStringFunc(src, replacer)
}
