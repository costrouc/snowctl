package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
)

type ListingsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ListingsOptions
}

type ListingsOptions struct {
}

func NewListingsView(connectionManager *snowflake.ConnectionManager, opts *ListingsOptions) *ListingsView {
	listings := &ListingsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	listings.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return listings
}

func (v *ListingsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating listings data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *ListingsView) getData(ctx context.Context) (*Table, error) {
	listings, err := v.connectionManager.GetClient().Listings.Show(ctx)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show listings %w", err)
	}

	columns := []string{"Name", "Global Name", "State", "Title", "Owner", "Profile"}
	rows := make([][]string, 0)

	for _, listing := range listings {
		rows = append(rows, []string{
			listing.Name,
			listing.GlobalName,
			listing.State,
			listing.Title,
			listing.Owner,
			listing.Profile,
		})
	}

	return &Table{
		Title:   "listings",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ListingsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Open Listing",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 1)
				url := fmt.Sprintf(
					"https://app.snowflake.com/%s/%s/#/data/provider-studio/provider/listing/%s",
					strings.ToLower(applicationState.context.OrganizationName),
					strings.ToLower(applicationState.context.AccountName),
					cell.Text,
				)
				err := browser.OpenURL(url)
				if err != nil {
					applicationState.status.SetError(err)
					return event
				}
				applicationState.status.SetMessage("Opened Browser to Listing")
				return event
			},
		},
		{
			Description: "Public Listing",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				cell := v.table.GetCell(r, 1)
				err := browser.OpenURL(fmt.Sprintf("https://app.snowflake.com/marketplace/listing/%s", cell.Text))
				if err != nil {
					applicationState.status.SetError(err)
					return nil
				}
				applicationState.status.SetMessage("Opened Browser to Listing")
				return nil
			},
		},
	}
}

func (v *ListingsView) GetRender() tview.Primitive {
	return v.table
}
