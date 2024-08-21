package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Search struct {
	words      []string
	inputField *tview.InputField
}

func NewSearch() *Search {
	search := &Search{
		words: []string{
			"users",
			"roles",
			"databases",
			"schemas",
			"application packages",
			"applications",
			"compute pools",
			"listings",
			"security integrations",
			"stages",
			"tables",
			"views",
			"warehouses",
			"connections",
			"snapshots",
			"services",
			"image repositories",
			"procedures",
			"network rules",
			"network policies",
			"secrets",
		},
		inputField: tview.NewInputField().SetPlaceholder("snowflake object").SetFieldWidth(0),
	}

	search.inputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
		if len(currentText) == 0 {
			return
		}
		for _, word := range search.words {
			if strings.HasPrefix(strings.ToLower(word), strings.ToLower(currentText)) {
				entries = append(entries, word)
			}
		}
		if len(entries) < 1 {
			entries = nil
		}
		return
	})

	search.inputField.SetAutocompletedFunc(func(text string, index, source int) bool {
		if source != tview.AutocompletedNavigate {
			search.inputField.SetText(text)
		}
		return source == tview.AutocompletedEnter || source == tview.AutocompletedClick
	})

	return search
}

func (s *Search) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{
		{
			Description: "Search",
			Event:       tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				applicationState.Pages.SwitchToPage("main")

				switch applicationState.search.Value() {
				case "users":
					applicationState.Push(
						ctx,
						NewUsersView(applicationState.ConnectionManager, &UsersOptions{}),
					)
				case "roles":
					applicationState.Push(
						ctx,
						NewRolesView(applicationState.ConnectionManager, &RolesOptions{}),
					)
				case "databases":
					applicationState.Push(
						ctx,
						NewDatabasesView(applicationState.ConnectionManager, &DatabasesOptions{}),
					)
				case "schemas":
					applicationState.Push(
						ctx,
						NewSchemasView(applicationState.ConnectionManager, &SchemasOptions{}),
					)
				case "application packages":
					applicationState.Push(
						ctx,
						NewApplicationPackagesView(applicationState.ConnectionManager, &ApplicationPackagesOptions{}),
					)
				case "applications":
					applicationState.Push(
						ctx,
						NewApplicationsView(applicationState.ConnectionManager, &ApplicationsOptions{}),
					)
				case "compute pools":
					applicationState.Push(
						ctx,
						NewComputePoolsView(applicationState.ConnectionManager, &ComputePoolsOptions{}),
					)
				case "listings":
					applicationState.Push(
						ctx,
						NewListingsView(applicationState.ConnectionManager, &ListingsOptions{}),
					)
				case "security integrations":
					applicationState.Push(
						ctx,
						NewSecurityIntegrationsView(applicationState.ConnectionManager, &SecurityIntegrationsOptions{}),
					)
				case "stages":
					applicationState.Push(
						ctx,
						NewStagesView(applicationState.ConnectionManager, &StagesOptions{}),
					)
				case "tables":
					applicationState.Push(
						ctx,
						NewTablesView(applicationState.ConnectionManager, &TablesOptions{}),
					)
				case "views":
					applicationState.Push(
						ctx,
						NewViewsView(applicationState.ConnectionManager, &ViewsOptions{}),
					)
				case "warehouses":
					applicationState.Push(
						ctx,
						NewWarehousesView(applicationState.ConnectionManager, &WarehousesOptions{}),
					)
				case "connections":
					applicationState.Push(
						ctx,
						NewConnectionsView(applicationState.ConnectionManager, &ConnectionsOptions{}),
					)
				case "snapshots":
					applicationState.Push(
						ctx,
						NewSnapshotsView(applicationState.ConnectionManager, &SnapshotsOptions{}),
					)
				case "services":
					applicationState.Push(
						ctx,
						NewServicesView(applicationState.ConnectionManager, &ServicesOptions{}),
					)
				case "image repositories":
					applicationState.Push(
						ctx,
						NewImageRepositoriesView(applicationState.ConnectionManager, &ImageRepositoriesOptions{}),
					)
				case "procedures":
					applicationState.Push(
						ctx,
						NewProceduresView(applicationState.ConnectionManager, &ProceduresOptions{}),
					)
				case "network rules":
					applicationState.Push(
						ctx,
						NewNetworkRulesView(applicationState.ConnectionManager, &NetworkRulesOptions{}),
					)
				case "network policies":
					applicationState.Push(
						ctx,
						NewNetworkPoliciesView(applicationState.ConnectionManager, &NetworkPoliciesOptions{}),
					)
				case "streamlits":
					applicationState.Push(
						ctx,
						NewStreamlitsView(applicationState.ConnectionManager, &StreamlitsOptions{}),
					)
				case "secrets":
					applicationState.Push(
						ctx,
						NewSecretsView(applicationState.ConnectionManager, &SecretsOptions{}),
					)
				default:
					applicationState.status.SetError(fmt.Errorf("unknown snowflake object %s", applicationState.search.Value()))
					applicationState.Pages.SwitchToPage("search")
					return event
				}

				s.inputField.SetText("")
				return event
			},
		},
	}
}

func (s *Search) Clear() {
	s.inputField.SetText("")
}

func (s *Search) Value() string {
	return s.inputField.GetText()
}

func (s *Search) GetRender() *tview.InputField {
	return s.inputField
}
