package snowflake

import (
	"fmt"
	"strings"

	sf "github.com/snowflakedb/gosnowflake"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake/configuration"
	"github.com/costrouc/snowctl/internal/snowflake/snowsql"
)

type Client struct {
	SDKClient *sdk.Client

	// Modeled after the SDKClient we create the missing interfaces
	Listings                   Listings
	ImageRepositories          ImageRepositories
	ComputePools               ComputePools
	Services                   Services
	ServiceInstances           ServiceInstances
	ServiceContainers          ServiceContainers
	Endpoints                  Endpoints
	Snapshots                  Snapshots
	Secrets                    Secrets
	Files                      Files
	ReleaseDirectives          ReleaseDirectives
	ApplicationPackageVersions ApplicationPackageVersions
}

type Connection interface {
	SnowflakeConfig() *sf.Config
}

type ConnectionManager struct {
	snowflakeConnections map[string]*configuration.Connection
	snowsqlConnections   map[string]*snowsql.Connection
	currentClient        *Client
}

func NewConnectionManager() (*ConnectionManager, error) {
	var connectionManager ConnectionManager

	snowflakeConnections, err := configuration.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("reading snowflake standard configuration %w", err)
	}
	if snowflakeConnections != nil {
		connectionManager.snowflakeConnections = snowflakeConnections
	}

	snowsqlConnections, err := snowsql.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("reading snowflake snowsql configuration %w", err)
	}
	connectionManager.snowsqlConnections = snowsqlConnections

	return &connectionManager, nil
}

func (cm *ConnectionManager) AvailableClients() map[string]*sf.Config {
	var connections = make(map[string]*sf.Config, 0)

	for name, connection := range cm.snowsqlConnections {
		connections[fmt.Sprintf("snowsql.%s", name)] = connection.SnowflakeConfig()
	}
	for name, connection := range cm.snowflakeConnections {
		connections[fmt.Sprintf("snowflake.%s", name)] = connection.SnowflakeConfig()
	}

	return connections
}

func (cm *ConnectionManager) SetClient(name string) error {
	var connection Connection
	tokens := strings.Split(name, ".")
	if len(tokens) != 2 {
		return fmt.Errorf("expected client name with '.' separator e.g. snowsql.connectionname got %s", name)
	}
	switch tokens[0] {
	case "snowsql":
		connection = cm.snowsqlConnections[tokens[1]]
		if connection == nil {
			return fmt.Errorf("snowsql connection with name %s does not exist", tokens[1])
		}
	case "snowflake":
		connection = cm.snowflakeConnections[tokens[1]]
		if connection == nil {
			return fmt.Errorf("snowflake connection with name %s does not exist", tokens[1])
		}
	default:
		return fmt.Errorf("unhandled connection base type %s", tokens[0])
	}

	sdkClient, err := sdk.NewClient(connection.SnowflakeConfig())
	if err != nil {
		return err
	}

	client := &Client{
		SDKClient: sdkClient,
	}
	client.initialize()

	cm.currentClient = client

	return nil
}

func (cm *ConnectionManager) GetClient() *Client {
	return cm.currentClient
}

func (c *Client) initialize() {
	c.Listings = &listings{client: c}
	c.ImageRepositories = &imagerepositories{client: c}
	c.ComputePools = &computepools{client: c}
	c.Services = &services{client: c}
	c.ServiceInstances = &serviceinstances{client: c}
	c.ServiceContainers = &servicecontainers{client: c}
	c.Endpoints = &endpoints{client: c}
	c.Snapshots = &snapshots{client: c}
	c.Secrets = &secrets{client: c}
	c.Files = &files{client: c}
	c.ReleaseDirectives = &releasedirectives{client: c}
	c.ApplicationPackageVersions = &applicationpackageversions{client: c}
}

func (c *Client) Close() {
	c.SDKClient.Close()
}
