package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type VersionsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *VersionsOptions
}

type VersionsOptions struct {
	ApplicationPackage sdk.AccountObjectIdentifier
}

func NewVersionsView(connectionManager *snowflake.ConnectionManager, opts *VersionsOptions) *VersionsView {
	versions := &VersionsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	versions.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return versions
}

func (v *VersionsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating application packages versions data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *VersionsView) getData(ctx context.Context, opts *VersionsOptions) (*Table, error) {
	versions, err := t.connectionManager.GetClient().ApplicationPackageVersions.Show(ctx, opts.ApplicationPackage)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show versions %w", err)
	}

	columns := []string{"Version", "Patch", "Label", "CreatedOn"}
	rows := make([][]string, 0)

	for _, version := range versions {
		rows = append(rows, []string{
			version.Version,
			strconv.Itoa(version.Patch),
			version.Label,
			version.CreatedOn.Time.String(),
		})
	}

	return &Table{
		Title:   fmt.Sprintf("versions([pink]%s[blue])", opts.ApplicationPackage.Name()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *VersionsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	bindings := make([]*KeyBinding, 0)

	return bindings
}

func (v *VersionsView) GetRender() tview.Primitive {
	return v.table
}
