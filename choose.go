package main

import (
	"howm/frame"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/wingo/prompt"
)

type CycleWrap struct {
	Cycle *prompt.Cycle
	Choices []*prompt.CycleItem
}

func (c *CycleWrap) Destroy() {
	c.Cycle.Destroy()
	c.Cycle = nil
	c.Choices = make([]*prompt.CycleItem, 0)
}

type Choice struct {
	Win *xwindow.Window
	Context *frame.Context
	Wrapper *CycleWrap
}

func (c *Choice) CycleIsActive() bool {
	return true
}

func (c *Choice) CycleImage() *xgraphics.Image {
	ximg, _ := xgraphics.FindIcon(c.Context.X, c.Win.Id,
		prompt.DefaultCycleTheme.IconSize, prompt.DefaultCycleTheme.IconSize)
	// TODO: return empty image
	return ximg
}

func (c *Choice) CycleText() string {
	name, err := ewmh.WmNameGet(c.Context.X, c.Win.Id)
	if err != nil {
		return "N/A"
	}
	return name
}

func (c *Choice) CycleHighlighted() {
	// Chill
}

func (c *Choice) CycleSelected() {
	if f := c.Context.Get(c.Win.Id); f != nil {
		f.Container.RaiseFindFocus(c.Context)
	}
	c.Wrapper.Destroy()
}

func RegisterChooseHooks(ctx *frame.Context) {
	wrapper := &CycleWrap{}

	cycle := func(cycleDir string){
		if wrapper.Cycle != nil {
			if cycleDir == ctx.Config.TabBackward {
				wrapper.Cycle.Prev()
			} else {
				wrapper.Cycle.Next()
			}
			return
		}
	}

	register := func(cycleDir string){
		if wrapper.Cycle != nil {
			shown := wrapper.Cycle.Show(ctx.Screens[0].ToXRect(), cycleDir, wrapper.Choices)
			if !shown {
				wrapper.Destroy()
			} else {
				return
			}
		}

		wrapper.Cycle = prompt.NewCycle(ctx.X,
			prompt.DefaultCycleTheme, prompt.DefaultCycleConfig)

		wrapper.Choices = make([]*prompt.CycleItem, 0)
		if ctx.Config.TabByFrame {
			for _, f := range(ctx.Tracked) {
				if !f.IsLeaf() {
					continue
				}
				item := wrapper.Cycle.AddChoice(&Choice{f.Window, ctx, wrapper})
				wrapper.Choices = append(wrapper.Choices, item)
			}
		} else {
			for c, _ := range(ctx.Containers) {
				if f := c.Root.Find(func(fr *frame.Frame)bool{ return fr.IsLeaf() }); f != nil {
					item := wrapper.Cycle.AddChoice(&Choice{f.Window, ctx, wrapper})
					wrapper.Choices = append(wrapper.Choices, item)
				}
			}
		}
		wrapper.Cycle.Show(ctx.Screens[0].ToXRect(), cycleDir, wrapper.Choices)
		cycle(cycleDir)
	}

	keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent){
		register(ctx.Config.TabForward)
	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.TabForward, true)

	keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent){
		cycle(ctx.Config.TabForward)
	}).Connect(ctx.X, ctx.X.Dummy(), ctx.Config.TabForward, true)

	keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent){
		register(ctx.Config.TabBackward)
	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.TabBackward, true)

	keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent){
		cycle(ctx.Config.TabBackward)
	}).Connect(ctx.X, ctx.X.Dummy(), ctx.Config.TabBackward, true)
}