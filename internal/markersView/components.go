package markersview

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
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
	if !gtx.Enabled() {
		props.disabled = true
	}
	iconS := gtx.Dp(props.iconSize)
	gtx.Constraints.Min.X = iconS
	iconSizeHalf := iconS / 2
	cl := props.cl
	color := th.Palette.Backdrop
	if props.disabled {
		color = th.Palette.IconButton.Disabled.Bg
		cl = nil
	}
	if !gtx.Enabled() {
		cl = nil
	}
	common.DrawBox(gtx, common.Box{
		Size:      image.Rect(0, 0, iconS, iconS),
		R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
		Clickable: cl,
	})
	return props.icon.Layout(gtx, color)
}

func drawAddMarkerButton(gtx layout.Context, th *theme.RepeatTheme, cl *widget.Clickable, x, y int) {
	addIconM, addIconDims := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		iconStyle := material.IconButton(th.Theme, cl, micons.ContentAddCircle, "")
		iconStyle.Size = 20
		iconStyle.Inset = layout.UniformInset(7)
		iconStyle.Background = th.Palette.Editor.SoundWave
		gtx.Constraints.Min = image.Point{}
		return iconStyle.Layout(gtx)
	})
	common.OffsetBy(gtx, image.Pt(x, y-addIconDims.Size.Y/2), func(gtx layout.Context) {
		addIconM.Add(gtx.Ops)
	})
}

type fieldGroupStyle struct {
	fieldsYMargin unit.Dp
	fieldsXMargin unit.Dp
	fieldW        unit.Dp
	gap           unit.Dp
}

func defaultFieldGroupStyle() fieldGroupStyle {
	return fieldGroupStyle{
		fieldsYMargin: 10,
		fieldsXMargin: 10,
		fieldW:        270,
		gap:           20,
	}
}

type playerStateStyle struct {
	width     float32
	maxWidth  unit.Dp
	yOffset   unit.Dp
	bgH       unit.Dp
	bgShape   int
	trackH    unit.Dp
	volumeH   unit.Dp
	volumeX   unit.Dp
	lineShape int
	thumbDiam unit.Dp
	inset     unit.Dp
}

func defaultPlayerStateStyle() playerStateStyle {
	return playerStateStyle{
		maxWidth:  780,
		width:     80.0,
		yOffset:   90,
		bgH:       70,
		bgShape:   10,
		trackH:    3,
		volumeH:   1,
		volumeX:   150,
		lineShape: 5,
		thumbDiam: 16,
		inset:     10,
	}
}

func drawThumb(gtx layout.Context, bg color.NRGBA, diameter int) {
	thumbRadi := diameter / 2
	common.DrawBox(gtx, common.Box{
		Size:  image.Rect(0, 0, diameter, diameter),
		Color: bg,
		R:     theme.CornerR(thumbRadi, thumbRadi, thumbRadi, thumbRadi),
	})
}

type playerControllable struct {
	totalS           float64
	isVolumeHoevered bool
	volumeMaxX       int
	volume           float64
	volumeTag        struct{}
	hasNewVolume     bool
	isSilent         bool
	muteTag          struct{}
	isMutedHovered   bool
}

func (p *playerControllable) setVolume(vol float64, silent bool) {
	p.volume = vol
	p.isSilent = silent
}

func (p *playerControllable) getCursorType() (pointer.Cursor, bool) {
	if p.isMutedHovered || p.isVolumeHoevered {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}

func (p *playerControllable) update(gtx layout.Context) {
	common.HandlePointerEvents(gtx, &p.volumeTag, pointer.Enter|pointer.Leave|pointer.Press, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			p.isVolumeHoevered = true
		case pointer.Leave:
			p.isVolumeHoevered = false
		case pointer.Press:
			p.volume = float64(e.Position.X / float32(p.volumeMaxX))
			p.hasNewVolume = true
			p.isSilent = false
		}
	})
	common.HandlePointerEvents(gtx, &p.muteTag, pointer.Enter|pointer.Leave|pointer.Press, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			p.isMutedHovered = true
		case pointer.Leave:
			p.isMutedHovered = false
		case pointer.Press:
			p.isSilent = !p.isSilent
			p.hasNewVolume = true
		}
	})
}

func (p *playerControllable) getNewVolume() (float64, bool, bool) {
	hasNewVolume := p.hasNewVolume
	p.hasNewVolume = false
	return p.volume, p.isSilent, hasNewVolume
}

