package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DatabasesView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *DatabasesOptions
}

type DatabasesOptions struct{}

func NewDatabasesView(connectionManager *snowflake.ConnectionManager, opts *DatabasesOptions) *DatabasesView {
	databases := &DatabasesView{
		table:             tview.NewTable(),
		connectionManager: connectionManager,
		options:           opts,
	}

	databases.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return databases
}

func (v *DatabasesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *DatabasesView) getData(ctx context.Context) (*Table, error) {
	databases, err := v.connectionManager.GetClient().SDKClient.Databases.Show(ctx, &sdk.ShowDatabasesOptions{})
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show databases %w", err)
	}

	columns := []string{"Name", "Owner", "Kind", "Comment"}
	rows := make([][]string, 0)

	for _, database := range databases {
		rows = append(rows, []string{
			database.Name,
			database.Owner,
			database.Kind,
			database.Comment,
		})
	}

	return &Table{
		Title:   "databases",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *DatabasesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Use Database",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 0)
				_, err := v.connectionManager.GetClient().SDKClient.GetConn().ExecContext(ctx, fmt.Sprintf("USE DATABASE %s", cell.Text))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.context.Update(ctx)
				return nil
			},
		},
		{
			Description: "Select Database",
			Event:       tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				name := v.table.GetCell(r, 0).Text

				applicationState.Push(
					ctx,
					NewSchemasView(
						applicationState.ConnectionManager,
						&SchemasOptions{
							Database: sdk.String(name),
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Network Rules",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				database := sdk.NewAccountObjectIdentifier(
					databaseName,
				)

				applicationState.Push(
					ctx,
					NewNetworkRulesView(
						applicationState.ConnectionManager,
						&NetworkRulesOptions{
							Database: &database,
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
				database := sdk.NewAccountObjectIdentifier(
					databaseName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeDatabase,
							ObjectIdentifier: database,
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
				database := sdk.NewAccountObjectIdentifier(databaseName)

				message := fmt.Sprintf("Drop database %s?", database.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().SDKClient.Databases.Drop(ctx, database, &sdk.DropDatabaseOptions{})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped database %s", database.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop database %s", database.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *DatabasesView) GetRender() tview.Primitive {
	return v.table
}
