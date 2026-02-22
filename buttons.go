package main

import "gioui.org/widget"

func newButtons() *buttons {
	return &buttons{
		arr: [3]*button{
			{
				name:      "Project",
				tab:       Project,
				tag:       &struct{}{},
				clickable: &widget.Clickable{},
			},
			{
				name:      "Markers",
				tab:       Markers,
				tag:       &struct{}{},
				clickable: &widget.Clickable{},
			},
			{
				name:      "Editor",
				tab:       Editor,
				tag:       &struct{}{},
				clickable: &widget.Clickable{},
			},
		},
	}
}

type button struct {
	name string
	tab
	tag       *struct{}
	isHovered bool
	clickable *widget.Clickable
}

type buttons struct {
	arr              [3]*button
	isPointerHitting bool
	isDisabled       bool
}

func (b *buttons) setHover(curButton *button) {
	b.isPointerHitting = true
	curButton.isHovered = true
}

func (b *buttons) stopHover(curButton *button) {
	b.isPointerHitting = false
	curButton.isHovered = false
}

func (b *buttons) disable() {
	b.isDisabled = true
}

func (b *buttons) enable() {
	b.isDisabled = false
}
