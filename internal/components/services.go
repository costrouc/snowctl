package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ServicesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ServicesOptions
}

type ServicesOptions struct {
	ComputePool *sdk.AccountObjectIdentifier
	Database    *sdk.AccountObjectIdentifier
	Schema      *sdk.DatabaseObjectIdentifier
}

func NewServicesView(connectionManager *snowflake.ConnectionManager, opts *ServicesOptions) *ServicesView {
	services := &ServicesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	services.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return services
}

func (v *ServicesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating services data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ServicesView) getData(ctx context.Context, opts *ServicesOptions) (*Table, error) {
	services, err := t.connectionManager.GetClient().Services.Show(ctx, &snowflake.ShowServiceOptions{
		ComputePool: opts.ComputePool,
		Database:    opts.Database,
		Schema:      opts.Schema,
	})
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show services %w", err)
	}

	title := "services"
	if opts.ComputePool != nil {
		title = fmt.Sprintf("services([pink]%s[blue])", opts.ComputePool.FullyQualifiedName())
	}
	if opts.Database != nil {
		title = fmt.Sprintf("services([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		title = fmt.Sprintf("services([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}

	columns := []string{"Database", "Schema", "Name", "Compute Pool", "DNS Name"}
	rows := make([][]string, 0)

	for _, service := range services {
		rows = append(rows, []string{
			service.DatabaseName,
			service.SchemaName,
			service.Name,
			service.ComputePool,
			service.DNSName,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ServicesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Endpoints",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				database := v.table.GetCell(r, 0).Text
				schema := v.table.GetCell(r, 1).Text
				name := v.table.GetCell(r, 2).Text

				service := sdk.NewSchemaObjectIdentifier(database, schema, name)

				applicationState.Push(
					ctx,
					NewEndpointsView(
						applicationState.ConnectionManager,
						&EndpointsOptions{
							Service: &service,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Containers",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				database := v.table.GetCell(r, 0).Text
				schema := v.table.GetCell(r, 1).Text
				name := v.table.GetCell(r, 2).Text

				service := sdk.NewSchemaObjectIdentifier(database, schema, name)

				applicationState.Push(
					ctx,
					NewServiceContainersView(
						applicationState.ConnectionManager,
						&ServiceContainersOptions{
							Service: &service,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Instances",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				database := v.table.GetCell(r, 0).Text
				schema := v.table.GetCell(r, 1).Text
				name := v.table.GetCell(r, 2).Text

				service := sdk.NewSchemaObjectIdentifier(database, schema, name)

				applicationState.Push(
					ctx,
					NewServiceInstancesView(
						applicationState.ConnectionManager,
						&ServiceInstancesOptions{
							Service: &service,
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
				serviceName := v.table.GetCell(r, 2).Text
				service := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, serviceName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeService,
							ObjectIdentifier: service,
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
				serviceName := v.table.GetCell(r, 2).Text
				service := sdk.NewSchemaObjectIdentifier(databaseName, schemaName, serviceName)

				message := fmt.Sprintf("Drop service %s?", service.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().Services.Drop(ctx, service, &snowflake.DropServiceOptions{})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped service %s", service.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop service %s", service.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *ServicesView) GetRender() tview.Primitive {
	return v.table
}
