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

type playerControllable struct {
	currentSec       float64
	totalS           float64
	isVolumeHoevered bool
	volumeMaxX       int
	volume           float64
	volumeTag        struct{}
	hasNewVolume     bool
	isSilent         bool
	muteTag          struct{}
	isMutedHovered   bool
	isPlayHovered    bool
	playTag          struct{}
	playbackEvent    bool
}

func (p *playerControllable) getCursorType() (pointer.Cursor, bool) {
	if p.isPlayHovered || p.isMutedHovered || p.isVolumeHoevered {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}

func (p *playerControllable) update(gtx layout.Context) {
	common.HandlePointerEvents(gtx, &p.playTag, pointer.Enter|pointer.Leave|pointer.Press, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			p.isPlayHovered = true
		case pointer.Leave:
			p.isPlayHovered = false
		case pointer.Press:
			p.playbackEvent = true
		}
	})
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

func (p *playerControllable) hasPlayEvent() bool {
	hasPlayEvent := p.playbackEvent
	p.playbackEvent = false
	return hasPlayEvent
}

func (p *playerControllable) getNewVolume() (float64, bool, bool) {
	hasNewVolume := p.hasNewVolume
	p.hasNewVolume = false
	return p.volume, p.isSilent, hasNewVolume
}
func (p *playerControllable) setVolume(vol float64, silent bool) {
	p.volume = vol
	p.isSilent = silent
}

type playerStateStyles struct {
	pc        *playerControllable
	th        *theme.RepeatTheme
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

func playerState(th *theme.RepeatTheme, pc *playerControllable) playerStateStyles {
	return playerStateStyles{
		pc:        pc,
		th:        th,
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

func (pss playerStateStyles) drawThumb(gtx layout.Context, bg color.NRGBA, diameter int) {
	thumbRadi := diameter / 2
	common.DrawBox(gtx, common.Box{
		Size:  image.Rect(0, 0, diameter, diameter),
		Color: bg,
		R:     theme.CornerR(thumbRadi, thumbRadi, thumbRadi, thumbRadi),
	})
}

func (pss playerStateStyles) getVolumeIcon(volume float64, isSilent bool) *widget.Icon {
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

func (pss playerStateStyles) Layout(gtx layout.Context) {
	pss.pc.update(gtx)
	var timeLabel strings.Builder
	fmt.Fprint(&timeLabel, common.FormatSeconds(pss.pc.currentSec))
	timeLabel.WriteString(" / ")
	fmt.Fprint(&timeLabel, common.FormatSeconds(pss.pc.totalS))

	lineH := gtx.Dp(pss.trackH)
	common.OffsetBy(gtx, image.Pt(0, gtx.Constraints.Max.Y-gtx.Dp(pss.yOffset)), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {

			// Bg
			width := min(common.PrcToPx(gtx.Constraints.Max.X, pss.width), gtx.Dp(pss.maxWidth))
			bgDims := common.DrawBox(gtx, common.Box{
				Size:  image.Rect(0, 0, width, gtx.Dp(pss.bgH)),
				Color: pss.th.Palette.Backdrop,
				R:     theme.CornerR(pss.bgShape, pss.bgShape, pss.bgShape, pss.bgShape),
			})

			gtx.Constraints.Min = bgDims.Size
			gtx.Constraints.Max = bgDims.Size
			timeM, timeDims := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txtStyle := material.H5(pss.th.Theme, timeLabel.String())
				txtStyle.Color = pss.th.Bg
				return txtStyle.Layout(gtx)
			})
			layout.UniformInset(pss.inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.Y = timeDims.Size.Y
						gtx.Constraints.Min.Y = timeDims.Size.Y
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = 60
									common.RegisterTag(gtx, &pss.pc.playTag, image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y))
									return micons.Pause.Layout(gtx, pss.th.Bg)
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
												volIcon := pss.getVolumeIcon(pss.pc.volume, pss.pc.isSilent)
												common.RegisterTag(gtx, &pss.pc.muteTag, image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Max.Y))
												return volIcon.Layout(gtx, pss.th.Bg)
											})
										}),
										layout.Rigid(layout.Spacer{Width: 15}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lineH := gtx.Dp(pss.volumeH)
											half := gtx.Constraints.Min.Y/2 - lineH/2
											volX := gtx.Dp(pss.volumeX)
											pss.pc.volumeMaxX = volX
											lineDims := common.DrawBox(gtx, common.Box{
												Size:  image.Rect(0, half, volX, half+lineH),
												Color: pss.th.Bg,
											})
											areaPad := gtx.Dp(6)
											common.RegisterTag(gtx, &pss.pc.volumeTag, image.Rect(0, areaPad, volX, gtx.Constraints.Max.Y-areaPad))
											thumbDiam := gtx.Dp(pss.thumbDiam)
											thumbRadi := thumbDiam / 2
											xOffset := int(pss.pc.volume * float64(volX))
											common.OffsetBy(gtx, image.Pt(xOffset-thumbRadi, half-thumbDiam/2), func(gtx layout.Context) {
												pss.drawThumb(gtx, pss.th.Bg, thumbDiam)
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
							Color: pss.th.Bg,
							R:     theme.CornerR(pss.lineShape, pss.lineShape, pss.lineShape, pss.lineShape),
						})

						// Thumb
						xOffset := int(pss.pc.currentSec) * gtx.Constraints.Max.X / int(pss.pc.totalS)
						thumbDiam := gtx.Dp(pss.thumbDiam)
						thumbRadi := thumbDiam / 2
						common.OffsetBy(gtx, image.Pt(xOffset-thumbRadi, -thumbDiam/2+lineH/2), func(gtx layout.Context) {
							pss.drawThumb(gtx, pss.th.Bg, thumbDiam)
						})
						return lineDims
					}),
				)
			})
			return bgDims
		})
	})
}
