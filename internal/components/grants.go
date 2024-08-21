package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type GrantsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *GrantsOptions
}

type GrantsOptions struct {
	ObjectType       sdk.ObjectType
	ObjectIdentifier sdk.ObjectIdentifier
}

func NewGrantsView(connectionManager *snowflake.ConnectionManager, opts *GrantsOptions) *GrantsView {
	grants := &GrantsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	grants.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return grants
}

func (v *GrantsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating service instances data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *GrantsView) getData(ctx context.Context, opts *GrantsOptions) (*Table, error) {
	title := fmt.Sprintf("grants([pink]%s[blue])", opts.ObjectIdentifier.FullyQualifiedName())

	var snowflakeOpts sdk.ShowGrantOptions
	switch opts.ObjectType {
	case sdk.ObjectTypeRole:
		snowflakeOpts = sdk.ShowGrantOptions{
			To: &sdk.ShowGrantsTo{
				Role: sdk.NewAccountObjectIdentifier(opts.ObjectIdentifier.Name()),
			},
		}
	default:
		snowflakeOpts = sdk.ShowGrantOptions{
			On: &sdk.ShowGrantsOn{
				Object: &sdk.Object{
					ObjectType: opts.ObjectType,
					Name:       opts.ObjectIdentifier,
				},
			},
		}
	}

	grants, err := t.connectionManager.GetClient().SDKClient.Grants.Show(ctx, &snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show grants instances %w", err)
	}

	columns := []string{"Grant On", "Privilege", "Grant To", "Name"}
	rows := make([][]string, 0)

	for _, grant := range grants {
		rows = append(rows, []string{
			string(grant.GrantedOn),
			grant.Privilege,
			string(grant.GrantedTo),
			grant.GranteeName.FullyQualifiedName(),
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *GrantsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}

func (v *GrantsView) GetRender() tview.Primitive {
	return v.table
}
