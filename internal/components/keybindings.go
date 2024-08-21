package components

import (
	"fmt"

	"github.com/rivo/tview"
)

type KeyBindings struct {
	view     *tview.Table
	Bindings []*KeyBinding
}

func NewKeyBindings() *KeyBindings {
	keybindings := &KeyBindings{
		view:     tview.NewTable(),
		Bindings: make([]*KeyBinding, 0),
	}

	return keybindings
}

func (k *KeyBindings) Update() {
	cols, rows := 4, 3
	currentIndex := 0
	k.view.Clear()
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			for currentIndex < len(k.Bindings) && k.Bindings[currentIndex].Hidden {
				currentIndex++
			}
			if currentIndex >= len(k.Bindings) {
				continue
			}

			binding := k.Bindings[currentIndex]
			k.view.SetCell(r, c, tview.NewTableCell(fmt.Sprintf("[blue]<%s>: [grey]%s", binding.Name(), binding.Description)))
			currentIndex++
		}
	}
}

func (k *KeyBindings) Add(binding *KeyBinding) {
	k.Bindings = append(k.Bindings, binding)
}

func (k *KeyBindings) Clear() {
	k.Bindings = make([]*KeyBinding, 0)
}

func (k *KeyBindings) GetRender() *tview.Table {
	return k.view
}
