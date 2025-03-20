package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wasya-io/petit-misskey/config"
)

func NewConfig(t *testing.T) *config.Config {
	t.Helper()

	wd, _ := os.Getwd()
	cwd := filepath.Dir(wd)
	dirs := strings.Split(cwd, "/")
	traceBack := make([]string, 0)
	for i := len(dirs) - 1; i >= 0; i-- {
		traceBack = append(traceBack, "..")
		if dirs[i] == "app" {
			break
		}
	}

	require.Nil(t, os.Chdir(filepath.Join(traceBack...)))

	config := config.NewConfig()

	require.NotNil(t, config)

	return config
}
