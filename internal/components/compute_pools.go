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

type ComputePoolsView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *ComputePoolsOptions
}

type ComputePoolsOptions struct{}

func NewComputePoolsView(connectionManager *snowflake.ConnectionManager, opts *ComputePoolsOptions) *ComputePoolsView {
	computePools := &ComputePoolsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	computePools.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return computePools
}

func (v *ComputePoolsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ComputePoolsView) getData(ctx context.Context) (*Table, error) {
	computePools, err := t.connectionManager.GetClient().ComputePools.Show(ctx)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show compute pools %w", err)
	}

	columns := []string{"Name", "Owner", "Instance Family", "State", "Application", "Auto Suspend Secs"}
	rows := make([][]string, 0)

	for _, computePool := range computePools {
		rows = append(rows, []string{
			computePool.Name,
			computePool.Owner,
			computePool.InstanceFamily,
			computePool.State,
			computePool.Application.String,
			strconv.Itoa(computePool.AutoSuspendSecs),
		})
	}

	return &Table{
		Title:   "compute pools",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ComputePoolsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Suspend",
			Event:       tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				name := v.table.GetCell(r, 0).Text
				_, err := v.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, fmt.Sprintf("ALTER COMPUTE POOL %s SUSPEND", name))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.status.SetMessage(fmt.Sprintf("Suspended compute pool %s", name))
				return nil
			},
		},
		{
			Description: "Resume",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				name := v.table.GetCell(r, 0).Text
				_, err := v.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, fmt.Sprintf("ALTER COMPUTE POOL %s RESUME", name))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.status.SetMessage(fmt.Sprintf("Resumed compute pool %s", name))
				return nil
			},
		},
		{
			Description: "Drop",
			Event:       tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				computePoolName := v.table.GetCell(r, 0).Text
				computePool := sdk.NewAccountObjectIdentifier(computePoolName)

				message := fmt.Sprintf("Drop compute pool %s?", computePool.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().ComputePools.Drop(ctx, computePool, &snowflake.DropComputePoolOptions{})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped compute pool %s", computePool.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop compute pool %s", computePool.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
		{
			Description: "Services",
			Event:       tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 0)

				computePool := sdk.NewAccountObjectIdentifier(cell.Text)

				applicationState.Push(
					ctx,
					NewServicesView(
						applicationState.ConnectionManager,
						&ServicesOptions{
							ComputePool: &computePool,
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
				computePoolName := v.table.GetCell(r, 0).Text
				computePool := sdk.NewAccountObjectIdentifier(
					computePoolName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeComputePool,
							ObjectIdentifier: computePool,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ComputePoolsView) GetRender() tview.Primitive {
	return v.table
}
