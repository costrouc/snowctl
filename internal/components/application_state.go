package components

import (
	"context"
	"fmt"

	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Component interface {
	GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding
	GetRender() tview.Primitive
	Update(ctx context.Context) error
}

type ApplicationState struct {
	// current key bindings for the given application
	bindings []*KeyBinding

	// snowflake client
	ConnectionManager *snowflake.ConnectionManager

	// history is a stack that represents the state of the application
	history []Component

	Application *tview.Application
	Pages       *tview.Pages
	Main        *tview.Pages

	context     *SnowflakeContext
	keyBindings *KeyBindings
	search      *Search
	status      *Status
	modal       *ConfirmModal
}

func NewApplication(cm *snowflake.ConnectionManager) *ApplicationState {
	applicationState := &ApplicationState{
		bindings:          make([]*KeyBinding, 0),
		history:           make([]Component, 0),
		ConnectionManager: cm,

		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		Main:        tview.NewPages(),

		context:     NewSnowflakeContext(cm),
		keyBindings: NewKeyBindings(),
		search:      NewSearch(),
		status:      NewStatus(),
		modal:       NewConfirmModal(),
	}

	applicationState.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		for _, keyBinding := range applicationState.bindings {
			if event.Name() == keyBinding.Event.Name() {
				return keyBinding.Callback(event)
			}
		}
		return event
	})

	applicationState.Pages.AddPage("main", viewPage(applicationState), true, true)
	applicationState.Pages.AddPage("search", searchPage(applicationState), true, false)
	applicationState.Pages.AddPage("modal", applicationState.modal.GetRender(), true, false)

	applicationState.Application.SetRoot(applicationState.Pages, true).EnableMouse(true).EnablePaste(true)

	return applicationState
}

func viewPage(applicationState *ApplicationState) *tview.Grid {
	grid := tview.NewGrid().SetRows(4, 1, 0, 2).SetColumns(0, 0).
		AddItem(applicationState.context.GetRender(), 0, 0, 1, 1, 0, 0, false).
		AddItem(applicationState.keyBindings.GetRender(), 0, 1, 1, 1, 0, 0, false).
		AddItem(applicationState.Main, 1, 0, 2, 2, 0, 0, true).
		AddItem(applicationState.status.GetRender(), 3, 0, 1, 2, 0, 0, false)

	return grid
}

func searchPage(applicationState *ApplicationState) *tview.Grid {
	grid := tview.NewGrid().SetRows(4, 1, 0, 2).SetColumns(0, 0).
		AddItem(applicationState.context.GetRender(), 0, 0, 1, 1, 0, 0, false).
		AddItem(applicationState.keyBindings.GetRender(), 0, 1, 1, 1, 0, 0, false).
		AddItem(applicationState.search.GetRender(), 1, 0, 1, 2, 0, 0, true).
		AddItem(applicationState.Main, 2, 0, 1, 2, 0, 0, false).
		AddItem(applicationState.status.GetRender(), 3, 0, 1, 2, 0, 0, false)

	return grid
}

func (a *ApplicationState) Push(ctx context.Context, component Component) {
	a.history = append(a.history, component)
	a.UpdateView(ctx, true)
}

func (a *ApplicationState) Pop(ctx context.Context) {
	if len(a.history) == 1 {
		return
	}

	a.Main.RemovePage(fmt.Sprintf("page%d", len(a.history)))
	a.history = a.history[:len(a.history)-1]
	a.UpdateView(ctx, false)
}

func (a *ApplicationState) UpdateView(ctx context.Context, newPage bool) {
	a.bindings = []*KeyBinding{
		{
			Description: "quit",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
			Hidden:      false,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				if name, _ := a.Pages.GetFrontPage(); name != "search" {
					a.Application.Stop()
				}
				return event
			},
		},
		{
			Description: "search",
			Event:       tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				a.search.Clear()
				a.Pages.SwitchToPage("search")
				a.UpdateView(ctx, false)
				return nil
			},
		},
		{
			Description: "cancel",
			Event:       tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				pageName, _ := a.Pages.GetFrontPage()
				switch pageName {
				case "search":
					a.search.Clear()
					a.Pages.SwitchToPage("main")
					a.UpdateView(ctx, false)
				case "main":
					a.Pop(ctx)
				}
				return event
			},
		},
		// emacs compatibility
		{
			Description: "page_down",
			Event:       tcell.NewEventKey(tcell.KeyCtrlV, 0, tcell.ModCtrl),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
			},
		},
		{
			Description: "page_up",
			Event:       tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModAlt),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
			},
		},
		{
			Description: "down",
			Event:       tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModCtrl),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			},
		},
		{
			Description: "up",
			Event:       tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModCtrl),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			},
		},
		{
			Description: "right",
			Event:       tcell.NewEventKey(tcell.KeyCtrlF, 0, tcell.ModCtrl),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
			},
		},
		{
			Description: "left",
			Event:       tcell.NewEventKey(tcell.KeyCtrlB, 0, tcell.ModCtrl),
			Hidden:      true,
			Callback: func(event *tcell.EventKey) *tcell.EventKey {
				return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
			},
		},
	}

	name, _ := a.Pages.GetFrontPage()
	switch name {
	case "search":
		a.bindings = append(a.bindings, a.search.GetBindings(ctx, a)...)
		return
	case "modal":
		return
	}

	component := a.history[len(a.history)-1]
	a.bindings = append(a.bindings, component.GetBindings(ctx, a)...)
	err := component.Update(ctx)
	if err != nil {
		a.status.SetError(err)
		return
	}
	if newPage {
		a.Main.AddAndSwitchToPage(fmt.Sprintf("page%d", len(a.history)), component.GetRender(), true)
	}

	a.context.Update(ctx)

	a.keyBindings.Clear()
	for _, binding := range a.bindings {
		a.keyBindings.Add(binding)
	}
	a.keyBindings.Update()
}
