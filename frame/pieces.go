package frame

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xcursor"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/levavakian/rowm/ext"
	"time"
)

func GeneratePieces(ctx *Context, c *Container) error {
	// Create Decorations
	var err error
	c.Decorations.Grab, err = CreateDecoration(
		ctx,
		GrabShape(ctx, c.Shape),
		ctx.Config.GrabColor,
		0,
	)
	ext.Logerr(err)

	c.Decorations.Top, err = CreateDecoration(
		ctx,
		TopShape(ctx, c.Shape),
		ctx.Config.SeparatorColor,
		uint32(ctx.Cursors[xcursor.TopSide]),
	)
	ext.Logerr(err)

	c.Decorations.Bottom, err = CreateDecoration(
		ctx,
		BottomShape(ctx, c.Shape),
		ctx.Config.SeparatorColor,
		uint32(ctx.Cursors[xcursor.BottomSide]),
	)
	ext.Logerr(err)

	c.Decorations.Left, err = CreateDecoration(
		ctx,
		LeftShape(ctx, c.Shape),
		ctx.Config.SeparatorColor,
		uint32(ctx.Cursors[xcursor.LeftSide]),
	)
	ext.Logerr(err)

	c.Decorations.Right, err = CreateDecoration(
		ctx,
		RightShape(ctx, c.Shape),
		ctx.Config.SeparatorColor,
		uint32(ctx.Cursors[xcursor.RightSide]),
	)
	ext.Logerr(err)

	c.Decorations.BottomRight, err = CreateDecoration(
		ctx,
		BottomRightShape(ctx, c.Shape),
		ctx.Config.ResizeColor,
		uint32(ctx.Cursors[xcursor.BottomRightCorner]),
	)
	ext.Logerr(err)

	c.Decorations.BottomLeft, err = CreateDecoration(
		ctx,
		BottomLeftShape(ctx, c.Shape),
		ctx.Config.ResizeColor,
		uint32(ctx.Cursors[xcursor.BottomLeftCorner]),
	)
	ext.Logerr(err)

	c.Decorations.TopRight, err = CreateDecoration(
		ctx,
		TopRightShape(ctx, c.Shape),
		ctx.Config.ResizeColor,
		uint32(ctx.Cursors[xcursor.TopRightCorner]),
	)
	ext.Logerr(err)

	c.Decorations.TopLeft, err = CreateDecoration(
		ctx,
		TopLeftShape(ctx, c.Shape),
		ctx.Config.ResizeColor,
		uint32(ctx.Cursors[xcursor.TopLeftCorner]),
	)
	ext.Logerr(err)

	c.Decorations.Close, err = CreateDecoration(
		ctx,
		CloseShape(ctx, c.Shape),
		ctx.Config.CloseColor,
		uint32(ctx.Cursors[xcursor.DiamondCross]),
	)
	ext.Logerr(err)

	c.Decorations.Maximize, err = CreateDecoration(
		ctx,
		MaximizeShape(ctx, c.Shape),
		ctx.Config.MaximizeColor,
		uint32(ctx.Cursors[xcursor.Plus]),
	)
	ext.Logerr(err)

	c.Decorations.Minimize, err = CreateDecoration(
		ctx,
		MinimizeShape(ctx, c.Shape),
		ctx.Config.MinimizeColor,
		uint32(ctx.Cursors[xcursor.BottomTee]),
	)
	ext.Logerr(err)

	// Add hooks
	err = c.AddCloseHook(ctx)
	ext.Logerr(err)
	c.AddTopHook(ctx)
	c.AddBottomHook(ctx)
	c.AddLeftHook(ctx)
	c.AddRightHook(ctx)
	c.AddBottomRightHook(ctx)
	c.AddBottomLeftHook(ctx)
	c.AddTopRightHook(ctx)
	c.AddTopLeftHook(ctx)
	c.AddGrabHook(ctx)
	c.AddMaximizeHook(ctx)
	c.AddMinimizeHook(ctx)
	return err
}

func (c *Container) AddGrabHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.Grab.Window.Id, c.Decorations.Grab.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			dX := rX - c.DragContext.MouseX
			dY := rY - c.DragContext.MouseY
			c.MoveResize(ctx, c.DragContext.Container.X+dX, c.DragContext.Container.Y+dY, c.Shape.W, c.Shape.H)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			now := time.Now()
			if now.Sub(c.LastGrabTime) < ctx.Config.DoubleClickTime {
				screen, _, _ := ctx.GetScreenForShape(c.Shape)
				fullshape := AnchorShape(ctx, screen, FULL)
				if c.Shape == fullshape {
					c.MoveResizeShape(ctx, c.RestingShape(ctx, screen))
				} else {
					c.MoveResizeShape(ctx, fullshape)
				}
			}
			c.RaiseFindFocus(ctx)
			c.LastGrabTime = now
		},
	)
}

func (c *Container) AddCloseHook(ctx *Context) error {
	return mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			c.Root.Close(ctx)
		}).Connect(ctx.X, c.Decorations.Close.Window.Id, ctx.Config.ButtonClick, false, true)
}

func (c *Container) AddMinimizeHook(ctx *Context) error {
	return mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			c.ChangeMinimizationState(ctx)
		}).Connect(ctx.X, c.Decorations.Minimize.Window.Id, ctx.Config.ButtonClick, false, true)
}

