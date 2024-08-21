package components

import (
	"github.com/gdamore/tcell/v2"
)

type KeyBinding struct {
	Description string
	Event       *tcell.EventKey
	Hidden      bool
	Rune        rune
	Callback    func(event *tcell.EventKey) *tcell.EventKey
}

func (k *KeyBinding) Name() string {
	name := ""
	switch k.Event.Modifiers() {
	case tcell.ModAlt:
		name = "alt-"
	case tcell.ModCtrl:
		name = "ctrl-"
	case tcell.ModMeta:
		name = "meta-"
	case tcell.ModShift:
		name = "shift-"
	}

	if k.Event.Key() == tcell.KeyRune {
		return name + string(k.Event.Rune())
	}
	return name + tcell.KeyNames[k.Event.Key()]
}

type Table struct {
	Title   string
	Columns []string
	Rows    [][]string
}
