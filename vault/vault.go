package vault

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	//"github.com/acidlemon/go-dumper"
	"github.com/hashicorp/vault/api"
)

type Params struct {
	Username string
	Password string
	Method   string
	Host     string
	Port     int
}

const (
	apiVersion = "v1"
)

func New(p Params) (*api.Client, error) {

	// create the login URL
	url := fmt.Sprintf("https://%s:%v", p.Host, p.Port)

	log.Debug("Using Vault URL: ", url)

	// vault API config
	config := &api.Config{Address: url}

	// read environment variables
	if err := config.ReadEnvironment(); err != nil {
		log.Warn("Error reading environment variables", err)
	}

	// create a new client
	client, err := api.NewClient(config)

	if err != nil {
		log.Fatal("Error creating vault client", err)
	}

	// set password for auth
	options := map[string]interface{}{
		"password": p.Password,
	}

	// create the login URL
	path := fmt.Sprintf("auth/%s/login/%s", p.Method, p.Username)

	log.Debug("Login path: ", path)

	// retrieve a login token
	secret, err := client.Logical().Write(path, options)

	if err != nil {
		log.Fatal("Error retrieving login token: ", err)
	}

	if secret == nil {
		log.Fatal("No token retrieved during login - check your login method")
	}

	// set the token to be used to the one retrieved upon login
	client.SetToken(secret.Auth.ClientToken)

	// return a vault client!
	return client, nil

}
