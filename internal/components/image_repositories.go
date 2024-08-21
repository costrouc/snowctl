package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ImageRepositoriesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ImageRepositoriesOptions
}

type ImageRepositoriesOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func NewImageRepositoriesView(connectionManager *snowflake.ConnectionManager, opts *ImageRepositoriesOptions) *ImageRepositoriesView {
	imageRepositories := &ImageRepositoriesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	imageRepositories.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return imageRepositories
}

func (v *ImageRepositoriesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating service instances data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ImageRepositoriesView) getData(ctx context.Context, opts *ImageRepositoriesOptions) (*Table, error) {
	title := "image repositories"
	snowflakeOpts := snowflake.ShowImageRepositoryOptions{}
	if opts.Database != nil {
		snowflakeOpts.Database = opts.Database
		title = fmt.Sprintf("image repositories([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		snowflakeOpts.Schema = opts.Schema
		title = fmt.Sprintf("image repositories([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}

	imagerepositories, err := t.connectionManager.GetClient().ImageRepositories.Show(ctx, &snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show service instances %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Repository URL"}
	rows := make([][]string, 0)

	for _, serviceInstance := range imagerepositories {
		rows = append(rows, []string{
			serviceInstance.DatabaseName,
			serviceInstance.SchemaName,
			serviceInstance.Name,
			serviceInstance.RepositoryURL,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ImageRepositoriesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				name := v.table.GetCell(r, 2).Text
				imageRepository := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, name,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeImageRepository,
							ObjectIdentifier: imageRepository,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ImageRepositoriesView) GetRender() tview.Primitive {
	return v.table
}
