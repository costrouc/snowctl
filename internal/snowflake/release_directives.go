package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ReleaseDirectives interface {
	Show(ctx context.Context, id sdk.AccountObjectIdentifier) ([]ReleaseDirective, error)
}

type releasedirectives struct {
	client *Client
}

type ReleaseDirective struct {
	Name           string         `db:"name"`
	TargetType     sql.NullString `db:"target_type"`
	TargetName     sql.NullString `db:"target_name"`
	CreatedOn      sql.NullTime   `db:"created_on"`
	Version        string         `db:"version"`
	Patch          int            `db:"patch"`
	ModifiedOn     sql.NullTime   `db:"modified_on"`
	ActiveRegions  string         `db:"active_regions"`
	PendingRegions sql.NullString `db:"pending_regions"`
	ReleaseStatus  string         `db:"release_status"`
	DeployedOn     sql.NullTime   `db:"deployed_on"`
}

func (c *releasedirectives) Show(ctx context.Context, id sdk.AccountObjectIdentifier) ([]ReleaseDirective, error) {
	stmt := fmt.Sprintf("SHOW RELEASE DIRECTIVES IN APPLICATION PACKAGE %s;", id.Name())
	rows, err := c.client.SDKClient.GetConn().Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ReleaseDirective
	for rows.Next() {
		var r ReleaseDirective
		err = rows.Scan(
			&r.Name,
			&r.TargetType,
			&r.TargetName,
			&r.CreatedOn,
			&r.Version,
			&r.Patch,
			&r.ModifiedOn,
			&r.ActiveRegions,
			&r.PendingRegions,
			&r.ReleaseStatus,
			&r.DeployedOn,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
