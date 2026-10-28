package main

// GDI+ through its flat C API: the anti-aliased tray icon.

import (
	"math"
	"syscall"
	"unsafe"
)

var (
	gdiplus = syscall.NewLazyDLL("gdiplus.dll")

	procGdiplusStartup              = gdiplus.NewProc("GdiplusStartup")
	procGdipDeleteGraphics          = gdiplus.NewProc("GdipDeleteGraphics")
	procGdipSetSmoothingMode        = gdiplus.NewProc("GdipSetSmoothingMode")
	procGdipSetTextRenderingHint    = gdiplus.NewProc("GdipSetTextRenderingHint")
	procGdipSetPixelOffsetMode      = gdiplus.NewProc("GdipSetPixelOffsetMode")
	procGdipCreateSolidFill         = gdiplus.NewProc("GdipCreateSolidFill")
	procGdipDeleteBrush             = gdiplus.NewProc("GdipDeleteBrush")
	procGdipFillEllipse             = gdiplus.NewProc("GdipFillEllipse")
	procGdipCreateBitmapFromScan0   = gdiplus.NewProc("GdipCreateBitmapFromScan0")
	procGdipGetImageGraphicsContext = gdiplus.NewProc("GdipGetImageGraphicsContext")
	procGdipCreateHICONFromBitmap   = gdiplus.NewProc("GdipCreateHICONFromBitmap")
	procGdipDisposeImage            = gdiplus.NewProc("GdipDisposeImage")
	procGdipCreatePen1              = gdiplus.NewProc("GdipCreatePen1")
	procGdipDeletePen               = gdiplus.NewProc("GdipDeletePen")
	procGdipDrawLine                = gdiplus.NewProc("GdipDrawLine")
	procGdipSetPenStartCap          = gdiplus.NewProc("GdipSetPenStartCap")
	procGdipSetPenEndCap            = gdiplus.NewProc("GdipSetPenEndCap")
)

const (
	smoothingAntiAlias   = 4
	textClearTypeGridFit = 5
	textAntiAliasGridFit = 3
	pixelOffsetHalf      = 4
	unitPixel            = 2
	fontRegular          = 0
	fontBold             = 1
	alignNear            = 0
	alignCenter          = 1
	alignFar             = 2
	trimmingEllipsis     = 3
	formatNoWrap         = 0x00001000
	formatNoClip         = 0x00004000
	pixelFormat32bppARGB = 0x26200A
	lineCapRound         = 2
)

func f32(value float32) uintptr { return uintptr(math.Float32bits(value)) }

func startGdiplus() {
	input := struct {
		Version                  uint32
		DebugEventCallback       uintptr
		SuppressBackgroundThread int32
		SuppressExternalCodecs   int32
	}{Version: 1}
	var token uintptr
	procGdiplusStartup.Call(uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0)
}

// canvas draws on a device context through GDI+.
type canvas struct {
	graphics uintptr
	scale    float32
}

func (c *canvas) setup(clearType bool) {
	procGdipSetSmoothingMode.Call(c.graphics, smoothingAntiAlias)
	procGdipSetPixelOffsetMode.Call(c.graphics, pixelOffsetHalf)
	hint := uintptr(textAntiAliasGridFit)
	if clearType {
		hint = textClearTypeGridFit
	}
	procGdipSetTextRenderingHint.Call(c.graphics, hint)
}

func (c *canvas) close() { procGdipDeleteGraphics.Call(c.graphics) }

// px turns layout units (1/16 of the icon) into pixels.
func (c *canvas) px(value float32) float32 { return value * c.scale }

func withBrush(color uint32, body func(brush uintptr)) {
	var brush uintptr
	procGdipCreateSolidFill.Call(uintptr(color), uintptr(unsafe.Pointer(&brush)))
	body(brush)
	procGdipDeleteBrush.Call(brush)
}

func (c *canvas) circle(cx, cy, radius float32, color uint32) {
	cx, cy, radius = c.px(cx), c.px(cy), c.px(radius)
	withBrush(color, func(brush uintptr) {
		procGdipFillEllipse.Call(c.graphics, brush, f32(cx-radius), f32(cy-radius), f32(radius*2), f32(radius*2))
	})
}

func (c *canvas) line(x1, y1, x2, y2, width float32, color uint32) {
	var pen uintptr
	procGdipCreatePen1.Call(uintptr(color), f32(c.px(width)), unitPixel, uintptr(unsafe.Pointer(&pen)))
	procGdipSetPenStartCap.Call(pen, lineCapRound)
	procGdipSetPenEndCap.Call(pen, lineCapRound)
	procGdipDrawLine.Call(c.graphics, pen, f32(c.px(x1)), f32(c.px(y1)), f32(c.px(x2)), f32(c.px(y2)))
	procGdipDeletePen.Call(pen)
}

// iconImage draws with GDI+ into a transparent bitmap and returns it as an icon.
func iconImage(size int, draw func(c *canvas)) uintptr {
	var bitmap uintptr
	procGdipCreateBitmapFromScan0.Call(uintptr(size), uintptr(size), 0, pixelFormat32bppARGB, 0,
		uintptr(unsafe.Pointer(&bitmap)))
	c := &canvas{scale: float32(size) / 16}
	procGdipGetImageGraphicsContext.Call(bitmap, uintptr(unsafe.Pointer(&c.graphics)))
	c.setup(false)
	draw(c)
	c.close()
	var icon uintptr
	procGdipCreateHICONFromBitmap.Call(bitmap, uintptr(unsafe.Pointer(&icon)))
	procGdipDisposeImage.Call(bitmap)
	return icon
}
