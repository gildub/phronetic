package cmd_test

import (
	"os"
	"testing"

	_ "github.com/gildub/phronetic/cmd"
	"github.com/gildub/phronetic/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestInitDefaults(t *testing.T) {
	assert.Equal(t, "", env.Config().GetString("MigrationCluster"))
	assert.Equal(t, false, env.Config().Get("Debug"))
	assert.Equal(t, false, env.Config().Get("Silent"))
	assert.Equal(t, "", env.Config().GetString("WorkDIr"))
}

func TestInitSetValues(t *testing.T) {
	defer func() {
		os.Unsetenv("PHRONETIC_MIGRATIONCLUSTER")
		os.Unsetenv("PHRONETIC_DEBUG")
		os.Unsetenv("PHRONETIC_SILENT")
		os.Unsetenv("PHRONETIC_WORKDIR")
	}()

	os.Setenv("PHRONETIC_MIGRATIONCLUSTER", "cluster1.example.com")
	os.Setenv("PHRONETIC_DEBUG", "true")
	os.Setenv("PHRONETIC_SILENT", "true")
	os.Setenv("PHRONETIC_WORKDIR", "./testdir")
	env.InitConfig()

	assert.Equal(t, "cluster1.example.com", env.Config().GetString("MigrationCluster"))
	assert.Equal(t, true, env.Config().GetBool("Debug"))
	assert.Equal(t, true, env.Config().GetBool("Silent"))
	assert.Equal(t, "./testdir", env.Config().GetString("WorkDIr"))
}
