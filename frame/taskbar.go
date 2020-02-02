package frame

import (
	"log"
	"time"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/text"
	"github.com/BurntSushi/wingo/render"
)


type Taskbar struct {
	Base Decoration
	TimeWin *xwindow.Window
	Hidden bool
}

func TaskbarShape(ctx *Context) Rect {
	if len(ctx.Screens) == 0 {
		return Rect{
			X: 0,
			Y: 0,
			W: ctx.Config.ElemSize,
			H: ctx.Config.ElemSize,
		}
	}
	return Rect{
		X: ctx.Screens[0].X,
		Y: ctx.Screens[0].Y + ctx.Screens[0].H - ctx.Config.TaskbarHeight,
		W: ctx.Screens[0].W,
		H: ctx.Config.TaskbarHeight,
	}
}

func TimeShape(ctx *Context, time time.Time) Rect {
	ew, eh := xgraphics.Extents(prompt.DefaultInputTheme.Font, ctx.Config.TaskbarFontSize, time.Format(ctx.Config.TaskbarTimeFormat))
	s := TaskbarShape(ctx)
	return Rect{
		X: s.X + s.W - ew - ctx.Config.TaskbarXPad,
		Y: s.Y + s.H - eh - ctx.Config.TaskbarYPad,
		W: ew,
		H: eh,
	}
}

func NewTaskbar(ctx *Context) *Taskbar {
	t := &Taskbar{}
	var err error

	// Base background
	t.Base, err = CreateDecoration(ctx, TaskbarShape(ctx), ctx.Config.TaskbarBaseColor, 0)
	if err != nil {
		log.Fatal(err)
	}
	t.Base.Window.Map()

	// Time
	now := time.Now()
	s := TimeShape(ctx, now)
	win, err := xwindow.Generate(ctx.X)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	win.Create(ctx.X.RootWin(), s.X, s.Y, s.W, s.H, 0)
	win.Map()
	t.TimeWin = win

	// Initial render
	t.Update(ctx)
	return t
}

func (t *Taskbar) MoveResize(ctx *Context) {
	s := TaskbarShape(ctx)
	t.Base.Window.MoveResize(s.X, s.Y, s.W, s.H)
	st := TimeShape(ctx, time.Now())
	t.TimeWin.MoveResize(st.X, st.Y, st.W, st.H)
}

func (t *Taskbar) Update(ctx *Context) {
	now := time.Now()
	s := TimeShape(ctx, now)

	text.DrawText(
		t.TimeWin,
		prompt.DefaultInputTheme.Font,
		ctx.Config.TaskbarFontSize,
		render.NewColor(int(ctx.Config.TaskbarTextColor)),
		render.NewColor(int(ctx.Config.TaskbarBaseColor)),
		now.Format(ctx.Config.TaskbarTimeFormat),
	)
	t.TimeWin.MoveResize(s.X, s.Y, s.W, s.H)
}

func (t *Taskbar) Map() {
	t.Base.Window.Map()
	t.TimeWin.Map()
}

func (t *Taskbar) Unmap() {
	t.Base.Window.Unmap()
	t.TimeWin.Unmap()
}

func (t *Taskbar) Raise(ctx *Context) {
	t.Base.Window.Stack(xproto.StackModeAbove)
	t.TimeWin.Stack(xproto.StackModeAbove)
}

func (t *Taskbar) Lower(ctx *Context) {
	t.Base.Window.Stack(xproto.StackModeBelow)
	t.TimeWin.Stack(xproto.StackModeBelow)
}