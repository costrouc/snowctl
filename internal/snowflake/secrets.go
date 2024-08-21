package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type Secrets interface {
	Show(ctx context.Context, opts *ShowSecretsOptions) ([]Secret, error)
	Drop(ctx context.Context, opts *DropSecretsOptions) error
}

type secrets struct {
	client *Client
}

type Secret struct {
	CreatedOn     string         `db:"created_on"`
	Name          string         `db:"name"`
	SchemaName    string         `db:"schema_name"`
	DatabaseName  string         `db:"database_name"`
	Owner         string         `db:"owner"`
	Comment       sql.NullString `db:"comment"`
	SecretType    string         `db:"secret_type"`
	OAuthScopes   sql.NullString `db:"oauth_scopes"`
	OwnerRoleType string         `db:"owner_role_type"`
}

type ShowSecretsOptions struct {
	Database           *sdk.AccountObjectIdentifier
	Schema             *sdk.DatabaseObjectIdentifier
	Application        *sdk.AccountObjectIdentifier
	ApplicationPackage *sdk.AccountObjectIdentifier
}

func (s *secrets) Show(ctx context.Context, opts *ShowSecretsOptions) ([]Secret, error) {
	query := "SHOW SECRETS"
	if opts.Database != nil {
		query = fmt.Sprintf("SHOW SECRETS IN DATABASE %s", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		query = fmt.Sprintf("SHOW SECRETS IN SCHEMA %s", opts.Schema.FullyQualifiedName())
	}
	if opts.Application != nil {
		query = fmt.Sprintf("SHOW SECRETS IN APPLICATION %s", opts.Application.FullyQualifiedName())
	}
	if opts.Application != nil {
		query = fmt.Sprintf("SHOW SECRETS IN APPLICATION PACKAGE %s", opts.ApplicationPackage.FullyQualifiedName())
	}

	rows, err := s.client.SDKClient.GetConn().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	secrets := make([]Secret, 0)
	for rows.Next() {
		var secret Secret

		if err := rows.Scan(&secret.CreatedOn, &secret.Name, &secret.SchemaName, &secret.DatabaseName, &secret.Owner, &secret.Comment, &secret.SecretType, &secret.OAuthScopes, &secret.OwnerRoleType); err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

type DropSecretsOptions struct {
	IfExists bool
	Secret   *sdk.SchemaObjectIdentifier
}

func (s *secrets) Drop(ctx context.Context, opts *DropSecretsOptions) error {
	query := fmt.Sprintf("DROP SECRET %s", opts.Secret.FullyQualifiedName())
	if opts.IfExists {
		query = fmt.Sprintf("DROP SECRET IF EXISTS %s", opts.Secret.FullyQualifiedName())
	}
	_, err := s.client.SDKClient.GetConn().ExecContext(ctx, query)
	return err
}
