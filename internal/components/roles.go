package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type RolesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *RolesOptions
}

type RolesOptions struct{}

func NewRolesView(connectionManager *snowflake.ConnectionManager, opts *RolesOptions) *RolesView {
	roles := &RolesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
	}

	roles.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return roles
}

func (v *RolesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating roles data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *RolesView) getData(ctx context.Context) (*Table, error) {
	roles, err := t.connectionManager.GetClient().SDKClient.Roles.Show(ctx, sdk.NewShowRoleRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show roles %w", err)
	}

	columns := []string{"Name", "Owner"}
	rows := make([][]string, 0)

	for _, role := range roles {
		rows = append(rows, []string{
			role.Name,
			role.Owner,
		})
	}

	return &Table{
		Title:   "roles",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *RolesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Use Role",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 0)
				query := fmt.Sprintf("USE ROLE %s", cell.Text)
				_, err := v.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, query)
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
				r, _ := v.table.GetSelection()
				roleName := v.table.GetCell(r, 0).Text
				role := sdk.NewAccountObjectIdentifier(roleName)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeRole,
							ObjectIdentifier: role,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *RolesView) GetRender() tview.Primitive {
	return v.table
}
