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

type TablesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *TablesOptions
}

type TablesOptions struct{}

func NewTablesView(connectionManager *snowflake.ConnectionManager, opts *TablesOptions) *TablesView {
	tables := &TablesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	tables.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return tables
}

func (v *TablesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating tables data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *TablesView) getData(ctx context.Context) (*Table, error) {
	tables, err := t.connectionManager.GetClient().SDKClient.Tables.Show(ctx, sdk.NewShowTableRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show tables %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Owner", "Kind", "Rows"}
	rows := make([][]string, 0)

	for _, table := range tables {
		rows = append(rows, []string{
			table.DatabaseName,
			table.SchemaName,
			table.Name,
			table.Owner,
			table.Kind,
			strconv.Itoa(table.Rows),
		})
	}

	return &Table{
		Title:   "services",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *TablesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				tableName := v.table.GetCell(r, 2).Text
				table := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, tableName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeTable,
							ObjectIdentifier: table,
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
				tableName := v.table.GetCell(r, 2).Text
				table := sdk.NewSchemaObjectIdentifier(databaseName, schemaName, tableName)

				message := fmt.Sprintf("Drop table %s?", table.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().SDKClient.Tables.Drop(ctx, sdk.NewDropTableRequest(table))
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped table %s", table.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop table %s", table.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *TablesView) GetRender() tview.Primitive {
	return v.table
}
