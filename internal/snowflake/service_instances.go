// https://docs.snowflake.com/en/sql-reference/sql/show-service-instances-in-service
// SHOW SERVICE INSTANCES ...

package snowflake

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ServiceInstances interface {
	Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]ServiceInstance, error)
}

type serviceinstances struct {
	client *Client
}

type ServiceInstance struct {
	DatabaseName string `db:"database_name"`
	SchemaName   string `db:"schema_name"`
	ServiceName  string `db:"service_name"`
	InstanceId   int    `db:"instance_id"`
	Status       string `db:"status"`
	SpecDigest   string `db:"spec_digest"`
	CreationTime string `db:"creation_time"`
	StartTime    string `db:"start_time"`
}

// https://docs.snowflake.com/en/sql-reference/sql/show-service-containers-in-service
func (s *serviceinstances) Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]ServiceInstance, error) {
	rows, err := s.client.SDKClient.GetConn().Query(fmt.Sprintf("SHOW SERVICE INSTANCES IN SERVICE %s", id.FullyQualifiedName()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serviceInstances := make([]ServiceInstance, 0)
	for rows.Next() {
		var serviceInstance ServiceInstance

		if err := rows.Scan(&serviceInstance.DatabaseName, &serviceInstance.SchemaName, &serviceInstance.ServiceName, &serviceInstance.InstanceId, &serviceInstance.Status, &serviceInstance.SpecDigest, &serviceInstance.CreationTime, &serviceInstance.StartTime); err != nil {
			return nil, err
		}
		serviceInstances = append(serviceInstances, serviceInstance)
	}

	return serviceInstances, nil
}
