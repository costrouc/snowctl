package snowflake

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ServiceContainers interface {
	Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]ServiceContainer, error)
}

type servicecontainers struct {
	client *Client
}

type ServiceContainer struct {
	DatabaseName  string `db:"database_name"`
	SchemaName    string `db:"schema_name"`
	ServiceName   string `db:"service_name"`
	InstanceId    int    `db:"instance_id"`
	ContainerName string `db:"container_name"`
	Status        string `db:"status"`
	Message       string `db:"message"`
	ImageName     string `db:"image_name"`
	ImageDigest   string `db:"image_digest"`
	RestartCount  int    `db:"restart_count"`
	StartTime     string `db:"start_time"`
}

// https://docs.snowflake.com/en/sql-reference/sql/show-service-containers-in-service
func (s *servicecontainers) Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]ServiceContainer, error) {
	rows, err := s.client.SDKClient.GetConn().Query(fmt.Sprintf("SHOW SERVICE CONTAINERS IN SERVICE %s", id.FullyQualifiedName()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serviceContainers := make([]ServiceContainer, 0)
	for rows.Next() {
		var serviceContainer ServiceContainer

		if err := rows.Scan(&serviceContainer.DatabaseName, &serviceContainer.SchemaName, &serviceContainer.ServiceName, &serviceContainer.InstanceId, &serviceContainer.ContainerName, &serviceContainer.Status, &serviceContainer.Message, &serviceContainer.ImageName, &serviceContainer.ImageDigest, &serviceContainer.RestartCount, &serviceContainer.StartTime); err != nil {
			return nil, err
		}
		serviceContainers = append(serviceContainers, serviceContainer)
	}

	return serviceContainers, nil
}
