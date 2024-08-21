package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ProceduresView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ProceduresOptions
}

type ProceduresOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func NewProceduresView(connectionManager *snowflake.ConnectionManager, opts *ProceduresOptions) *ProceduresView {
	Procedures := &ProceduresView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	Procedures.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return Procedures
}

func (v *ProceduresView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating procedures data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ProceduresView) getData(ctx context.Context, opts *ProceduresOptions) (*Table, error) {
	title := "procedures"
	snowflakeOpts := sdk.NewShowProcedureRequest()
	if opts.Database != nil {
		snowflakeOpts.WithIn(&sdk.In{Database: *opts.Database})
		title = fmt.Sprintf("procedures([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		snowflakeOpts.WithIn(&sdk.In{Schema: *opts.Schema})
		title = fmt.Sprintf("procedures([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}

	procedures, err := t.connectionManager.GetClient().SDKClient.Procedures.Show(ctx, snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show procedures %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Arguments"}
	rows := make([][]string, 0)

	for _, procedure := range procedures {
		rows = append(rows, []string{
			procedure.CatalogName,
			procedure.SchemaName,
			procedure.Name,
			procedure.Arguments,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ProceduresView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				procedureName := v.table.GetCell(r, 2).Text
				procedure := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, procedureName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeProcedure,
							ObjectIdentifier: procedure,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ProceduresView) GetRender() tview.Primitive {
	return v.table
}
