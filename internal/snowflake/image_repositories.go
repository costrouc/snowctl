package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ImageRepositories interface {
	Create(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *CreateImageRepositoryOptions) error
	Show(ctx context.Context, opts *ShowImageRepositoryOptions) ([]ImageRepository, error)
	ShowImages(ctx context.Context, id sdk.SchemaObjectIdentifier) ([]Image, error)
	Drop(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *DropImageRepositoryOptions) error
}

type imagerepositories struct {
	client *Client
}

type CreateImageRepositoryOptions struct {
	IfNotExists bool
}

func (c *imagerepositories) Create(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *CreateImageRepositoryOptions) error {
	if opts == nil {
		opts = &CreateImageRepositoryOptions{}
	}

	createListingTemplate := fmt.Sprintf("CREATE IMAGE REPOSITORY {{ if .IfNotExists }}IF NOT EXISTS{{ end }} %s.%s.%s", id.DatabaseName(), id.SchemaName(), id.Name())
	stmt := templateToQuery(createListingTemplate, opts)
	_, err := c.client.SDKClient.GetConn().Exec(stmt)
	return err
}

type ImageRepository struct {
	CreatedOn     sql.NullTime `db:"created_on"`
	Name          string       `db:"name"`
	DatabaseName  string       `db:"database_name"`
	SchemaName    string       `db:"schema_name"`
	RepositoryURL string       `db:"repository_url"`
	Owner         string       `db:"owner"`
	OwnerRoleType string       `db:"owner_role_type"`
	Comment       string       `db:"comment"`
}

type ShowImageRepositoryOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func (c *imagerepositories) Show(ctx context.Context, opts *ShowImageRepositoryOptions) ([]ImageRepository, error) {
	query := "SHOW IMAGE REPOSITORIES"
	if opts.Database != nil {
		query = fmt.Sprintf("SHOW IMAGE REPOSITORIES IN DATABASE %s", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		query = fmt.Sprintf("SHOW IMAGE REPOSITORIES IN SCHEMA %s", opts.Schema.FullyQualifiedName())
	}

	rows, err := c.client.SDKClient.GetConn().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ImageRepository
	for rows.Next() {
		var r ImageRepository
		err = rows.Scan(
			&r.CreatedOn,
			&r.Name,
			&r.DatabaseName,
			&r.SchemaName,
			&r.RepositoryURL,
			&r.Owner,
			&r.OwnerRoleType,
			&r.Comment,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

type Image struct {
	CreatedOn sql.NullTime `db:"created_on"`
	ImageName string       `db:"image_name"`
	Tags      string       `db:"tags"`
	Digest    string       `db:"digest"`
	ImagePath string       `db:"image_path"`
}

func (c *imagerepositories) ShowImages(ctx context.Context, id sdk.SchemaObjectIdentifier) ([]Image, error) {
	stmt := fmt.Sprintf("SHOW IMAGES IN IMAGE REPOSITORY %s;", id.FullyQualifiedName())
	rows, err := c.client.SDKClient.GetConn().QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Image
	for rows.Next() {
		var r Image
		err = rows.Scan(
			&r.CreatedOn,
			&r.ImageName,
			&r.Tags,
			&r.Digest,
			&r.ImagePath,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

type DropImageRepositoryOptions struct {
	IfExists bool
}

func (c *imagerepositories) Drop(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *DropImageRepositoryOptions) error {
	if opts == nil {
		opts = &DropImageRepositoryOptions{}
	}

	dropListingTemplate := fmt.Sprintf("DROP IMAGE REPOSITORY {{ if .IfExists }}IF EXISTS{{ end }} %s.%s.%s", id.DatabaseName(), id.SchemaName(), id.Name())
	stmt := templateToQuery(dropListingTemplate, opts)
	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}
