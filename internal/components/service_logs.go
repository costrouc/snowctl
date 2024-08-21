package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type ServiceLogsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ServiceLogsOptions
}

type ServiceLogsOptions struct {
	Service       *sdk.SchemaObjectIdentifier
	InstanceId    int
	ContainerName string
}

func NewServiceLogsView(connectionManager *snowflake.ConnectionManager, opts *ServiceLogsOptions) *ServiceLogsView {
	serviceLogs := &ServiceLogsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	serviceLogs.table.SetSelectable(true, false).SetBorder(true)

	return serviceLogs
}

func (v *ServiceLogsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating service logs data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ServiceLogsView) getData(ctx context.Context, opts *ServiceLogsOptions) (*Table, error) {
	var serviceLogs string
	query := fmt.Sprintf("CALL SYSTEM$GET_SERVICE_LOGS('%s', %d, '%s')", opts.Service.FullyQualifiedName(), opts.InstanceId, opts.ContainerName)
	err := t.connectionManager.GetClient().SDKClient.GetConn().QueryRow(query).Scan(&serviceLogs)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show services logs %w", err)
	}

	columns := []string{"Logs"}
	rows := make([][]string, 0)

	for _, line := range strings.Split(serviceLogs, "\n") {
		rows = append(rows, []string{line})
	}

	return &Table{
		Title:   fmt.Sprintf("logs([pink]%s[blue])", opts.Service.FullyQualifiedName()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ServiceLogsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *ServiceLogsView) GetRender() tview.Primitive {
	return v.table
}
