package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ApplicationPackageVersions interface {
	Show(ctx context.Context, id sdk.AccountObjectIdentifier) ([]ApplicationPackageVersion, error)
}

type applicationpackageversions struct {
	client *Client
}

type ApplicationPackageVersion struct {
	Version      string         `db:"version"`
	Patch        int            `db:"patch"`
	Label        string         `db:"label"`
	Comment      sql.NullString `db:"comment"`
	CreatedOn    sql.NullTime   `db:"created_on"`
	DroppedOn    sql.NullTime   `db:"dropped_on"`
	LogLevel     string         `db:"log_level"`
	TraceLevel   string         `db:"trace_level"`
	State        string         `db:"state"`
	ReviewStatus string         `db:"review_status"`
}

type ApplicationPackageManifest struct {
}

func (c *applicationpackageversions) Show(ctx context.Context, id sdk.AccountObjectIdentifier) ([]ApplicationPackageVersion, error) {
	stmt := fmt.Sprintf("SHOW VERSIONS IN APPLICATION PACKAGE %s", id.Name())
	rows, err := c.client.SDKClient.GetConn().QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ApplicationPackageVersion
	for rows.Next() {
		var r ApplicationPackageVersion
		err = rows.Scan(
			&r.Version,
			&r.Patch,
			&r.Label,
			&r.Comment,
			&r.CreatedOn,
			&r.DroppedOn,
			&r.LogLevel,
			&r.TraceLevel,
			&r.State,
			&r.ReviewStatus,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
