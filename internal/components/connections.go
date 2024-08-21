package components

import (
	"context"
	"fmt"

	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ConnectionsView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *ConnectionsOptions
}

type ConnectionsOptions struct {
}

func NewConnectionsView(connectionManager *snowflake.ConnectionManager, opts *ConnectionsOptions) *ConnectionsView {
	connections := &ConnectionsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	connections.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return connections
}

func (v *ConnectionsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ConnectionsView) getData(ctx context.Context) (*Table, error) {
	columns := []string{"Connection", "Account", "Region", "User", "Role"}
	rows := make([][]string, 0)

	for name, connection := range t.connectionManager.AvailableClients() {
		rows = append(rows, []string{
			name,
			connection.Account,
			connection.Region,
			connection.User,
			connection.Role,
		})
	}

	return &Table{
		Title:   "connections",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *ConnectionsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Use Connection",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				cell := t.table.GetCell(r, 0)
				err := t.connectionManager.SetClient(cell.Text)
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.context.Update(ctx)
				return nil
			},
		},
	}
}

func (v *ConnectionsView) GetRender() tview.Primitive {
	return v.table
}
