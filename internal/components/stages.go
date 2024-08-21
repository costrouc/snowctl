package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StagesView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *StagesOptions
}

type StagesOptions struct {
	Database *string
	Schema   *string
}

func NewStagesView(connectionManager *snowflake.ConnectionManager, opts *StagesOptions) *StagesView {
	stages := &StagesView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	stages.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return stages
}

func (v *StagesView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating stages data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *StagesView) getData(ctx context.Context, opts *StagesOptions) (*Table, error) {
	sdkOpts := sdk.NewShowStageRequest()
	title := "stages"
	if opts.Database != nil && opts.Schema != nil {
		sdkOpts.WithIn(&sdk.In{
			Schema: sdk.NewDatabaseObjectIdentifier(*opts.Database, *opts.Schema),
		})
		title = fmt.Sprintf("stages([pink]%s.%s[blue])", *opts.Database, *opts.Schema)
	}

	stages, err := t.connectionManager.GetClient().SDKClient.Stages.Show(ctx, sdkOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show stages %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Owner", "Type"}
	rows := make([][]string, 0)

	for _, stage := range stages {
		rows = append(rows, []string{
			stage.DatabaseName,
			stage.SchemaName,
			stage.Name,
			stage.Owner,
			stage.Type,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (t *StagesView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				databaseName := t.table.GetCell(r, 0).Text
				schemaName := t.table.GetCell(r, 1).Text
				name := t.table.GetCell(r, 2).Text
				stage := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, name,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeStage,
							ObjectIdentifier: stage,
						},
					),
				)
				return nil
			},
		},
		{
			Description: "Drop",
			Event:       tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModCtrl),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := t.table.GetSelection()
				databaseName := t.table.GetCell(r, 0).Text
				schemaName := t.table.GetCell(r, 1).Text
				stageName := t.table.GetCell(r, 2).Text
				stage := sdk.NewSchemaObjectIdentifier(databaseName, schemaName, stageName)

				message := fmt.Sprintf("Drop stage %s?", stage.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := t.connectionManager.GetClient().SDKClient.Stages.Drop(ctx, sdk.NewDropStageRequest(stage))
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped stage %s", stage.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop stage %s", stage.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *StagesView) GetRender() tview.Primitive {
	return v.table
}
