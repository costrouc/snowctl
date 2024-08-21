package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ViewsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ViewsOptions
}

type ViewsOptions struct{}

func NewViewsView(connectionManager *snowflake.ConnectionManager, opts *ViewsOptions) *ViewsView {
	views := &ViewsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	views.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return views
}

func (v *ViewsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating views data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ViewsView) getData(ctx context.Context) (*Table, error) {
	views, err := t.connectionManager.GetClient().SDKClient.Views.Show(ctx, sdk.NewShowViewRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show views %w", err)
	}

	columns := []string{"Name", "Schema Name", "Database Name", "Owner", "Kind"}
	rows := make([][]string, 0)

	for _, view := range views {
		rows = append(rows, []string{
			view.Name,
			view.SchemaName,
			view.DatabaseName,
			view.Owner,
			view.Kind,
		})
	}

	return &Table{
		Title:   "views",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ViewsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				viewName := v.table.GetCell(r, 2).Text
				view := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, viewName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeView,
							ObjectIdentifier: view,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ViewsView) GetRender() tview.Primitive {
	return v.table
}
