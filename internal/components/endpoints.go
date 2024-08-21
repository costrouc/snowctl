package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
)

type EndpointsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *EndpointsOptions
}

type EndpointsOptions struct {
	Service *sdk.SchemaObjectIdentifier
}

func NewEndpointsView(connectionManager *snowflake.ConnectionManager, opts *EndpointsOptions) *EndpointsView {
	endpoints := &EndpointsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	endpoints.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return endpoints
}

func (v *EndpointsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating endpoints data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *EndpointsView) getData(ctx context.Context, opts *EndpointsOptions) (*Table, error) {
	tables, err := t.connectionManager.GetClient().Endpoints.Show(ctx, opts.Service)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show endpoints %w", err)
	}

	columns := []string{"Name", "Port", "Is Public", "Ingress URL"}
	rows := make([][]string, 0)

	for _, table := range tables {
		rows = append(rows, []string{
			table.Name,
			table.Port,
			strconv.FormatBool(table.IsPublic),
			table.IngressUrl,
		})
	}

	return &Table{
		Title:   fmt.Sprintf("endpoints([pink]%s[blue])", opts.Service.FullyQualifiedName()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *EndpointsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Open Endpoint",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 3)
				url := fmt.Sprintf("https://%s", cell.Text)

				err := browser.OpenURL(url)
				if err != nil {
					applicationState.status.SetError(err)
					return event
				}
				applicationState.status.SetMessage("Opened Browser to Service Endpoint")
				return event
			},
		},
	}
}

func (v *EndpointsView) GetRender() tview.Primitive {
	return v.table
}
