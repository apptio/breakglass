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
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	v "github.com/apptio/breakglass/vault"
	garbler "github.com/michaelbironneau/garbler/lib"

	//"github.com/acidlemon/go-dumper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/bgentry/speakeasy"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/Sirupsen/logrus"
)

var awsCreateLoginProfile bool
var awsRole string

type AWSCredentialResp struct {
	AccessKey     string `mapstructure:"access_key"`
	SecretKey     string `mapstructure:"secret_key"`
	SecurityToken string `mapstructure:"security_token"`
}

// mysqlCmd represents the mysql command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get temporary login credentials for aws",
	Long: `Generates temporary credentials for an AWS account
and returns a user name and password you can use to login`,
	Run: func(cmd *cobra.Command, args []string) {
		debug = viper.GetBool("debug")
		execConn = viper.GetBool("exec")

		if debug == true {
			log.SetLevel(log.DebugLevel)
		}

		if awsRole == "" {
			log.Fatal("No AWS role host specified. See --help")
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

		// Get a Vault client
		client, err := v.New(userName, userPass, authMethod, vaultHost, vaultPort)

		// Read new AWS credentials from Vault
		log.Debug("Reading Vault role: ", awsRole)
		secret, err := client.Logical().Read(awsRole)
		if err != nil {
			log.Fatal("Error getting credentials: ", err)
		}
		log.Debug("Vault LeaseID: ", secret.LeaseID)

		// Decode Vault response
		var response AWSCredentialResp
		if err := mapstructure.Decode(secret.Data, &response); err != nil {
			log.Fatal("Error parsing vault's credential response: ", err)
		}

		// Print credentials
		fmt.Println("Your AWS Credentials are below:")
		fmt.Println("access_key: ", response.AccessKey)
		fmt.Println("secret_key: ", response.SecretKey)
		fmt.Println("security_token: ", response.SecurityToken)

		// Quit unless we're creating a login profile
		if !awsCreateLoginProfile {
			os.Exit(0)
		}

		// Set up AWS credentials
		creds := credentials.NewStaticCredentials(response.AccessKey, response.SecretKey, "")
		_, err = creds.Get()
		if err != nil {
			fmt.Printf("bad credentials: %s", err)
		}

		// Create a new AWS session
		cfg := aws.NewConfig().WithCredentials(creds)
		svc := iam.New(session.New(), cfg)

		// AWS is eventually consistent so the account might not be available
		// immediately. So do this janky loop thing until it becomes available.
		log.Info("Waiting for account to become available...")
		for {
			_, err = svc.ListUsers(nil)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == "InvalidClientTokenId" {
						// We'll see this error code until the account is available
						log.Debug("Account is not available yet...")
						time.Sleep(time.Second)
						continue
					} else {
						// If the error code is anything else we have other problems
						log.Fatal("AWS returned an unexpected error: ", err)
					}
				} else {
					// And catch any non-AWS errors:
					log.Fatal("Error: ", err.Error())
				}
			}
			break
		}
		log.Info("Account is available. Creating Login Profile...")

		// GetAccessKeyLastUsed returns a struct with the Username that vault has generated
		lastused, err := svc.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{
			AccessKeyId: &response.AccessKey,
		})
		if err != nil {
			log.Fatal("Could not determine AWS Username: ", err.Error())
		}

		// Generate random password
		reqs := garbler.PasswordStrengthRequirements{
			MinimumTotalLength: 20,
			Punctuation:        5,
			Uppercase:          1,
			Digits:             1,
		}
		password, _ := garbler.NewPassword(&reqs)
		required := false

		// Print out the username
		fmt.Println("UserName: ", *lastused.UserName)
		fmt.Println("Password: ", password)

		// Create the Login Profile
		_, err = svc.CreateLoginProfile(&iam.CreateLoginProfileInput{
			Password:              &password,
			PasswordResetRequired: &required,
			UserName:              lastused.UserName,
		})
		if err != nil {
			log.Fatal("Could not create Login Profile: ", err)
		}

		// Set up a trap for Ctrl-C so we can clean up the Login Profile
		var wg sync.WaitGroup
		wg.Add(1)
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			defer wg.Done()
			// If Ctrl-C is hit really quickly the login profile might not be
			// available yet. So do another janky loop thing.
			for {
				_, err = svc.DeleteLoginProfile(&iam.DeleteLoginProfileInput{
					UserName: lastused.UserName,
				})
				if err != nil {
					if awsErr, ok := err.(awserr.Error); ok {
						if awsErr.Code() == "EntityTemporarilyUnmodifiable" {
							// If the Login Profile is not ready we'll see this error code
							log.Debug("Login Profile is not available yet...")
							time.Sleep(time.Second)
							continue
						} else {
							// If the error code is anything else we have other problems
							log.Fatal("AWS returned an unexpected error: ", err)
						}
					} else {
						// And catch any non-AWS errors:
						log.Fatal("Problem removing Login Profile: ", err.Error())
					}
				}
				break
			}
		}()

		// Wait for Ctrl-C
		fmt.Println("Press Ctrl-C when finished...")
		runtime.Gosched()
		wg.Wait()

		// Revoke Vault lease to remove AWS account
		err = client.Sys().Revoke(secret.LeaseID)
		if err != nil {
			log.Fatal("Problem revoking Vault lease: ", err)
		}
		log.Info("Vault Lease revoked. AWS account removed.")
	},
}

func init() {
	RootCmd.AddCommand(awsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mysqlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	awsCmd.Flags().StringVarP(&awsRole, "role", "R", "", "Vault AWS Role to generate credentials for")
	awsCmd.Flags().BoolVarP(&awsCreateLoginProfile, "create-login-profile", "L", false, "Create a Login Profile for the AWS account")
}
