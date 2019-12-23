package cmd_test

import (
	"os"
	"testing"

	_ "github.com/gildub/analyze/cmd"
	"github.com/gildub/analyze/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestInitDefaults(t *testing.T) {
	assert.Equal(t, "", env.Config().GetString("MigrationCluster"))
	assert.Equal(t, false, env.Config().Get("Debug"))
	assert.Equal(t, false, env.Config().Get("InsecureHostKey"))
	assert.Equal(t, true, env.Config().Get("Manifests"))
	assert.Equal(t, true, env.Config().Get("Reporting"))
	assert.Equal(t, false, env.Config().Get("Silent"))
	assert.Equal(t, "", env.Config().GetString("WorkDIr"))
}

func TestInitSetValues(t *testing.T) {
	defer func() {
		os.Unsetenv("ANALYTICS_MIGRATIONCLUSTER")
		os.Unsetenv("ANALYTICS_DEBUG")
		os.Unsetenv("ANALYTICS_INSECUREHOSTKEY")
		os.Unsetenv("ANALYTICS_MANIFESTS")
		os.Unsetenv("ANALYTICS_REPORTING")
		os.Unsetenv("ANALYTICS_SILENT")
		os.Unsetenv("ANALYTICS_WORKDIR")
	}()

	os.Setenv("ANALYTICS_MIGRATIONCLUSTER", "cluster1.example.com")
	os.Setenv("ANALYTICS_DEBUG", "true")
	os.Setenv("ANALYTICS_INSECUREHOSTKEY", "true")
	os.Setenv("ANALYTICS_MANIFESTS", "false")
	os.Setenv("ANALYTICS_REPORTING", "false")
	os.Setenv("ANALYTICS_SILENT", "true")
	os.Setenv("ANALYTICS_WORKDIR", "./testdir")
	env.InitConfig()

	assert.Equal(t, "cluster1.example.com", env.Config().GetString("MigrationCluster"))
	assert.Equal(t, true, env.Config().GetBool("Debug"))
	assert.Equal(t, true, env.Config().GetBool("InsecureHostKey"))
	assert.Equal(t, false, env.Config().GetBool("Manifests"))
	assert.Equal(t, false, env.Config().GetBool("Reporting"))
	assert.Equal(t, true, env.Config().GetBool("Silent"))
	assert.Equal(t, "./testdir", env.Config().GetString("WorkDIr"))
}
