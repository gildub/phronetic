package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gildub/phronetic/pkg/api"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// AppName holds the name of this application
	AppName = "phronetic"
	logFile = "phronetic.log"
)

var (
	// ConfigFile - keeps full path to the configuration file
	ConfigFile string

	viperConfig *viper.Viper
)

func init() {
	viperConfig = viper.New()
}

// Config returns pointer to the viper config
func Config() *viper.Viper {
	return viperConfig
}

// InitConfig initializes application's configuration
func InitConfig() (err error) {
	// Fill in environment variables that match
	viperConfig.SetEnvPrefix("PHRONETIC")
	viperConfig.AutomaticEnv()

	if err := setConfigLocation(); err != nil {
		return err
	}

	// If a config file is found, read it in.
	readConfigErr := viperConfig.ReadInConfig()
	// If no config file and save config file is undetermined, ask to create or save it for future use
	if readConfigErr != nil && viperConfig.GetString("SaveConfig") != "false" {
		if err := surveySaveConfig(); err != nil {
			return handleInterrupt(err)
		}
		logrus.Debug("Can't read config file, all values were prompted and new config was asked to be created, err: ", readConfigErr)
	}

	// Parse kubeconfig for creating api client later
	if err := api.ParseKubeConfig(); err != nil {
		return errors.Wrap(err, "kubeconfig parsing failed")
	}

	// Ask for all values that are missing in ENV, flags or config yaml
	if err := surveyMissingValues(); err != nil {
		return handleInterrupt(err)
	}

	if viperConfig.GetString("SaveConfig") == "true" {
		viperConfig.WriteConfig()
	}

	if err := createClients(); err != nil {
		return handleInterrupt(err)
	}

	return nil
}

// setConfigLocation sets location for phronetic configuration
func setConfigLocation() (err error) {
	var home string
	// Find home directory.
	home, err = homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "Can't detect home user directory")
	}
	viperConfig.Set("home", home)

	// Try to find config file if it wasn't provided as a flag
	if ConfigFile == "" {
		ConfigFile = path.Join(home, "phronetic.yaml")
	}
	viperConfig.SetConfigFile(ConfigFile)
	return
}

func surveyMissingValues() error {
	if err := surveySaveConfig(); err != nil {
		return err
	}

	if err := surveyMigCluster(); err != nil {
		return err
	}

	if err := surveyMigPlan(); err != nil {
		return err
	}

	if err := surveyNamespace(); err != nil {
		return err
	}

	if viperConfig.GetString("WorkDir") == "" {
		workDir := "."
		prompt := &survey.Input{
			Message: "Path to application data, skip to use current directory",
			Default: ".",
		}
		if err := survey.AskOne(prompt, &workDir); err != nil {
			return err
		}

		viperConfig.Set("WorkDir", workDir)
	}

	return nil
}

func surveyMigCluster() error {
	migClusterName := viperConfig.GetString("MigrationCluster")
	if !viperConfig.InConfig("MigrationCluster") && migClusterName == "" {
		discoverCluster := ""
		clusterName := ""
		var err error

		// Ask for source of master hostname, prompt or find it using KUBECONFIG
		prompt := &survey.Select{
			Message: "Do wish to find source cluster using KUBECONFIG or prompt it?",
			Options: []string{"KUBECONFIG", "prompt"},
		}
		if err := survey.AskOne(prompt, &discoverCluster); err != nil {
			return err
		}

		if discoverCluster == "KUBECONFIG" {
			if clusterName, err = discoverMigCluster(); err != nil {
				return err
			}
			viperConfig.Set("MigrationCluster", clusterName)
		} else {
			prompt := &survey.Input{
				Message: "Cluster name",
			}
			if err := survey.AskOne(prompt, &clusterName); err != nil {
				return err
			}

			viperConfig.Set("MigrationCluster", clusterName)
		}
	}

	return nil
}

func surveyNamespace() error {
	// Ask namespace to run analysis for
	namespace := ""
	if viperConfig.GetString("Namespace") == "" {
		prompt := &survey.Input{
			Message: "What namespace to inspect?",
		}
		if err := survey.AskOne(prompt, &namespace); err != nil {
			return err
		}
		viperConfig.Set("Namespace", &namespace)
	}
	return nil
}

func surveyMigPlan() error {
	// Ask MigPlan to run analysis for
	migPlan := ""
	if viperConfig.GetString("MigPlan") == "" {
		prompt := &survey.Input{
			Message: "What MigPlan to search for?",
		}
		if err := survey.AskOne(prompt, &migPlan); err != nil {
			return err
		}
		viperConfig.Set("MigPlan", &migPlan)
	}
	return nil

}

// discoverMigCluster Get kubeconfig using $KUBECONFIG, if not try ~/.kube/config
// parse kubeconfig and select targeted cluster from available contexts
func discoverMigCluster() (string, error) {
	selectedCluster := surveyClusters()
	if selectedCluster == "" {
		return "", nil
	}
	return selectedCluster, nil
}

