package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SecurityIntegrationsView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *SecurityIntegrationsOptions
}

type SecurityIntegrationsOptions struct{}

func NewSecurityIntegrationsView(connectionManager *snowflake.ConnectionManager, opts *SecurityIntegrationsOptions) *SecurityIntegrationsView {
	securityIntegrations := &SecurityIntegrationsView{
		connectionManager: connectionManager,
		table:             tview.NewTable(),
		options:           opts,
	}

	securityIntegrations.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return securityIntegrations
}

func (v *SecurityIntegrationsView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating security integrations data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *SecurityIntegrationsView) getData(ctx context.Context) (*Table, error) {
	securityIntegrations, err := v.connectionManager.GetClient().SDKClient.SecurityIntegrations.Show(ctx, sdk.NewShowSecurityIntegrationRequest())
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show integrations %w", err)
	}

	columns := []string{"Name", "Integration Type"}
	rows := make([][]string, 0)

	for _, securityIntegration := range securityIntegrations {
		rows = append(rows, []string{
			securityIntegration.Name,
			securityIntegration.IntegrationType,
		})
	}

	return &Table{
		Title:   "security integrations",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *SecurityIntegrationsView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Grants",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				r, _ := v.table.GetSelection()
				securityIntegrationName := v.table.GetCell(r, 0).Text
				securityIntegration := sdk.NewAccountObjectIdentifier(
					securityIntegrationName,
				)

				applicationState.Push(
					ctx,
					NewGrantsView(
						applicationState.ConnectionManager,
						&GrantsOptions{
							ObjectType:       sdk.ObjectTypeIntegration,
							ObjectIdentifier: securityIntegration,
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
				securityIntegrationName := v.table.GetCell(r, 0).Text
				securityIntegration := sdk.NewAccountObjectIdentifier(securityIntegrationName)

				message := fmt.Sprintf("Drop security integration %s?", securityIntegration.FullyQualifiedName())

				applicationState.modal.Prompt(ctx, applicationState, message, func(action bool) {
					if action {
						err := v.connectionManager.GetClient().SDKClient.SecurityIntegrations.Drop(ctx, sdk.NewDropSecurityIntegrationRequest(securityIntegration))
						if err != nil {
							applicationState.status.SetError(err)
						}
						applicationState.status.SetMessage(fmt.Sprintf("Dropped security integration %s", securityIntegration.FullyQualifiedName()))
					} else {
						applicationState.status.SetMessage(fmt.Sprintf("Canceled drop security integration %s", securityIntegration.FullyQualifiedName()))
					}
				})

				return nil
			},
		},
	}
}

func (v *SecurityIntegrationsView) GetRender() tview.Primitive {
	return v.table
}
