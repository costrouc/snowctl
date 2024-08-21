package snowsql

import (
	"cmp"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	sf "github.com/snowflakedb/gosnowflake"
	"gopkg.in/ini.v1"
)

type Connection struct {
	AccountName   string `ini:"accountname"`
	Region        string `ini:"region"`
	Username      string `ini:"username"`
	Password      string `ini:"password"`
	DBName        string `ini:"dbname"`
	SchemaName    string `ini:"schemaname"`
	WarehouseName string `ini:"warehousename"`
	RoleName      string `ini:"rolename"`
	ProxyHost     string `ini:"proxyhost"`
	ProxyPort     string `ini:"proxyport"`
}

func (c *Connection) SnowflakeConfig() *sf.Config {
	return &sf.Config{
		Account:   c.AccountName,
		Region:    c.Region,
		User:      c.Username,
		Password:  c.Password,
		Database:  c.DBName,
		Schema:    c.SchemaName,
		Warehouse: c.WarehouseName,
		Role:      c.RoleName,
	}
}

func getConfigurationPaths() []string {
	usr, _ := user.Current()
	homedir := usr.HomeDir

	return []string{
		"/etc/snowsql.cnf",
		"/etc/snowflake/snowsql.cnf",
		"/usr/local/etc/snowsql.cnf",
		filepath.Join(homedir, ".snowsql.cnf"),
		filepath.Join(homedir, ".snowsql/config"),
	}
}

func ReadConfig() (map[string]*Connection, error) {
	var connections map[string]*Connection

	paths := getConfigurationPaths()

	for _, path := range paths {
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			connections, err = readConfig(path)
			if err != nil {
				return nil, err
			}
		}
	}
	return connections, nil
}

func readSectionConnection(section *ini.Section) (*Connection, error) {
	var connection Connection
	err := section.MapTo(&connection)
	if err != nil {
		return nil, err
	}

	connection.AccountName = cmp.Or(os.Getenv("SNOWSQL_ACCOUNT"), connection.AccountName)
	connection.Region = cmp.Or(os.Getenv("SNOWSQL_REGION"), connection.Region)
	connection.Username = cmp.Or(os.Getenv("SNOWSQL_USER"), connection.Username)
	connection.Password = cmp.Or(os.Getenv("SNOWSQL_PWD"), connection.Password)
	connection.DBName = cmp.Or(os.Getenv("SNOWSQL_DATABASE"), connection.DBName)
	connection.SchemaName = cmp.Or(os.Getenv("SNOWSQL_SCHEMA"), connection.SchemaName)
	connection.WarehouseName = cmp.Or(os.Getenv("SNOWSQL_WAREHOUSE"), connection.WarehouseName)
	connection.RoleName = cmp.Or(os.Getenv("SNOWSQL_ROLE"), connection.RoleName)
	connection.ProxyHost = cmp.Or(os.Getenv("PROXY_HOST"), connection.ProxyHost)
	connection.ProxyPort = cmp.Or(os.Getenv("PROXY_PORT"), connection.ProxyPort)

	return &connection, nil
}

func readConfig(path string) (map[string]*Connection, error) {
	connections := make(map[string]*Connection, 0)

	cfg, err := ini.InsensitiveLoad(path)
	if err != nil {
		return nil, fmt.Errorf("reading init file %s %w", path, err)
	}

	sections := []string{""}
	section, err := cfg.GetSection("connections")
	if err != nil {
		return connections, nil
	}

	ignoredNames := map[string]bool{
		"accountname":   true,
		"region":        true,
		"username":      true,
		"password":      true,
		"dbname":        true,
		"schemaname":    true,
		"warehousename": true,
		"rolename":      true,
		"proxyhost":     true,
		"proxyport":     true,
	}
	for _, key := range section.KeyStrings() {
		if !ignoredNames[key] {
			sections = append(sections, key)
		}
	}

	for _, name := range sections {
		sectionName := "connections"
		connectionName := "default"
		if name != "" {
			sectionName = fmt.Sprintf("connections.%s", name)
			connectionName = name
		}

		section, err = cfg.GetSection(sectionName)
		if err != nil {
			return connections, nil
		}

		connection, err := readSectionConnection(section)
		if err != nil {
			return nil, fmt.Errorf("reading section %s %w", sectionName, err)
		}
		connections[connectionName] = connection
	}

	return connections, nil
}
