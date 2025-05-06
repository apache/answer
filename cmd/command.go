/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package answercmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/apache/answer/internal/base/conf"
	"github.com/apache/answer/internal/cli"
	"github.com/apache/answer/internal/install"
	"github.com/apache/answer/internal/migrations"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
	"github.com/spf13/cobra"
)

var (
	// dataDirPath save all answer application data in this directory. like config file, upload file...
	dataDirPath string
	// dumpDataPath dump data path
	dumpDataPath string
	// place to build new answer
	buildDir string
	// plugins needed to build in answer application
	buildWithPlugins []string
	// build output path
	buildOutput string
	// This config is used to upgrade the database from a specific version manually.
	// If you want to upgrade the database to version 1.1.0, you can use `answer upgrade -f v1.1.0`.
	upgradeVersion string
	// The fields that need to be set to the default value
	configFields []string
	// i18nSourcePath i18n from path
	i18nSourcePath string
	// i18nTargetPath i18n to path
	i18nTargetPath string
)

func init() {
	rootCmd.Version = fmt.Sprintf("%s\nrevision: %s\nbuild time: %s", Version, Revision, Time)

	rootCmd.PersistentFlags().StringVarP(&dataDirPath, "data-path", "C", "/data/", "data path, eg: -C ./data/")

	dumpCmd.Flags().StringVarP(&dumpDataPath, "path", "p", "./", "dump data path, eg: -p ./dump/data/")

	buildCmd.Flags().StringSliceVarP(&buildWithPlugins, "with", "w", []string{}, "plugins needed to build")

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "build output path")

	buildCmd.Flags().StringVarP(&buildDir, "build-dir", "b", "", "dir for build process")

	upgradeCmd.Flags().StringVarP(&upgradeVersion, "from", "f", "", "upgrade from specific version, eg: -f v1.1.0")

	configCmd.Flags().StringSliceVarP(&configFields, "with", "w", []string{}, "the fields that need to be set to the default value, eg: -w allow_password_login")

	i18nCmd.Flags().StringVarP(&i18nSourcePath, "source", "s", "", "i18n source path, eg: -s ./i18n/source")

	i18nCmd.Flags().StringVarP(&i18nTargetPath, "target", "t", "", "i18n target path, eg: -t ./i18n/target")

	for _, cmd := range []*cobra.Command{initCmd, checkCmd, runCmd, dumpCmd, upgradeCmd, buildCmd, pluginCmd, configCmd, i18nCmd} {
		rootCmd.AddCommand(cmd)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "answer",
		Short: "Answer is a minimalist open source Q&A community.",
		Long: `Answer is a minimalist open source Q&A community.
To run answer, use:
	- 'answer init' to initialize the required environment.
	- 'answer run' to launch application.`,
	}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Answer",
		Long:  `Start running Answer`,
		Run: func(_ *cobra.Command, _ []string) {
			cli.FormatAllPath(dataDirPath)
			fmt.Println("config file path: ", cli.GetConfigFilePath())
			fmt.Println("Answer is starting..........................")
			runApp()
		},
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Answer",
		Long:  `Initialize Answer with specified configuration`,
		Run: func(_ *cobra.Command, _ []string) {
			// check config file and database. if config file exists and database is already created, init done
			cli.InstallAllInitialEnvironment(dataDirPath)

			configFileExist := cli.CheckConfigFile(cli.GetConfigFilePath())
			if configFileExist {
				fmt.Println("config file exists, try to read the config...")
				c, err := conf.ReadConfig(cli.GetConfigFilePath())
				if err != nil {
					fmt.Println("read config failed: ", err.Error())
					return
				}

				fmt.Println("config file read successfully, try to connect database...")
				if cli.CheckDBTableExist(c.Data.Database) {
					fmt.Println("connect to database successfully and table already exists, do nothing.")
					return
				}
			}

			// start installation server to install
			install.Run(cli.GetConfigFilePath())
		},
	}

	upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Answer",
		Long:  `Upgrade Answer to the latest version`,
		Run: func(_ *cobra.Command, _ []string) {
			log.SetLogger(log.NewStdLogger(os.Stdout))
			cli.FormatAllPath(dataDirPath)
			cli.InstallI18nBundle(true)
			c, err := conf.ReadConfig(cli.GetConfigFilePath())
			if err != nil {
				fmt.Println("read config failed: ", err.Error())
				return
			}
			if err = migrations.Migrate(c.Debug, c.Data.Database, c.Data.Cache, upgradeVersion); err != nil {
				fmt.Println("migrate failed: ", err.Error())
				return
			}
			fmt.Println("upgrade done")
		},
	}

	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Back up data",
		Long:  `Back up database into an SQL file`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("Answer is backing up data")
			cli.FormatAllPath(dataDirPath)
			c, err := conf.ReadConfig(cli.GetConfigFilePath())
			if err != nil {
				fmt.Println("read config failed: ", err.Error())
				return
			}
			err = cli.DumpAllData(c.Data.Database, dumpDataPath)
			if err != nil {
				fmt.Println("dump failed: ", err.Error())
				return
			}
			fmt.Println("Answer backed up the data successfully.")
		},
	}

	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Check the required environment",
		Long:  `Check if the current environment meets the startup requirements`,
		Run: func(_ *cobra.Command, _ []string) {
			cli.FormatAllPath(dataDirPath)
			fmt.Println("Start checking the required environment...")
			if cli.CheckConfigFile(cli.GetConfigFilePath()) {
				fmt.Println("config file exists [✔]")
			} else {
				fmt.Println("config file not exists [x]")
			}

			if cli.CheckUploadDir() {
				fmt.Println("upload directory exists [✔]")
			} else {
				fmt.Println("upload directory not exists [x]")
			}

			c, err := conf.ReadConfig(cli.GetConfigFilePath())
			if err != nil {
				fmt.Println("read config failed: ", err.Error())
				return
			}

			if cli.CheckDBConnection(c.Data.Database) {
				fmt.Println("db connection successfully [✔]")
			} else {
				fmt.Println("db connection failed [x]")
			}
			fmt.Println("check environment all done")
		},
	}

	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build Answer with plugins",
		Long:  `Build a new Answer with plugins that you need`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("try to build a new answer with plugins:\n%s\n", strings.Join(buildWithPlugins, "\n"))
			err := cli.BuildNewAnswer(buildDir, buildOutput, buildWithPlugins, cli.OriginalAnswerInfo{
				Version:  Version,
				Revision: Revision,
				Time:     Time,
			})
			if err != nil {
				fmt.Printf("build failed %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("build new answer successfully %s\n", buildOutput)
		},
	}

	pluginCmd = &cobra.Command{
		Use:   "plugin",
		Short: "Print all plugins packed in the binary",
		Long:  `Print all plugins packed in the binary`,
		Run: func(_ *cobra.Command, _ []string) {
			_ = plugin.CallBase(func(base plugin.Base) error {
				info := base.Info()
				fmt.Printf("%s[%s] made by %s\n", info.SlugName, info.Version, info.Author)
				return nil
			})
		},
	}

	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Set some config to default value",
		Long:  `Set some config to default value`,
		Run: func(_ *cobra.Command, _ []string) {
			cli.FormatAllPath(dataDirPath)

			c, err := conf.ReadConfig(cli.GetConfigFilePath())
			if err != nil {
				fmt.Println("read config failed: ", err.Error())
				return
			}

			field := &cli.ConfigField{}
			fmt.Println(configFields)
			if len(configFields) > 0 {
				switch configFields[0] {
				case "allow_password_login":
					field.AllowPasswordLogin = true
				case "deactivate_plugin":
					if len(configFields) > 1 {
						field.DeactivatePluginSlugName = configFields[1]
					}
				default:
					fmt.Printf("field %s not support\n", configFields[0])
				}
			}
			err = cli.SetDefaultConfig(c.Data.Database, c.Data.Cache, field)
			if err != nil {
				fmt.Println("set default config failed: ", err.Error())
			} else {
				fmt.Println("set default config successfully")
			}
		},
	}

	i18nCmd = &cobra.Command{
		Use:   "i18n",
		Short: "Overwrite i18n files",
		Long:  `Merge i18n files from plugins to original i18n files. It will overwrite the original i18n files`,
		Run: func(_ *cobra.Command, _ []string) {
			if err := cli.ReplaceI18nFilesLocal(i18nTargetPath); err != nil {
				fmt.Printf("replace i18n files failed %v\n", err)
			} else {
				fmt.Printf("replace i18n files successfully\n")
			}

			fmt.Printf("try to merge i18n files from %q to %q\n", i18nSourcePath, i18nTargetPath)

			if err := cli.MergeI18nFilesLocal(i18nTargetPath, i18nSourcePath); err != nil {
				fmt.Printf("merge i18n files failed %v\n", err)
			} else {
				fmt.Printf("merge i18n files successfully\n")
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
