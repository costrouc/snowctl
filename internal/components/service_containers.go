package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ServiceContainersView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ServiceContainersOptions
}

type ServiceContainersOptions struct {
	Service *sdk.SchemaObjectIdentifier
}

func NewServiceContainersView(connectionManager *snowflake.ConnectionManager, opts *ServiceContainersOptions) *ServiceContainersView {
	serviceContainers := &ServiceContainersView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	serviceContainers.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return serviceContainers
}

func (v *ServiceContainersView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating service containers data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ServiceContainersView) getData(ctx context.Context, opts *ServiceContainersOptions) (*Table, error) {
	serviceContainers, err := t.connectionManager.GetClient().ServiceContainers.Show(ctx, opts.Service)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show service containers %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Instance Id", "Container", "Restart Count"}
	rows := make([][]string, 0)

	for _, serviceContainer := range serviceContainers {
		rows = append(rows, []string{
			serviceContainer.DatabaseName,
			serviceContainer.SchemaName,
			serviceContainer.ServiceName,
			strconv.Itoa(serviceContainer.InstanceId),
			serviceContainer.ContainerName,
			strconv.Itoa(serviceContainer.RestartCount),
		})
	}

	return &Table{
		Title:   fmt.Sprintf("service containers([pink]%s[blue])", opts.Service.FullyQualifiedName()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ServiceContainersView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Logs",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				database := v.table.GetCell(r, 0).Text
				schema := v.table.GetCell(r, 1).Text
				name := v.table.GetCell(r, 2).Text

				instanceId, err := strconv.Atoi(v.table.GetCell(r, 3).Text)
				if err != nil {
					applicationState.status.SetError(err)
					return event
				}
				containerName := v.table.GetCell(r, 4).Text

				service := sdk.NewSchemaObjectIdentifier(database, schema, name)

				applicationState.Push(
					ctx,
					NewServiceLogsView(
						applicationState.ConnectionManager,
						&ServiceLogsOptions{
							Service:       &service,
							InstanceId:    instanceId,
							ContainerName: containerName,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ServiceContainersView) GetRender() tview.Primitive {
	return v.table
}
