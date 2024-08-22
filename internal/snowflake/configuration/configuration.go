package configuration

import (
	"cmp"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	sf "github.com/snowflakedb/gosnowflake"
)

type ConnectionsToml struct {
	Connections map[string]Connection `toml:"connections"`
}

type Connection struct {
	Account       string `toml:"account"`
	User          string `toml:"user"`
	Password      string `toml:"password"`
	Region        string `toml:"region"`
	Host          string `toml:"host"`
	Database      string `toml:"database"`
	Schema        string `toml:"schema"`
	Role          string `toml:"role"`
	Warehouse     string `toml:"warehouse"`
	Authenticator string `toml:"authenticator"`
}

type ConfigToml struct {
	DefaultConnectionName string `toml:"default_connection_name"`
}

func (c *Connection) SnowflakeConfig() *sf.Config {
	var authType sf.AuthType
	switch c.Authenticator {
	case "externalbrowser":
		authType = sf.AuthTypeExternalBrowser
	default:
		authType = sf.AuthTypeSnowflake
	}

	return &sf.Config{
		Account:       c.Account,
		User:          c.User,
		Password:      c.Password,
		Region:        c.Region,
		Host:          c.Host,
		Database:      c.Database,
		Schema:        c.Schema,
		Role:          c.Role,
		Warehouse:     c.Warehouse,
		Authenticator: authType,
	}
}

func getConfigurationDirectories() []string {
	paths := make([]string, 0)

	if os.Getenv("SNOWFLAKE_HOME") == "" {
		paths = append(paths, os.Getenv("SNOWFLAKE_HOME"))
	}

	usr, _ := user.Current()
	homedir := usr.HomeDir

	switch os := runtime.GOOS; os {
	case "darwin":
		paths = append(
			paths,
			filepath.Join(homedir, "Library/Application Support/snowflake"),
		)
	case "windows":
		paths = append(
			paths,
			"%USERPROFILE%\\AppData\\Local\\snowflake",
		)
	default: // assume linux like
		paths = append(
			paths,
			filepath.Join(homedir, ".config/snowflake"),
			filepath.Join(homedir, ".snowflake"),
		)
	}

	return paths
}

func readConnectionsToml(directory string) (*ConnectionsToml, error) {
	path := filepath.Join(directory, "connections.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s file %w", path, err)
	}

	var connectionsToml ConnectionsToml
	_, err = toml.Decode(string(data), &connectionsToml)
	if err != nil {
		return nil, fmt.Errorf("decoding %s from toml %w", path, err)
	}

	for name, connection := range connectionsToml.Connections {
		connection.Account = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_ACCOUNT", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_ACCOUNT"),
			connection.Account,
		)
		connection.User = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_USER", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_USER"),
			connection.User,
		)
		connection.Password = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_PASSWORD", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_PASSWORD"),
			connection.Password,
		)
		connection.Region = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_REGION", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_REGION"),
			connection.Region,
		)
		connection.Host = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_HOST", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_HOST"),
			connection.Host,
		)
		connection.Database = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_DATABASE", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_DATABASE"),
			connection.Database,
		)
		connection.Schema = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_SCHEMA", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_SCHEMA"),
			connection.Schema,
		)
		connection.Role = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_ROLE", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_ROLE"),
			connection.Role,
		)
		connection.Warehouse = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_WAREHOUSE", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_WAREHOUSE"),
			connection.Warehouse,
		)
		connection.Authenticator = cmp.Or(
			os.Getenv(fmt.Sprintf("SNOWFLAKE_CONNECTIONS_%s_AUTHENTICATOR", strings.ToUpper(name))),
			os.Getenv("SNOWFLAKE_AUTHENTICATOR"),
			connection.Authenticator,
		)
	}

	return &connectionsToml, nil
}

func readConfigToml(directory string) (*ConfigToml, error) {
	path := filepath.Join(directory, "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s file %w", path, err)
	}

	var configToml ConfigToml
	_, err = toml.Decode(string(data), &configToml)
	if err != nil {
		return nil, fmt.Errorf("decoding %s from toml %w", path, err)
	}

	configToml.DefaultConnectionName = cmp.Or(
		configToml.DefaultConnectionName,
		os.Getenv("SNOWFLAKE_DEFAULT_CONNECTION_NAME"),
	)

	return &configToml, nil
}

func ReadConfig() (map[string]*Connection, error) {
	paths := getConfigurationDirectories()

	var connections = make(map[string]*Connection, 0)

	for _, path := range paths {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			configToml, err := readConfigToml(path)
			if err != nil {
				return nil, err
			}

			connectionsToml, err := readConnectionsToml(path)
			if err != nil {
				return nil, err
			}

			for name, connection := range connectionsToml.Connections {
				if name == configToml.DefaultConnectionName {
					connections["default"] = &connection
				}
				connections[name] = &connection
			}
			break
		}
	}
	return connections, nil
}
