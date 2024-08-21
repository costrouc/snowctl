package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type NetworkRulesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *NetworkRulesOptions
}

type NetworkRulesOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func NewNetworkRulesView(connectionManager *snowflake.ConnectionManager, opts *NetworkRulesOptions) *NetworkRulesView {
	networkRules := &NetworkRulesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	networkRules.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return networkRules
}

func (v *NetworkRulesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating network rules data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *NetworkRulesView) getData(ctx context.Context, opts *NetworkRulesOptions) (*Table, error) {
	title := "network rules"
	snowflakeOpts := sdk.NewShowNetworkRuleRequest()
	if opts.Database != nil {
		snowflakeOpts.WithIn(&sdk.In{Database: *opts.Database})
		title = fmt.Sprintf("network rules([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		snowflakeOpts.WithIn(&sdk.In{Schema: *opts.Schema})
		title = fmt.Sprintf("network rules([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}

	networkRules, err := t.connectionManager.GetClient().SDKClient.NetworkRules.Show(ctx, snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show service instances %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Type", "Mode"}
	rows := make([][]string, 0)

	for _, networkRule := range networkRules {
		rows = append(rows, []string{
			networkRule.DatabaseName,
			networkRule.SchemaName,
			networkRule.Name,
			string(networkRule.Type),
			string(networkRule.Mode),
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *NetworkRulesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *NetworkRulesView) GetRender() tview.Primitive {
	return v.table
}
