package markersview

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type clickableIconProps struct {
	icon     *widget.Icon
	iconSize unit.Dp
	cl       *widget.Clickable
	disabled bool
}

// TODO: Looks like this should be a part of common DrawIconButton
func drawClickableIcon(gtx layout.Context, th *theme.RepeatTheme, props clickableIconProps) layout.Dimensions {
	iconS := gtx.Dp(props.iconSize)
	gtx.Constraints.Min.X = iconS
	iconSizeHalf := iconS / 2
	cl := props.cl
	color := th.Palette.Backdrop
	if props.disabled {
		color = th.Palette.IconButton.Disabled.Bg
		cl = nil
	}
	common.DrawBox(gtx, common.Box{
		Size:      image.Rect(0, 0, iconS, iconS),
		R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
		Clickable: cl,
	})
	return props.icon.Layout(gtx, color)
}

type drawMarkerDialogSizeSpecs struct {
	fieldsYMargin unit.Dp
	fieldsXMargin unit.Dp
	fieldW        unit.Dp
	gap           unit.Dp
}

var drawMarkerDialogSpecs = drawMarkerDialogSizeSpecs{
	fieldsYMargin: 10,
	fieldsXMargin: 10,
	fieldW:        270,
	gap:           20,
}

type markerDialogFieldsProps struct {
	name         *common.Inputable
	time         *common.Inputable
	tags         *common.Inputable
	chips        []string
	totalSeconds float64
	tagOptions   []common.ComboboxOption
}

func drawMarkerDialogFields(gtx layout.Context, th *theme.RepeatTheme, props markerDialogFieldsProps) layout.Dimensions {
	s := drawMarkerDialogSpecs
	inset := layout.Inset{Top: s.fieldsYMargin, Bottom: s.fieldsYMargin, Left: s.fieldsXMargin, Right: s.fieldsXMargin}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gapPx := gtx.Dp(s.gap)
		gtx.Constraints.Max.X = gtx.Constraints.Min.X
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		fieldW := gtx.Dp(s.fieldW)
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					inputDims := common.DrawInputField(gtx, th, common.InputFieldProps{
						Base: common.InputFieldBase{
							LabelText: "Имя",
							Inputable: props.name,
						},
						MaxLen:      20,
						Placeholder: "Новый маркер...",
					})
					inputDims.Size.Y += gapPx
					return inputDims
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					inputDims := common.DrawInputField(gtx, th, common.InputFieldProps{
						Base: common.InputFieldBase{
							LabelText: "Время",
							Inputable: props.time,
						},
						MaxLen:      7,
						Placeholder: common.FormatSeconds(props.totalSeconds),
					})
					inputDims.Size.Y += gapPx
					return inputDims
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					return common.DrawCombobox(gtx, th, common.ComboboxProps{
						Base: common.InputFieldBase{
							LabelText: "Категории",
							Inputable: props.tags,
						},
						Chips:   props.chips,
						MaxLen:  20,
						Options: props.tagOptions,
					})
				}),
			)
		})
	})
}
