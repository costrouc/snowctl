package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type NetworkPoliciesView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *NetworkPoliciesOptions
}

type NetworkPoliciesOptions struct {
}

func NewNetworkPoliciesView(connectionManager *snowflake.ConnectionManager, opts *NetworkPoliciesOptions) *NetworkPoliciesView {
	networkPolicies := &NetworkPoliciesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	networkPolicies.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return networkPolicies
}

func (v *NetworkPoliciesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating network rules data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *NetworkPoliciesView) getData(ctx context.Context, opts *NetworkPoliciesOptions) (*Table, error) {
	title := "network policies"
	networkPolicies, err := t.connectionManager.GetClient().SDKClient.NetworkPolicies.Show(ctx, sdk.NewShowNetworkPolicyRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show network policies %w", err)
	}

	columns := []string{"Name"}
	rows := make([][]string, 0)

	for _, networkPolicy := range networkPolicies {
		rows = append(rows, []string{
			networkPolicy.Name,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *NetworkPoliciesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *NetworkPoliciesView) GetRender() tview.Primitive {
	return v.table
}
