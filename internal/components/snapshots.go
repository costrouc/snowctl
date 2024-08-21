package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SnapshotsView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *SnapshotsOptions
}

type SnapshotsOptions struct {
	Database *sdk.AccountObjectIdentifier
	Schema   *sdk.DatabaseObjectIdentifier
}

func NewSnapshotsView(connectionManager *snowflake.ConnectionManager, opts *SnapshotsOptions) *SnapshotsView {
	snapshots := &SnapshotsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	snapshots.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return snapshots
}

func (v *SnapshotsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating stages data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *SnapshotsView) getData(ctx context.Context, opts *SnapshotsOptions) (*Table, error) {
	title := "snapshots"
	snowflakeOpts := snowflake.ShowSnapshotOptions{}
	if opts.Database != nil {
		title = fmt.Sprintf("snapshots([pink]%s[blue])", opts.Database.FullyQualifiedName())
		snowflakeOpts.Database = opts.Database
	}
	if opts.Schema != nil {
		title = fmt.Sprintf("stages([pink]%s[blue])", opts.Schema.FullyQualifiedName())
		snowflakeOpts.Schema = opts.Schema
	}

	snapshots, err := t.connectionManager.GetClient().Snapshots.Show(ctx, &snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show stages %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Service", "Volume", "Size", "State"}
	rows := make([][]string, 0)

	for _, snapshot := range snapshots {
		rows = append(rows, []string{
			snapshot.DatabaseName,
			snapshot.SchemaName,
			snapshot.Name,
			snapshot.ServiceName,
			snapshot.VolumeName,
			snapshot.Size,
			snapshot.State,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *SnapshotsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Drop",
			Event:       tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				databaseName := t.table.GetCell(r, 0).Text
				schemaName := t.table.GetCell(r, 1).Text
				snapshotName := t.table.GetCell(r, 2).Text
				snapshot := sdk.NewSchemaObjectIdentifier(databaseName, schemaName, snapshotName)

				message := fmt.Sprintf("Drop snapshot %s?", snapshot.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := t.connectionManager.GetClient().Snapshots.Drop(ctx, &snowflake.DropSnapshotOptions{
							Snapshot: &snapshot,
						})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped snapshot %s", snapshot.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop snapshot %s", snapshot.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *SnapshotsView) GetRender() tview.Primitive {
	return v.table
}
