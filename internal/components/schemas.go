package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SchemasView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *SchemasOptions
}

type SchemasOptions struct {
	Database *string
}

func NewSchemasView(connectionManager *snowflake.ConnectionManager, opts *SchemasOptions) *SchemasView {
	schemas := &SchemasView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	schemas.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return schemas
}

func (v *SchemasView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating schemas data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *SchemasView) getData(ctx context.Context, opts *SchemasOptions) (*Table, error) {
	sdkOpts := &sdk.ShowSchemaOptions{}
	title := "schemas"
	if opts.Database != nil {
		sdkOpts.In = &sdk.SchemaIn{Database: sdk.Bool(true), Name: sdk.NewAccountObjectIdentifier(*opts.Database)}
		title = fmt.Sprintf("schemas([pink]%s[blue])", *opts.Database)
	}

	schemas, err := v.connectionManager.GetClient().SDKClient.Schemas.Show(ctx, sdkOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show schemas %w", err)
	}

	columns := []string{"Database", "Schema", "Owner"}
	rows := make([][]string, 0)

	for _, schema := range schemas {
		rows = append(rows, []string{
			schema.DatabaseName,
			schema.Name,
			schema.Owner,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *SchemasView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Use Schema",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()

				database := v.table.GetCell(r, 0).Text
				schema := v.table.GetCell(r, 1).Text
				_, err := v.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, fmt.Sprintf("USE SCHEMA %s.%s", database, schema))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.context.Update(ctx)
				return nil
			},
		},
		{
			Description: "Stages",
			Event:       tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text

				applicationState.Push(
					ctx,
					NewStagesView(
						applicationState.ConnectionManager,
						&StagesOptions{
							Database: sdk.String(databaseName),
							Schema:   sdk.String(schemaName),
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Services",
			Event:       tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewServicesView(
						applicationState.ConnectionManager,
						&ServicesOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Snapshots",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewSnapshotsView(
						applicationState.ConnectionManager,
						&SnapshotsOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Image Repositories",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewImageRepositoriesView(
						applicationState.ConnectionManager,
						&ImageRepositoriesOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Procedures",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewProceduresView(
						applicationState.ConnectionManager,
						&ProceduresOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Streamlits",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewStreamlitsView(
						applicationState.ConnectionManager,
						&StreamlitsOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Secrets",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewSecretsView(
						applicationState.ConnectionManager,
						&SecretsOptions{
							Schema: &schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeSchema,
							ObjectIdentifier: schema,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Drop",
			Event:       tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				schema := sdk.NewDatabaseObjectIdentifier(
					databaseName, schemaName,
				)

				message := fmt.Sprintf("Drop schema %s?", schema.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().SDKClient.Schemas.Drop(
							ctx,
							schema,
							&sdk.DropSchemaOptions{},
						)
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped schema %s", schema.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop schema %s", schema.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *SchemasView) GetRender() tview.Primitive {
	return v.table
}
