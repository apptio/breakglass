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
	"os"

	v "github.com/apptio/breakglass/vault"

	log "github.com/Sirupsen/logrus"
	"github.com/bgentry/speakeasy"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/mitchellh/mapstructure"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//type TLSCredentialResp struct {
//	IssuingCA  string `mapstructure:"issuing_ca"`
//	PrivateKey string `mapstructure:"private_key"`
//	CAChain    string `mapstructure:"ca_chain"`
//	Cert       string `mapstructure:"certificate"`
//}

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Get temporary TLS credentials for docker daemon",
	Long:  `Will grab a TLS cert from the Vault PKI, and then use it to connect to a Docker Daemon on a remote host you specify`,
	Run: func(cmd *cobra.Command, args []string) {

		debug = viper.GetBool("debug")

		if debug == true {
			log.SetLevel(log.DebugLevel)
		}

		userName = viper.GetString("username")
		authMethod = viper.GetString("authmethod")
		vaultHost = viper.GetString("vault")

		userPass, _ = speakeasy.Ask("Please enter your password: ")

		client, err := v.New(userName, userPass, authMethod, vaultHost, vaultPort)

		if err != nil {
			log.Fatal("Error logging into vault: ", err)
		}

		options := map[string]interface{}{
			"format":      "pem",
			"common_name": "lbriggs-test",
		}

		docker, err := client.Logical().Write("ca/issue/docker", options)

		//dump.Dump(docker.Data["issuing_ca"])

		if err != nil {
			log.Fatal("Error getting credentials: ", err)
		}

		homeDir, err := homedir.Dir()

		certFile, err := os.Create(homeDir + "/.docker/cert.pem")
		defer certFile.Close()
		_, err = certFile.WriteString(docker.Data["certificate"].(string))
		if err != nil {
			log.Fatal("Error writing cert file to: "+homeDir+"/.docker/cert.pem\n", err)
		}
		certFile.Sync()

		keyFile, err := os.Create(homeDir + "/.docker/key.pem")
		defer keyFile.Close()
		_, err = keyFile.WriteString(docker.Data["private_key"].(string))

		if err != nil {
			log.Fatal("Error writing key file to: "+homeDir+"/.docker/key.pem\n", err)
		}
		keyFile.Sync()

		caFile, err := os.Create(homeDir + "/.docker/ca.pem\n")
		defer caFile.Close()
		var writeChain string
		if chains, ok := docker.Data["ca_chain"].([]interface{}); ok {
			for _, chain := range chains {
				writeChain = chain.(string)
			}
		}
		_, err = caFile.WriteString(writeChain)
		if err != nil {
			log.Fatal("Error writing ca file to: "+homeDir+"/.docker/ca.pem", err)
		}
		caFile.Sync()

		log.Info("Docker keys have been generated. They have been saved " + homeDir + "/.docker")
		log.Info("You should now be able to use docker -H tcp://<host>:4243 --tls to connect to your docker daemon")
	},
}

func init() {
	RootCmd.AddCommand(dockerCmd)

	// dockerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