func getVolumeIcon(volume float64, isSilent bool) *widget.Icon {
	volIcon := micons.VolumeOff
	if volume <= 0 || isSilent {
		return volIcon
	}
	if volume > 0.6 {
		volIcon = micons.VolumeUp
	} else if volume > 0.3 {
		volIcon = micons.VolumeDown
	} else if volume > 0.01 {
		volIcon = micons.VolumeMuted
	}
	return volIcon
}

func drawPlayerState(gtx layout.Context, th *theme.RepeatTheme, curS float64, pc *playerControllable) {
	pc.update(gtx)
	var timeLabel strings.Builder
	fmt.Fprint(&timeLabel, common.FormatSeconds(curS))
	timeLabel.WriteString(" / ")
	fmt.Fprint(&timeLabel, common.FormatSeconds(pc.totalS))

	s := defaultPlayerStateStyle()
	lineH := gtx.Dp(s.trackH)
	common.OffsetBy(gtx, image.Pt(0, gtx.Constraints.Max.Y-gtx.Dp(s.yOffset)), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {

			// Bg
			width := min(common.PrcToPx(gtx.Constraints.Max.X, s.width), gtx.Dp(s.maxWidth))
			bgDims := common.DrawBox(gtx, common.Box{
				Size:  image.Rect(0, 0, width, gtx.Dp(s.bgH)),
				Color: th.Palette.Backdrop,
				R:     theme.CornerR(s.bgShape, s.bgShape, s.bgShape, s.bgShape),
			})

			gtx.Constraints.Min = bgDims.Size
			gtx.Constraints.Max = bgDims.Size
			timeM, timeDims := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txtStyle := material.H5(th.Theme, timeLabel.String())
				txtStyle.Color = th.Bg
				return txtStyle.Layout(gtx)
			})
			layout.UniformInset(s.inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.Y = timeDims.Size.Y
						gtx.Constraints.Min.Y = timeDims.Size.Y
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = 60
									return micons.Pause.Layout(gtx, th.Bg)
								})
							}),
							layout.Rigid(layout.Spacer{Width: 25}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								timeM.Add(gtx.Ops)
								return timeDims
							}),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												gtx.Constraints.Min.X = 42
												volIcon := getVolumeIcon(pc.volume, pc.isSilent)
												common.RegisterTag(gtx, &pc.muteTag, image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Max.Y))
												return volIcon.Layout(gtx, th.Bg)
											})
										}),
										layout.Rigid(layout.Spacer{Width: 15}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lineH := gtx.Dp(s.volumeH)
											half := gtx.Constraints.Min.Y/2 - lineH/2
											volX := gtx.Dp(s.volumeX)
											pc.volumeMaxX = volX
											lineDims := common.DrawBox(gtx, common.Box{
												Size:  image.Rect(0, half, volX, half+lineH),
												Color: th.Bg,
											})
											areaPad := gtx.Dp(6)
											common.RegisterTag(gtx, &pc.volumeTag, image.Rect(0, areaPad, volX, gtx.Constraints.Max.Y-areaPad))
											thumbDiam := gtx.Dp(s.thumbDiam)
											thumbRadi := thumbDiam / 2
											xOffset := int(pc.volume * float64(volX))
											common.OffsetBy(gtx, image.Pt(xOffset-thumbRadi, half-thumbDiam/2), func(gtx layout.Context) {
												drawThumb(gtx, th.Bg, thumbDiam)
											})
											return lineDims
										}),
									)
								})
							}),
						)
					}),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Line
						lineDims := common.DrawBox(gtx, common.Box{
							Size:  image.Rect(0, 0, gtx.Constraints.Max.X, lineH),
							Color: th.Bg,
							R:     theme.CornerR(s.lineShape, s.lineShape, s.lineShape, s.lineShape),
						})

						// Thumb
						xOffset := int(curS) * gtx.Constraints.Max.X / int(pc.totalS)
						thumbDiam := gtx.Dp(s.thumbDiam)
						thumbRadi := thumbDiam / 2
						common.OffsetBy(gtx, image.Pt(xOffset-thumbRadi, -thumbDiam/2+lineH/2), func(gtx layout.Context) {
							drawThumb(gtx, th.Bg, thumbDiam)
						})
						return lineDims
					}),
				)
			})
			return bgDims
		})
	})
}
