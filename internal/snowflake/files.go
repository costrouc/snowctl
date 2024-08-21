package snowflake

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type Files interface {
	Put(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *PutFileOptions) error
}

type files struct {
	client *Client
}

type PutFileOptions struct {
	SourcePath      string
	DestinationPath string
	Overwrite       bool
	AutoCompress    bool
}

func (c *files) Put(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *PutFileOptions) error {
	if opts == nil {
		opts = &PutFileOptions{}
	}

	stmt := fmt.Sprintf("PUT 'file://%s' '@%s'", opts.SourcePath, filepath.Join(id.FullyQualifiedName(), opts.DestinationPath))
	if opts.Overwrite {
		stmt += " OVERWRITE = TRUE"
	} else {
		stmt += " OVERWRITE = FALSE"
	}

	if opts.AutoCompress {
		stmt += " AUTO_COMPRESS = TRUE"
	} else {
		stmt += " AUTO_COMPRESS = FALSE"
	}

	_, err := c.client.SDKClient.GetConn().Query(stmt)
	return err
}
