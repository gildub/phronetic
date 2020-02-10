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

	"github.com/gildub/phronetic/pkg/env"
	"github.com/gildub/phronetic/pkg/transform"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize()

	// Get config file from CLI argument an save it to viper config
	rootCmd.PersistentFlags().StringVar(&env.ConfigFile, "config", "", "config file (Default searches ./phronetic.yaml, $HOME/phronetic.yml)")

	// Set log level
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug ouput")
	env.Config().BindPFlag("Debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Migration cluster name for Kubeconfig context
	rootCmd.PersistentFlags().StringP("migration-cluster", "c", "", "Migration cluster")
	env.Config().BindPFlag("MigrationCluster", rootCmd.PersistentFlags().Lookup("migration-cluster"))

	// Namespace to inspect
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace")
	env.Config().BindPFlag("Namespace", rootCmd.PersistentFlags().Lookup("namespace"))

	// Flag for Differiential mode - Running by default in CAM mode
	rootCmd.PersistentFlags().StringP("mode", "m", "", "Execution mode: source/destination differential or CAM Operator")
	env.Config().BindPFlag("Mode", rootCmd.PersistentFlags().Lookup("mode"))

	// MigPlan to search for
	rootCmd.PersistentFlags().StringP("migplan", "p", "", "MigPlan")
	env.Config().BindPFlag("MigPlan", rootCmd.PersistentFlags().Lookup("migplan"))

	// Source cluster name for Kubeconfig context
	rootCmd.PersistentFlags().StringP("source-cluster", "o", "", "Source cluster")
	env.Config().BindPFlag("SourceCluster", rootCmd.PersistentFlags().Lookup("source-cluster"))

	// Destination cluster name for Kubeconfig context
	rootCmd.PersistentFlags().StringP("destination-cluster", "t", "", "Destination cluster")
	env.Config().BindPFlag("DestinationCluster", rootCmd.PersistentFlags().Lookup("destination-cluster"))

	// Don't output logs to console if true
	rootCmd.PersistentFlags().BoolP("silent", "s", false, "silent mode, disable logging output to console")
	env.Config().BindPFlag("Silent", rootCmd.PersistentFlags().Lookup("silent"))

	// Get config file from an save to viper config
	rootCmd.PersistentFlags().StringP("work-dir", "w", "", "set application data working directory (Default \".\")")
	env.Config().BindPFlag("WorkDir", rootCmd.PersistentFlags().Lookup("work-dir"))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "phronetic",
	Short: "Helps validate migration of namespace",
	Long:  `Helps validate migration of namespace`,
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
