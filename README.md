# Breakglass

Breakglass is a tool that will make API calls to [Hashicorp Vault](https://vaultproject.io) servers and then retrieve credentials for you.

It's designed to ease the process of getting elevated login credentials for a variety of servers.

It currently supports MySQL servers and SSH Command line access

# Features

  - Grab MySQL passwords for any vault enabled database
  - Grab a one-time-use SSH user and password to get root access to servers
  - Configuration file, so if you do the same command over and over, you don't need to remember a million flags

# Vault Configuration

The tool currently assumes you have mounted your databases and hosts into vault under specific paths.

All mysql databases should be mounted under `/mysql/<hostname>` in vault for example. For more information, please see the [docs](docs/VAULT.md)

# Using

To use breakglass, simply download the binary. Run the command with no arguments to see the possible commands:

```bash
breakglass allows you to get login credentials for a variety of vault backends, such as databases servers, Linux servers (ssh credentials)
and AWS IAM roles

Usage:
  breakglass [command]

Available Commands:
  help        Help about any command
  mysql       Get temporary login credentials for mysql servers
  ssh         Get temporary SSH credentials for Linux serers

Flags:
      --config string      config file (default is $HOME/.breakglass/config.yaml)
      --debug              enable debug logging
      --vault string   vault host to authenticate against
      --port int      port of vault servers to use when authenticating (default 8200)

Use "breakglass [command] --help" for more information about a command.
```

For more help on the subcommands, run `breakglass mysql help`

## Config

breakglass will do its best to try and detect sane defaults for you. However, some options will need to be configured.

They are configurable by either flag (meaning you have to set them every time you run breakglass) or for ease of use you can use a config file.

Place the config file in `$HOME/.breakglass/config.yaml`

An example config file looks like this:

```yaml
username: "lbriggs"
authmethod: "ldap"
vault: "consulserver-1.example.com"
debug: false
```

These options can be modified as follows:

### username:

This should be the username you use to authenticate to LDAP. If it's not set, breakglass will use the username you're currently logged in as

### authmethod:

This is the method you use to authenticate against vault. Currently only LDAP and userpass are supported. LDAP is the default.

### vault:

Specify the path to the vault server you wish to use.

```
$ breakglass mysql --host lbriggs-mysql.exampke.com --vault consulserver-2.example.com
```

_However_ if you're finding yourself using the same vaulthost over and over again, you can set the vault host in the config file, and it will always use this host.

### debug

Debug will enable debug logging for troubleshooting purposes. Ops may ask you to run with the debug option if you're experiencing problems.

## MySQL Credentials

Assuming you've configured breakglass with the config options above, simply run breakglass and specify the MySQL Server you want access to:

```bash
$ breakglass --host lbriggs-mysql.example.com
Your MySQL Credentials are below
 username: read-ldap-f273c0
 password: <redacted>
```

You can then use these credentials to connect to the MySQL server you specified.

## SSH Credentials

Assuming you've configured breakglass with the config options above, simple run breakglass and specify the SSH server you want access to:

```bash
Please enter your password:
Your SSH Credentials are:
 username: breakglass
 password: <redacted>
```

You can then use these credentials to connect to the Linux server you specified.

# Building

See the [docs](docs/BUILDING.md)

# Contributing

Fork the repo in gitlab and send a merge request!

# Caveats

There are currently no tests, and the code is not very [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself).

This was Apptio's first exercise in Go, and pull requests are very welcome.
