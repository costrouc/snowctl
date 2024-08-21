package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type Endpoints interface {
	Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]Endpoint, error)
}

type endpoints struct {
	client *Client
}

type Endpoint struct {
	Name       string         `db:"name"`
	Port       string         `db:"port"`
	PortRange  sql.NullString `db:"port_range"`
	Protocol   string         `db:"protocol"`
	IsPublic   bool           `db:"is_public"`
	IngressUrl string         `db:"ingress_url"`
}

func (s *endpoints) Show(ctx context.Context, id *sdk.SchemaObjectIdentifier) ([]Endpoint, error) {
	rows, err := s.client.SDKClient.GetConn().Query(fmt.Sprintf("SHOW ENDPOINTS IN SERVICE %s", id.FullyQualifiedName()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	endpoints := make([]Endpoint, 0)
	for rows.Next() {
		var endpoint Endpoint

		if err := rows.Scan(&endpoint.Name, &endpoint.Port, &endpoint.PortRange, &endpoint.Protocol, &endpoint.IsPublic, &endpoint.IngressUrl); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}
