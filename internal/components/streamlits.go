package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StreamlitsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *StreamlitsOptions
}

type StreamlitsOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func NewStreamlitsView(connectionManager *snowflake.ConnectionManager, opts *StreamlitsOptions) *StreamlitsView {
	streamlits := &StreamlitsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	streamlits.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return streamlits
}

func (v *StreamlitsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating network rules data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *StreamlitsView) getData(ctx context.Context, opts *StreamlitsOptions) (*Table, error) {
	title := "streamlits"
	snowflakeOpts := sdk.NewShowStreamlitRequest()
	if opts.Database != nil {
		snowflakeOpts.WithIn(&sdk.In{Database: *opts.Database})
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		snowflakeOpts.WithIn(&sdk.In{Schema: *opts.Schema})
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}

	streamlits, err := t.connectionManager.GetClient().SDKClient.Streamlits.Show(ctx, snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show streamlits %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Title"}
	rows := make([][]string, 0)

	for _, streamlit := range streamlits {
		rows = append(rows, []string{
			streamlit.DatabaseName,
			streamlit.SchemaName,
			streamlit.Name,
			streamlit.Title,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *StreamlitsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				streamlitName := v.table.GetCell(r, 2).Text
				streamlit := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, streamlitName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeStreamlit,
							ObjectIdentifier: streamlit,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *StreamlitsView) GetRender() tview.Primitive {
	return v.table
}
