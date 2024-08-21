package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SecretsView struct {
	connectionManager *snowflake.ConnectionManager
	table             *tview.Table
	options           *SecretsOptions
}

type SecretsOptions struct {
	Database           *sdk.AccountObjectIdentifier
	Schema             *sdk.DatabaseObjectIdentifier
	Application        *sdk.AccountObjectIdentifier
	ApplicationPackage *sdk.AccountObjectIdentifier
}

func NewSecretsView(connectionManager *snowflake.ConnectionManager, opts *SecretsOptions) *SecretsView {
	secrets := &SecretsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	secrets.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return secrets
}

func (v *SecretsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx, v.options)
	if err != nil {
		return fmt.Errorf("updating network rules data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (t *SecretsView) getData(ctx context.Context, opts *SecretsOptions) (*Table, error) {
	title := "secrets"
	snowflakeOpts := snowflake.ShowSecretsOptions{}
	if opts.Database != nil {
		snowflakeOpts.Database = opts.Database
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		snowflakeOpts.Schema = opts.Schema
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.Schema.FullyQualifiedName())
	}
	if opts.Application != nil {
		snowflakeOpts.Application = opts.Application
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.Application.FullyQualifiedName())
	}
	if opts.ApplicationPackage != nil {
		snowflakeOpts.ApplicationPackage = opts.ApplicationPackage
		title = fmt.Sprintf("streamlits([pink]%s[blue])", opts.ApplicationPackage.FullyQualifiedName())
	}

	secrets, err := t.connectionManager.GetClient().Secrets.Show(ctx, &snowflakeOpts)
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show streamlits %w", err)
	}

	columns := []string{"Database", "Schema", "Name", "Secret Type"}
	rows := make([][]string, 0)

	for _, secret := range secrets {
		rows = append(rows, []string{
			secret.DatabaseName,
			secret.SchemaName,
			secret.Name,
			secret.SecretType,
		})
	}

	return &Table{
		Title:   title,
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *SecretsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				secretName := v.table.GetCell(r, 2).Text
				secret := sdk.NewSchemaObjectIdentifier(
					databaseName, schemaName, secretName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeSecret,
							ObjectIdentifier: secret,
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
				r, _ := v.table.GetSelection()
				databaseName := v.table.GetCell(r, 0).Text
				schemaName := v.table.GetCell(r, 1).Text
				secretName := v.table.GetCell(r, 2).Text
				secret := sdk.NewSchemaObjectIdentifier(databaseName, schemaName, secretName)

				message := fmt.Sprintf("Drop secret %s?", secret.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().Secrets.Drop(ctx, &snowflake.DropSecretsOptions{
							Secret: &secret,
						})
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped secret %s", secret.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop secret %s", secret.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *SecretsView) GetRender() tview.Primitive {
	return v.table
}
