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
	tag *struct{}
}

type buttons struct {
	arr              [3]*button
	isPointerHitting bool
}

func (b *buttons) setHover() {
	b.isPointerHitting = true
}

func (b *buttons) stopHover() {
	b.isPointerHitting = false
}
