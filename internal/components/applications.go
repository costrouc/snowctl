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

type ApplicationsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ApplicationsOptions
}

type ApplicationsOptions struct{}

func NewApplicationsView(connectionManager *snowflake.ConnectionManager, opts *ApplicationsOptions) *ApplicationsView {
	applications := &ApplicationsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	applications.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return applications
}

func (v *ApplicationsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ApplicationsView) getData(ctx context.Context) (*Table, error) {
	applications, err := t.connectionManager.GetClient().SDKClient.Applications.Show(ctx, sdk.NewShowApplicationRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show applications %w", err)
	}

	columns := []string{"Name", "Version", "Patch", "Owner"}
	rows := make([][]string, 0)

	for _, application := range applications {
		rows = append(rows, []string{
			application.Name,
			application.Version,
			strconv.Itoa(application.Patch),
			application.Owner,
		})
	}

	return &Table{
		Title:   "applications",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *ApplicationsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				applicationName := t.table.GetCell(r, 0).Text
				application := sdk.NewAccountObjectIdentifier(
					applicationName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeApplication,
							ObjectIdentifier: application,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ApplicationsView) GetRender() tview.Primitive {
	return v.table
}
