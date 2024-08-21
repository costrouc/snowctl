package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type Snapshots interface {
	Show(ctx context.Context, opts *ShowSnapshotOptions) ([]Snapshot, error)
	Drop(ctx context.Context, opts *DropSnapshotOptions) error
}

type snapshots struct {
	client *Client
}

type Snapshot struct {
	Name          string         `db:"name"`
	State         string         `db:"state"`
	DatabaseName  string         `db:"database_name"`
	SchemaName    string         `db:"schema_name"`
	ServiceName   string         `db:"service_name"`
	VolumeName    string         `db:"volume_name"`
	Instance      int            `db:"instance"`
	Size          string         `db:"size"`
	Comment       sql.NullString `db:"comment"`
	Owner         string         `db:"owner"`
	OwnerRoleType string         `db:"owner_role_type"`
	CreatedOn     string         `db:"created_on"`
	UpdatedOn     string         `db:"updated_on"`
}

type ShowSnapshotOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func (s *snapshots) Show(ctx context.Context, opts *ShowSnapshotOptions) ([]Snapshot, error) {
	query := "SHOW SNAPSHOTS"
	if opts.Database != nil {
		query = fmt.Sprintf("SHOW SNAPSHOTS IN DATABASE %s", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		query = fmt.Sprintf("SHOW SNAPSHOTS IN SCHEMA %s", opts.Schema.FullyQualifiedName())
	}
	rows, err := s.client.SDKClient.GetConn().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make([]Snapshot, 0)
	for rows.Next() {
		var snapshot Snapshot

		if err := rows.Scan(&snapshot.Name, &snapshot.State, &snapshot.DatabaseName, &snapshot.SchemaName, &snapshot.ServiceName, &snapshot.VolumeName, &snapshot.Instance, &snapshot.Size, &snapshot.Comment, &snapshot.Owner, &snapshot.OwnerRoleType, &snapshot.CreatedOn, &snapshot.UpdatedOn); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

type DropSnapshotOptions struct {
	IfExists bool
	Snapshot *sdk.SchemaObjectIdentifier
}

func (s *snapshots) Drop(ctx context.Context, opts *DropSnapshotOptions) error {
	query := fmt.Sprintf("DROP SNAPSHOT %s", opts.Snapshot.FullyQualifiedName())
	if opts.IfExists {
		query = fmt.Sprintf("DROP SNAPSHOT IF EXISTS %s", opts.Snapshot.FullyQualifiedName())
	}
	_, err := s.client.SDKClient.GetConn().ExecContext(ctx, query)
	return err
}
