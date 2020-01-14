package env

import (
	"os"
	"testing"
	"time"

	"github.com/gildub/phronetic/pkg/api"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
)

func TestInitConfig(t *testing.T) {
	t.Parallel()
	os.Setenv("PHRONETIC_CRIOCONFIGFILE", "dummy")
	os.Setenv("PHRONETIC_ETCDCONFIGFILE", "dummy")
	os.Setenv("PHRONETIC_MASTERCONFIGFILE", "dummy")
	os.Setenv("PHRONETIC_NODECONFIGFILE", "dummy")
	os.Setenv("PHRONETIC_REGISTRIESCONFIGFILE", "dummy")
	os.Setenv("PHRONETIC_TARGETCLUSTER", "false")
	os.Setenv("PHRONETIC_TARGETCLUSTERNAME", "")

	ConfigFile = "testdata/cpma-config.yml"
	api.K8sSrcClient = &kubernetes.Clientset{}
	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}

	expectedHomeDir, err := homedir.Dir()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name                   string
		expectedHomeDir        string
		expectedConfigFilePath string
	}{
		{
			name:                   "Init config",
			expectedHomeDir:        expectedHomeDir,
			expectedConfigFilePath: ConfigFile,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualHomeDir := viperConfig.GetString("home")
			assert.Equal(t, tc.expectedHomeDir, actualHomeDir)
			actualConfigFilePath := viperConfig.ConfigFileUsed()
			assert.Equal(t, tc.expectedConfigFilePath, actualConfigFilePath)
		})
	}
}

func TestInitFromEnv(t *testing.T) {
	type configAsset struct {
		envKey           string
		envValue         string
		configEquivalent string
	}

	var err error

	testCases := []struct {
		name         string
		sourceConfig []configAsset
	}{
		{
			name: "Read from remote",
			sourceConfig: []configAsset{
				configAsset{
					envKey:           "PHRONETIC_HOSTNAME",
					envValue:         "master-0.test.example.com",
					configEquivalent: "hostname",
				},
				configAsset{
					envKey:           "PHRONETIC_SSHLOGIN",
					envValue:         "root",
					configEquivalent: "sshlogin",
				},
				configAsset{
					envKey:           "PHRONETIC_SSHPRIVATEKEY",
					envValue:         "/home/test/.ssh/testkey",
					configEquivalent: "sshprivatekey",
				},
				configAsset{
					envKey:           "PHRONETIC_SSHPORT",
					envValue:         "22",
					configEquivalent: "sshport",
				},
				configAsset{
					envKey:           "PHRONETIC_CONFIGSOURCE",
					envValue:         "remote",
					configEquivalent: "configsource",
				},
			},
		},
		{
			name: "Read from local",
			sourceConfig: []configAsset{
				configAsset{
					envKey:           "PHRONETIC_CONFIGSOURCE",
					envValue:         "local",
					configEquivalent: "configsource",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("PHRONETIC_CLUSTERNAME", "somename")
			os.Setenv("PHRONETIC_DEBUG", "false")
			os.Setenv("PHRONETIC_HOSTNAME", "master.example.org")
			os.Setenv("PHRONETIC_MANIFESTS", "true")
			os.Setenv("PHRONETIC_REPORTING", "true")
			os.Setenv("PHRONETIC_SAVECONFIG", "false")
			os.Setenv("PHRONETIC_SILENT", "false")
			os.Setenv("PHRONETIC_WORKDIR", "testdata/out")
			os.Setenv("PHRONETIC_TARGETCLUSTER", "false")
			os.Setenv("PHRONETIC_TARGETCLUSTERNAME", "")

			os.Setenv("PHRONETIC_CRIOCONFIGFILE", "/etc/crio/crio.conf")
			os.Setenv("PHRONETIC_ETCDCONFIGFILE", "/etc/etcd/etcd.conf")
			os.Setenv("PHRONETIC_MASTERCONFIGFILE", "/etc/origin/master/master-config.yaml")
			os.Setenv("PHRONETIC_NODECONFIGFILE", "/etc/origin/node/node-config.yaml")
			os.Setenv("PHRONETIC_REGISTRIESCONFIGFILE", "/etc/containers/registries.conf")

			api.K8sSrcClient = &kubernetes.Clientset{}
			for _, asset := range tc.sourceConfig {
				err = os.Setenv(asset.envKey, asset.envValue)
				assert.NoError(t, err, "Unable to export %s=%s", asset.envKey, asset.envValue)
			}

			err = InitConfig()
			assert.NoError(t, err, "Unable to initialize config")
			for _, asset := range tc.sourceConfig {
				assert.Equal(t, asset.envValue, viperConfig.GetString(asset.configEquivalent))
			}

			for _, asset := range tc.sourceConfig {
				err = os.Unsetenv(asset.envKey)
				assert.NoError(t, err, "Unable to unset %s", asset.envKey)
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	expectedFileHook, err := NewLogFileHook(
		LogFileConfig{
			Filename: logFile,
			MaxSize:  5, // MiB
			Level:    logrus.InfoLevel,
			Formatter: &logrus.JSONFormatter{
				PrettyPrint: true,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	consoleFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
		ForceColors:     true,
	}

	expectedStderrHook := &ConsoleWriterHook{
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
		Formatter: consoleFormatter,
	}

	expectedStdoutHook := &ConsoleWriterHook{
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
		Formatter: consoleFormatter,
	}

	testCases := []struct {
		name               string
		expectedLogLevel   logrus.Level
		expectedFormatter  *logrus.TextFormatter
		expectedFileHook   logrus.Hook
		expectedStderrHook logrus.Hook
		expectedStdoutHook logrus.Hook
		debugLevel         bool
	}{
		{
			name:             "init logger",
			expectedLogLevel: logrus.InfoLevel,
			expectedFormatter: &logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: time.RFC822,
			},
			expectedFileHook:   expectedFileHook,
			expectedStderrHook: expectedStderrHook,
			expectedStdoutHook: expectedStdoutHook,
			debugLevel:         false,
		},
		{
			name:             "init logger and set log level to debug",
			expectedLogLevel: logrus.DebugLevel,
			expectedFormatter: &logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: time.RFC822,
			},
			expectedFileHook: expectedFileHook,
			debugLevel:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viperConfig.Set("debug", tc.debugLevel)
			InitLogger()
			logger := logrus.StandardLogger()
			if tc.debugLevel {
				assert.Equal(t, tc.expectedLogLevel, logrus.GetLevel())
			} else {
				assert.Equal(t, tc.expectedLogLevel, logger.GetLevel())

				assert.Equal(t, tc.expectedFileHook, logger.Hooks[logrus.InfoLevel][0])
				assert.Equal(t, tc.expectedStdoutHook, logger.Hooks[logrus.InfoLevel][1])

				assert.Equal(t, tc.expectedFileHook, logger.Hooks[logrus.ErrorLevel][0])
				assert.Equal(t, tc.expectedStderrHook, logger.Hooks[logrus.ErrorLevel][1])
			}
		})
	}
}
