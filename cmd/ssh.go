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
	"net"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	//"github.com/acidlemon/go-dumper"

	"github.com/mitchellh/mapstructure"
	//"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SSHCredentialResp struct {
	KeyType  string `mapstructure:"key_type"`
	Key      string `mapstructure:"key"`
	Username string `mapstructure:"username"`
	IP       string `mapstructure:"ip"`
	Port     string `mapstructure:"port"`
}

var sshHost string
var sshRole string
var sshUser string

var err error

// ssh args
var sshCmdArgs []string

var sshCommand *exec.Cmd

var userKnownHostsFile, strictHostKeyChecking string

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Get temporary SSH credentials for Linux serers",
	Long: `Will retrieve a set of one time use SSH credentials for logging
into Linux servers. The credentials are generated for a "breakglass"
user, which has elevated sudo permissions on Linux machines.`,
	Run: func(cmd *cobra.Command, args []string) {

		// setup debug
		debug = viper.GetBool("debug")
		execConn = viper.GetBool("exec")

		if debug == true {
			log.SetLevel(log.DebugLevel)
		}

		// check specific info
		if sshHost == "" {
			log.Fatal("No SSH host specified. See --help")
		}

		log.Debug("ssh host is: ", sshHost)

		//do a reverse DNS lookup to get the IP
		ip, err := net.LookupHost(sshHost)
		log.Debug("returned IP is: ", ip[0])

		if err != nil {
			log.Fatal("Error getting host IP: ", err)
		}

		if len(ip) > 1 {
			log.Fatal("Error: returned more than 1 IP - check reverse DNS: ", ip)
		}

		// get vault client
		client := getVaultClient()

		options := map[string]interface{}{
			"ip":       ip[0],
			"username": sshUser,
		}

		ssh, err := client.SSHWithMountPoint("ssh").Credential(sshRole, options)
		//ssh, err := client.Logical().Write("ssh/creds/"+sshRole, options)

		if err != nil {
			log.Fatal("Error getting credentials: ", err)
		}

		// structure for decoding secret
		var response SSHCredentialResp

		if err := mapstructure.Decode(ssh.Data, &response); err != nil {
			log.Fatal("Error parsing vault's credential response: ", err)
		}

		fmt.Printf("Your SSH Credentials are:\n username: %s\n password: %s\n", response.Username, response.Key)

		if execConn == true {

			log.Info("Exec enabled, establishing connection")

			sshpassPath, err := exec.LookPath("sshpass")

			if err == nil {
				// if we're using sshpass, make some assumptions about how we want to make the connection
				// FIXME: we should probably make this a bit nicer
				sshCmdArgs = append(sshCmdArgs, []string{"-p", string(response.Key), "ssh", "-o PubkeyAuthentication=no", "-o UserKnownHostsFile=/dev/null", "-o StrictHostKeyChecking=no", response.Username + "@" + string(ip[0])}...)
				sshCommand = exec.Command(sshpassPath, sshCmdArgs...)
			} else {
				sshCmdArgs = append(sshCmdArgs, []string{response.Username + "@" + string(ip[0])}...)
				sshCommand = exec.Command("ssh", sshCmdArgs...)
				log.Warn("Note: Install `sshpass` to automate typing in OTP")
				log.Info("OTP for the session is: ", response.Key)
			}
			log.Debug("sshCmd ", sshCommand)
			sshCommand.Stdin = os.Stdin
			sshCommand.Stdout = os.Stdout
			err = sshCommand.Run()
			if err != nil {
				log.Fatal("Error creating ssh connection: ", err)
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(sshCmd)

	// sshCmd.PersistentFlags().String("foo", "", "A help for foo")

	sshCmd.Flags().StringVarP(&sshHost, "host", "H", "", "SSH Host to get credentials for")
	sshCmd.Flags().StringVarP(&sshUser, "user", "u", "breakglass", "SSH user to get credentials for")
	sshCmd.Flags().StringVarP(&sshRole, "role", "r", "breakglass", "SSH role to get credentials for")

}