// surveyClusters list clusters from kubeconfig
func surveyClusters() string {
	// Survey options should be an array
	clusters := make([]string, 0, len(api.ClusterNames))
	// It's better to have current context's cluster first, because
	// it will be easier to select it using survey
	currentContext := api.KubeConfig.CurrentContext

	currentContextCluster := api.KubeConfig.Contexts[currentContext].Cluster
	clusters = append(clusters, currentContextCluster)

	for cluster := range api.ClusterNames {
		if cluster != currentContextCluster {
			clusters = append(clusters, cluster)
		}
	}

	selectedCluster := ""
	prompt := &survey.Select{
		Message: "Select cluster obtained from KUBECONFIG contexts",
		Options: clusters,
	}
	survey.AskOne(prompt, &selectedCluster)

	return selectedCluster
}

func surveySaveConfig() (err error) {
	saveConfig := viperConfig.GetString("SaveConfig")
	if saveConfig == "" {
		prompt := &survey.Select{
			Message: "Do you wish to save configuration for future use?",
			Options: []string{"true", "false"},
		}
		if err := survey.AskOne(prompt, &saveConfig); err != nil {
			return err
		}
	}
	if saveConfig == "true" {
		viperConfig.Set("SaveConfig", true)
	} else {
		viperConfig.Set("SaveConfig", false)
	}

	return nil
}

func handleInterrupt(err error) error {
	switch {
	case err.Error() == "interrupt":
		return errors.Wrap(err, "Exiting.")
	default:
		return errors.Wrap(err, "Error in creating config file")
	}
}

func createClients() error {
	migClusterName := viperConfig.GetString("MigrationCluster")
	// set current context to selected cluster
	api.KubeConfig.CurrentContext = api.ClusterNames[migClusterName]
	if err := api.CreateCtrlClient(migClusterName); err != nil {
		return errors.Wrap(err, "k8s controller client failed to create")
	}

	migPlan, err := api.GetMigPlan(api.CtrlClient, viperConfig.GetString("MigPlan"))

	if err != nil {
		logrus.Fatal("No MigPlan avail.")
	}

	dstCluster := migPlan.Spec.DestMigClusterRef.Name
	srcCluster := migPlan.Spec.SrcMigClusterRef.Name

	srcMigCluster := api.GetMigCluster(api.CtrlClient, srcCluster)
	dstMigCluster := api.GetMigCluster(api.CtrlClient, dstCluster)

	if srcMigCluster.Spec.IsHostCluster {
		if err := api.CreateK8sSrcClient(migClusterName); err != nil {
			return errors.Wrap(err, "Source Cluster: k8s api client failed to create")
		}
	} else {
		noScheme := strings.Trim(srcMigCluster.Spec.URL, "https://")
		srcClusterEndpoint := strings.ReplaceAll(noScheme, ".", "-")
		srcContext, err := getContext(srcClusterEndpoint)
		if err != nil {
			logrus.Fatal(err)
		}

		// set current context to selected cluster
		api.KubeConfig.CurrentContext = srcContext
		if err := api.CreateK8sSrcClient(srcContext); err != nil {
			return errors.Wrap(err, "Source Cluster: k8s api client failed to create")
		}
	}

	if dstMigCluster.Spec.IsHostCluster {
		// set current context to selected cluster
		dstContext := api.ClusterNames[migClusterName]
		api.KubeConfig.CurrentContext = dstContext
		if err := api.CreateK8sDstClient(migClusterName); err != nil {
			return errors.Wrap(err, "Destination Cluster: k8s api client failed to create")
		}
	} else {
		noScheme := strings.Trim(dstMigCluster.Spec.URL, "https://")
		dstClusterEndpoint := strings.ReplaceAll(noScheme, ".", "-")
		dstContext, err := getContext(dstClusterEndpoint)
		if err != nil {
			logrus.Fatal(err)
		}
		// set current context to selected cluster
		api.KubeConfig.CurrentContext = dstContext
		if err := api.CreateK8sDstClient(dstContext); err != nil {
			return errors.Wrap(err, "Destination Cluster: k8s api client failed to create")
		}
	}
	return nil
}

func getContext(clusterEndpoint string) (string, error) {
	context := api.ClusterNames[clusterEndpoint]
	if context == "" {
		for _, e := range api.ClusterNames {
			if strings.Contains(e, clusterEndpoint) {
				return e, nil
			}
		}
		return "", errors.New(fmt.Sprintf("Can't find cluster %s in Kubeconfig", clusterEndpoint))
	}
	return context, nil
}

// InitLogger initializes stderr and logger to file
func InitLogger() {
	logLevel := logrus.InfoLevel
	if viperConfig.GetBool("debug") {
		logLevel = logrus.DebugLevel
		logrus.SetReportCaller(true)
	}
	logrus.SetLevel(logLevel)

	logrus.SetOutput(ioutil.Discard)

	fileHook, _ := NewLogFileHook(
		LogFileConfig{
			Filename: logFile,
			MaxSize:  5, // MiB
			Level:    logLevel,
			Formatter: &logrus.JSONFormatter{
				PrettyPrint: true,
			},
		},
	)
	logrus.AddHook(fileHook)

	consoleFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
		ForceColors:     true,
	}

	if !viperConfig.GetBool("silent") {
		stdoutHook := &ConsoleWriterHook{
			Writer: os.Stdout,
			LogLevels: []logrus.Level{
				logrus.InfoLevel,
				logrus.DebugLevel,
			},
			Formatter: consoleFormatter,
		}

		logrus.AddHook(stdoutHook)
	}

	stderrHook := &ConsoleWriterHook{
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
		Formatter: consoleFormatter,
	}

	logrus.AddHook(stderrHook)

	logrus.Debugf("%s is running in debug mode", AppName)
}
