package main

func newButtons() *buttons {
	return &buttons{
		arr: [3]*button{
			{
				name: "Project",
				tab:  Project,
				tag:  &struct{}{},
			},
			{
				name: "Markers",
				tab:  Markers,
				tag:  &struct{}{},
			},
			{
				name: "Editor",
				tab:  Editor,
				tag:  &struct{}{},
			},
		},
	}
}

type button struct {
	name string
	tab
	tag       *struct{}
	isHovered bool
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
