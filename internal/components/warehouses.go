package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WarehousesView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *WarehousesOptions
}

type WarehousesOptions struct{}

func NewWarehousesView(connectionManager *snowflake.ConnectionManager, opts *WarehousesOptions) *WarehousesView {
	warehouses := &WarehousesView{
		table:             tview.NewTable(),
		connectionManager: connectionManager,
		options:           opts,
	}

	warehouses.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return warehouses
}

func (v *WarehousesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating warehouses data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *WarehousesView) getData(ctx context.Context) (*Table, error) {
	warehouses, err := t.connectionManager.GetClient().SDKClient.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{})
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show warehouses %w", err)
	}

	columns := []string{"Name", "Owner", "Scaling Policy", "Size", "State", "Type"}
	rows := make([][]string, 0)

	for _, warehouse := range warehouses {
		rows = append(rows, []string{
			warehouse.Name,
			warehouse.Owner,
			string(warehouse.ScalingPolicy),
			string(warehouse.Size),
			string(warehouse.State),
			string(warehouse.Type),
		})
	}

	return &Table{
		Title:   "warehouses",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *WarehousesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Use Warehouse",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				cell := t.table.GetCell(r, 0)
				_, err := t.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, fmt.Sprintf("USE WAREHOUSE %s", cell.Text))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.context.Update(ctx)
				return nil
			},
		},
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				warehouseName := t.table.GetCell(r, 0).Text
				warehouse := sdk.NewAccountObjectIdentifier(
					warehouseName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeWarehouse,
							ObjectIdentifier: warehouse,
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
				r, _ := t.table.GetSelection()
				warehouseName := t.table.GetCell(r, 0).Text
				warehouse := sdk.NewAccountObjectIdentifier(warehouseName)

				message := fmt.Sprintf("Drop warehouse %s?", warehouse.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := t.connectionManager.GetClient().SDKClient.Warehouses.Drop(ctx, warehouse, &sdk.DropWarehouseOptions{})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped warehouse %s", warehouse.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop warehouse %s", warehouse.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *WarehousesView) GetRender() tview.Primitive {
	return v.table
}
