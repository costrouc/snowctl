package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Status struct {
	view *tview.TextView
}

func NewStatus() *Status {
	return &Status{}
}

func (s *Status) GetRender() *tview.TextView {
	if s.view == nil {
		s.view = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("SNOWCTL").SetTextColor(tcell.ColorGrey).SetTextStyle(tcell.StyleDefault.Bold(true))
	}

	return s.view
}

func (s *Status) SetError(err error) {
	s.view.SetText(fmt.Sprintf("Error: %s", err.Error())).SetTextColor(tcell.ColorRed)
}

func (s *Status) SetMessage(message string) {
	s.view.SetText(message).SetTextColor(tcell.ColorGrey)
}
