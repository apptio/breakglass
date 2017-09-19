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
	"os/exec"

	v "github.com/apptio/breakglass/vault"
	"github.com/bgentry/speakeasy"

	//"github.com/acidlemon/go-dumper"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
)

var mysqlHost string

var mysqlCmdArgs []string

var mysqlCommand *exec.Cmd

var mysqlRole string

type MySQLCredentialResp struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// mysqlCmd represents the mysql command
var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Get temporary login credentials for mysql servers",
	Long: `Generates temporary credentials on a mysql server that you specify
and returns a user name and password you can use to login`,
	Run: func(cmd *cobra.Command, args []string) {

		debug = viper.GetBool("debug")
		execConn = viper.GetBool("exec")

		if debug == true {
			log.SetLevel(log.DebugLevel)
		}

		if mysqlHost == "" {
			log.Fatal("No MySQL host specified. See --help")
		}

		userName = viper.GetString("username")
		authMethod = viper.GetString("authmethod")
		vaultHost = viper.GetString("vault")

		log.WithFields(log.Fields{"username": userName,
			"authmethod": authMethod,
			"vaulthost":  vaultHost,
			"vaultport":  vaultPort}).Debug("mysql host is: ", sshHost)

		if vaultHost == "" {
			log.Fatal("No Vault host specified. See --help")
		}

		log.WithFields(log.Fields{"username": userName,
			"authmethod": authMethod,
			"vaulthost":  vaultHost,
			"vaultport":  vaultPort}).Debug("prompting for password")

		userPass, _ = speakeasy.Ask("Please enter your password: ")

		client, err := v.New(v.Params{Username: userName, Password: userPass, Method: authMethod, Host: vaultHost, Port: vaultPort})

		mysql, err := client.Logical().Read("mysql/" + mysqlHost + "/creds/" + mysqlRole)

		if err != nil {
			log.Fatal("Error getting credentials: ", err)
		}

		if mysql == nil {
			log.Fatal("No credentials were retrieved. Check this host is enabled in vault: ", mysqlHost)
		}

		var response MySQLCredentialResp

		if err := mapstructure.Decode(mysql.Data, &response); err != nil {
			log.Fatal("Error parsing vault's credential response: ", err)
		}

		fmt.Printf("Your MySQL Credentials are below\n username: %s\n password: %s\n", response.Username, response.Password)

		if execConn == true {
			log.Info("Exec enabled, establishing connection")

			mysqlCmdPath, err := exec.LookPath("mysql")

			if err != nil {
				log.Fatal("mysql client not found in $PATH, can't establish connection", err)
			}

			mysqlCmdArgs = append(mysqlCmdArgs, []string{"-h", mysqlHost, "-u", string(response.Username), "--password=" + string(response.Password)}...)
			mysqlCommand = exec.Command(mysqlCmdPath, mysqlCmdArgs...)

			log.Debug("Initiating MySQL Connection", mysqlCommand)

			mysqlCommand.Stdin = os.Stdin
			mysqlCommand.Stdout = os.Stdout

			err = mysqlCommand.Run()

			if err != nil {
				log.Fatal("Error creating mysql connection: ", err)
			}

		}

		//dump.Dump(mysql.Data)

	},
}

func init() {
	RootCmd.AddCommand(mysqlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mysqlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	mysqlCmd.Flags().StringVarP(&mysqlHost, "host", "H", "", "MySQL Host to get credentials for")
	mysqlCmd.Flags().StringVarP(&mysqlRole, "role", "r", "readonly", "MySQL role to get credentials for")

}
