package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type ServiceInstancesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ServiceInstancesOptions
}

type ServiceInstancesOptions struct {
	Service *sdk.SchemaObjectIdentifier
}

func NewServiceInstancesView(connectionManager *snowflake.ConnectionManager, opts *ServiceInstancesOptions) *ServiceInstancesView {
	serviceInstances := &ServiceInstancesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	serviceInstances.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return serviceInstances
}

func (v *ServiceInstancesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating service instances data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ServiceInstancesView) getData(ctx context.Context, opts *ServiceInstancesOptions) (*Table, error) {
	serviceInstances, err := t.connectionManager.GetClient().ServiceInstances.Show(ctx, opts.Service)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show service instances %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Instance Id", "Status"}
	rows := make([][]string, 0)

	for _, serviceInstance := range serviceInstances {
		rows = append(rows, []string{
			serviceInstance.DatabaseName,
			serviceInstance.SchemaName,
			serviceInstance.ServiceName,
			strconv.Itoa(serviceInstance.InstanceId),
			serviceInstance.Status,
		})
	}

	return &Table{
		Title:   fmt.Sprintf("service instances([pink]%s[blue])", opts.Service.FullyQualifiedName()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ServiceInstancesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *ServiceInstancesView) GetRender() tview.Primitive {
	return v.table
}
