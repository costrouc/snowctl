package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type ReleaseDirectivesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *ReleaseDirectivesOptions
}

type ReleaseDirectivesOptions struct {
	ApplicationPackage *sdk.AccountObjectIdentifier
}

func NewReleaseDirectivesView(connectionManager *snowflake.ConnectionManager, opts *ReleaseDirectivesOptions) *ReleaseDirectivesView {
	releaseDirectives := &ReleaseDirectivesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	releaseDirectives.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return releaseDirectives
}

func (v *ReleaseDirectivesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating network rules data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *ReleaseDirectivesView) getData(ctx context.Context, opts *ReleaseDirectivesOptions) (*Table, error) {
	releaseDirectives, err := t.connectionManager.GetClient().ReleaseDirectives.Show(ctx, *opts.ApplicationPackage)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show release directives %w", err)
	}

	columns := []string{"Name", "Version", "Patch", "Target Type"}
	rows := make([][]string, 0)

	for _, releaseDirective := range releaseDirectives {
		rows = append(rows, []string{
			releaseDirective.Name,
			releaseDirective.Version,
			strconv.Itoa(releaseDirective.Patch),
			releaseDirective.TargetType.String,
		})
	}

	return &Table{
		Title:   fmt.Sprintf("release directives([pink]%s[blue])", opts.ApplicationPackage.FullyQualifiedName()),
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *ReleaseDirectivesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *ReleaseDirectivesView) GetRender() tview.Primitive {
	return v.table
}
