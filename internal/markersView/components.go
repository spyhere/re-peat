package markersview

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func drawClickableIcon(gtx layout.Context, icon *widget.Icon, iconSize unit.Dp, iconC color.NRGBA, cl *widget.Clickable) layout.Dimensions {
	iconS := gtx.Dp(iconSize)
	gtx.Constraints.Min.X = iconS
	iconSizeHalf := iconS / 2
	common.DrawBox(gtx, common.Box{
		Size:      image.Rect(0, 0, iconS, iconS),
		R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
		Clickable: cl,
	})
	return icon.Layout(gtx, iconC)
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
						Chips:  props.chips,
						MaxLen: 20,
					})
				}),
			)
		})
	})
}
