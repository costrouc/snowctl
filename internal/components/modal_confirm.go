package components

import (
	"context"

	"github.com/rivo/tview"
)

type ConfirmModal struct {
	modal *tview.Modal
}

func NewConfirmModal() *ConfirmModal {
	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Cancel", "Confirm"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {})

	return &ConfirmModal{
		modal: modal,
	}
}

func (m *ConfirmModal) Prompt(ctx context.Context, applicationState *ApplicationState, message string, confirm func(action bool)) {
	m.modal.SetText(message)
	m.modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		confirm(buttonLabel == "Confirm")
		applicationState.Pages.SwitchToPage("main")
		applicationState.UpdateView(ctx, false)
	})
	applicationState.Pages.SwitchToPage("modal")
	applicationState.UpdateView(ctx, false)
}

func (m *ConfirmModal) GetRender() *tview.Modal {
	return m.modal
}