func (c *Container) AddMaximizeHook(ctx *Context) error {
	return mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			screen, _, _ := ctx.GetScreenForShape(c.Shape)
			s := AnchorShape(ctx, screen, FULL)
			if c.Shape == s {
				c.MoveResizeShape(ctx, ctx.DefaultShapeForScreen(screen))
			} else {
				c.MoveResizeShape(ctx, s)
			}
		}).Connect(ctx.X, c.Decorations.Maximize.Window.Id, ctx.Config.ButtonClick, false, true)
}

func (c *Container) AddTopHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.Top.Window.Id, c.Decorations.Top.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			origYEnd := c.DragContext.Container.Y + c.DragContext.Container.H
			h := ext.IMax(origYEnd-rY, ctx.Config.MinShape().H)
			y := origYEnd - h
			c.MoveResize(ctx, c.DragContext.Container.X, y, c.DragContext.Container.W, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddBottomHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.Bottom.Window.Id, c.Decorations.Bottom.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			h := ext.IMax(rY-c.DragContext.Container.Y, ctx.Config.MinShape().H)
			c.MoveResize(ctx, c.DragContext.Container.X, c.DragContext.Container.Y, c.DragContext.Container.W, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddRightHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.Right.Window.Id, c.Decorations.Right.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			w := ext.IMax(rX-c.DragContext.Container.X, ctx.Config.MinShape().W)
			c.MoveResize(ctx, c.DragContext.Container.X, c.DragContext.Container.Y, w, c.DragContext.Container.H)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddLeftHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.Left.Window.Id, c.Decorations.Left.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			origXEnd := c.DragContext.Container.X + c.DragContext.Container.W
			w := ext.IMax(origXEnd-rX, ctx.Config.MinShape().W)
			x := origXEnd - w
			c.MoveResize(ctx, x, c.DragContext.Container.Y, w, c.DragContext.Container.H)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddBottomRightHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.BottomRight.Window.Id, c.Decorations.BottomRight.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			w := ext.IMax(rX-c.DragContext.Container.X, ctx.Config.MinShape().W)
			h := ext.IMax(rY-c.DragContext.Container.Y, ctx.Config.MinShape().H)
			c.MoveResize(ctx, c.DragContext.Container.X, c.DragContext.Container.Y, w, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddBottomLeftHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.BottomLeft.Window.Id, c.Decorations.BottomLeft.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			origXEnd := c.DragContext.Container.X + c.DragContext.Container.W
			w := ext.IMax(origXEnd-rX, ctx.Config.MinShape().W)
			x := origXEnd - w
			h := ext.IMax(rY-c.DragContext.Container.Y, ctx.Config.MinShape().H)
			c.MoveResize(ctx, x, c.DragContext.Container.Y, w, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddTopRightHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.TopRight.Window.Id, c.Decorations.TopRight.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			origYEnd := c.DragContext.Container.Y + c.DragContext.Container.H
			w := ext.IMax(rX-c.DragContext.Container.X, ctx.Config.MinShape().W)
			h := ext.IMax(origYEnd-rY, ctx.Config.MinShape().H)
			y := origYEnd - h
			c.MoveResize(ctx, c.DragContext.Container.X, y, w, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func (c *Container) AddTopLeftHook(ctx *Context) {
	mousebind.Drag(
		ctx.X, c.Decorations.TopLeft.Window.Id, c.Decorations.TopLeft.Window.Id, ctx.Config.ButtonDrag, true,
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) (bool, xproto.Cursor) {
			c.DragContext = GenerateDragContext(ctx, c, nil, rX, rY)
			c.RaiseFindFocus(ctx)
			return true, ctx.Cursors[xcursor.Circle]
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			origYEnd := c.DragContext.Container.Y + c.DragContext.Container.H
			origXEnd := c.DragContext.Container.X + c.DragContext.Container.W
			w := ext.IMax(origXEnd-rX, ctx.Config.MinShape().W)
			h := ext.IMax(origYEnd-rY, ctx.Config.MinShape().H)
			y := origYEnd - h
			x := origXEnd - w
			c.MoveResize(ctx, x, y, w, h)
		},
		func(X *xgbutil.XUtil, rX, rY, eX, eY int) {
			c.RaiseFindFocus(ctx)
		},
	)
}

func TopRightShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - context.Config.ElemSize,
		Y: cShape.Y + context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func TopLeftShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X,
		Y: cShape.Y + context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func BottomRightShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - context.Config.ElemSize,
		Y: cShape.Y + cShape.H - context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func BottomLeftShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X,
		Y: cShape.Y + cShape.H - context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func TopShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + context.Config.ElemSize,
		Y: cShape.Y + context.Config.ElemSize,
		W: cShape.W - 2*context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func BottomShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + context.Config.ElemSize,
		Y: cShape.Y + cShape.H - context.Config.ElemSize,
		W: cShape.W - 2*context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func LeftShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X,
		Y: cShape.Y + 2*context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: cShape.H - 3*context.Config.ElemSize,
	}
}

func RightShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - context.Config.ElemSize,
		Y: cShape.Y + 2*context.Config.ElemSize,
		W: context.Config.ElemSize,
		H: cShape.H - 3*context.Config.ElemSize,
	}
}

func GrabShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X,
		Y: cShape.Y,
		W: cShape.W - 3*context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func CloseShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - context.Config.ElemSize,
		Y: cShape.Y,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func MaximizeShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - 2*context.Config.ElemSize,
		Y: cShape.Y,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}

func MinimizeShape(context *Context, cShape Rect) Rect {
	return Rect{
		X: cShape.X + cShape.W - 3*context.Config.ElemSize,
		Y: cShape.Y,
		W: context.Config.ElemSize,
		H: context.Config.ElemSize,
	}
}
