// Copyright © 2017 Apptio
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
)

var cfgFile string
var vaultHost string
var vaultPort int

var userName string
var authMethod string
var userPass string

var debug bool
var execConn bool

var Version string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "breakglass",
	Short: "Get elevated privileges on Apptio Infrastructure",
	Long: `breakglass allows you to get login credentials for a variety of 
Apptio infrastructure, such as databases servers, Linux servers (ssh credentials)
and AWS IAM roles`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	Version = version
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.breakglass/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&vaultHost, "vault", "v", "", "vault host to authenticate against")
	RootCmd.PersistentFlags().IntVarP(&vaultPort, "port", "p", 8200, "port of vault servers to use when authenticating")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "enable debug logging")
	RootCmd.PersistentFlags().BoolVarP(&execConn, "exec", "", false, "Initiate connection with credentials")
	RootCmd.PersistentFlags().StringVarP(&userName, "username", "", "", "username to authenticate to vault with")
	viper.BindPFlag("vault", RootCmd.PersistentFlags().Lookup("vault"))
	viper.BindPFlag("username", RootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("exec", RootCmd.PersistentFlags().Lookup("exec"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/breakglass")
	viper.AddConfigPath("$HOME/.breakglass") // adding home directory as first search path
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}

	// get the current logged in user for defaults
	currentUser, err := user.Current()

	if err != nil {
		log.Fatal("Error retrieving current user: ", err)
	}

	// set some sane defaults
	viper.SetDefault("username", currentUser.Username)
	viper.SetDefault("authmethod", "ldap")

}
