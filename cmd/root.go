// Copyright © 2019 Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	//Workaround go mod vendor issue 27063
	_ "github.com/shurcooL/vfsgen"

	"github.com/gildub/analyze/pkg/env"
	"github.com/gildub/analyze/pkg/transform"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize()

	// Get config file from CLI argument an save it to viper config
	rootCmd.PersistentFlags().StringVar(&env.ConfigFile, "config", "", "config file (Default searches ./analyze.yaml, $HOME/analyze.yml)")

	// Allow insecure host key if true
	rootCmd.PersistentFlags().BoolP("allow-insecure-host", "i", false, "allow insecure ssh host key ")
	env.Config().BindPFlag("InsecureHostKey", rootCmd.PersistentFlags().Lookup("allow-insecure-host"))

	// Set log level
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug ouput")
	env.Config().BindPFlag("Debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Migration cluster name for Kubeconfig context
	rootCmd.PersistentFlags().Bool("migration-cluster", false, "Migration cluster")
	env.Config().BindPFlag("MigrationCluster", rootCmd.PersistentFlags().Lookup("migration-cluster"))

	// Flag to generate manifests
	rootCmd.PersistentFlags().BoolP("manifests", "m", true, "Generate manifests")
	env.Config().BindPFlag("Manifests", rootCmd.PersistentFlags().Lookup("manifests"))

	// Flag to generate reporting
	rootCmd.PersistentFlags().BoolP("reporting", "r", true, "Generate reporting ")
	env.Config().BindPFlag("Reporting", rootCmd.PersistentFlags().Lookup("reporting"))

	// Don't output logs to console if true
	rootCmd.PersistentFlags().BoolP("silent", "s", false, "silent mode, disable logging output to console")
	env.Config().BindPFlag("Silent", rootCmd.PersistentFlags().Lookup("silent"))

	// Get config file from an save to viper config
	rootCmd.PersistentFlags().StringP("work-dir", "w", "", "set application data working directory (Default \".\")")
	env.Config().BindPFlag("WorkDir", rootCmd.PersistentFlags().Lookup("work-dir"))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "anytics",
	Short: "Helps migration cluster configuration of a OCP 3.x cluster to OCP 4.x",
	Long:  `Helps migration cluster configuration of a OCP 3.x cluster to OCP 4.x`,
	Run: func(cmd *cobra.Command, args []string) {
		env.InitLogger()

		if err := env.InitConfig(); err != nil {
			logrus.Fatal(err)
		}

		transform.Start()
	},
	Args: cobra.MaximumNArgs(0),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It only needs to happen once.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
	}
}
