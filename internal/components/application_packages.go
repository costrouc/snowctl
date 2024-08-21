package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ApplicationPackagesView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *ApplicationPackagesOptions
}

type ApplicationPackagesOptions struct{}

func NewApplicationPackagesView(connectionManager *snowflake.ConnectionManager, opts *ApplicationPackagesOptions) *ApplicationPackagesView {
	applicationPackages := &ApplicationPackagesView{
		table:             tview.NewTable(),
		connectionManager: connectionManager,
		options:           opts,
	}

	applicationPackages.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return applicationPackages
}

func (v *ApplicationPackagesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *ApplicationPackagesView) getData(ctx context.Context) (*Table, error) {
	applicationPackages, err := v.connectionManager.GetClient().SDKClient.ApplicationPackages.Show(ctx, sdk.NewShowApplicationPackageRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show application packages %w", err)
	}

	columns := []string{"Name", "Distribution", "Owner"}
	rows := make([][]string, 0)

	for _, applicationPackage := range applicationPackages {
		rows = append(rows, []string{
			applicationPackage.Name,
			applicationPackage.Distribution,
			applicationPackage.Owner,
		})
	}

	return &Table{
		Title:   "application packages",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *ApplicationPackagesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Versions",
			Event:       tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				name := t.table.GetCell(r, 0).Text

				applicationState.Push(
					ctx,
					NewVersionsView(
						applicationState.ConnectionManager,
						&VersionsOptions{
							ApplicationPackage: sdk.NewAccountObjectIdentifier(name),
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Release Directives",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				name := t.table.GetCell(r, 0).Text
				applicationPackage := sdk.NewAccountObjectIdentifier(name)

				applicationState.Push(
					ctx,
					NewReleaseDirectivesView(
						applicationState.ConnectionManager,
						&ReleaseDirectivesOptions{
							ApplicationPackage: &applicationPackage,
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
				r, _ := t.table.GetSelection()
				applicationPackageName := t.table.GetCell(r, 0).Text
				applicationPackage := sdk.NewAccountObjectIdentifier(
					applicationPackageName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeApplicationPackage,
							ObjectIdentifier: applicationPackage,
						},
					),
				)
				return nil
			},
		},
	}
}

func (v *ApplicationPackagesView) GetRender() tview.Primitive {
	return v.table
}
