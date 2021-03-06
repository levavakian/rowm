package root

import (
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/levavakian/rowm/frame"
	"fmt"
	"sort"
	"log"
	"os/exec"
	"time"
	"reflect"
)

func GenerateHelp(ctx *frame.Context) string {
	v := reflect.ValueOf(ctx.Config)

	helpString:=""
	helpMap := make(map[string]string)

	// Add all the fields in Config with help
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Type() == reflect.TypeOf((*frame.StringWithHelp)(nil)).Elem() {
			a := v.Field(i).Interface().(frame.StringWithHelp)
			helpMap[a.Data] = a.Help
		}
	}

	// Add all the Builtin Commands
	for k, _ := range ctx.Config.BuiltinCommands {
		helpMap[k.Data] = k.Help
	}
	keys := make([]string, 0, len(helpMap))
	for k := range helpMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		helpString += fmt.Sprintf("%s = %s\n", k, helpMap[k])
	}
	return helpString
}

func Split(ctx *frame.Context) *frame.Frame {
	if ctx.SplitPrompt != nil {
		ctx.SplitPrompt.Destroy()
		ctx.SplitPrompt = nil
		ctx.AttachPoint = nil
	}

	attachF := ctx.GetFocusedFrame()
	if attachF == nil {
		msgPrompt := prompt.NewMessage(ctx.X, prompt.DefaultMessageTheme, prompt.DefaultMessageConfig)
		timeout := 1 * time.Second
		msgPrompt.Show(ctx.Screens[0].ToXRect(), "Cannot split when not focused on a window", timeout, func(msg *prompt.Message) {})
		return nil
	}

	nprompt := prompt.NewInput(ctx.X, prompt.DefaultInputTheme, prompt.DefaultInputConfig)
	ctx.SplitPrompt = nprompt

	canc := func(inp *prompt.Input) {
		if ctx.SplitPrompt == nprompt {
			ctx.SplitPrompt.Destroy()
			ctx.SplitPrompt = nil
		}
	}

	resp := func(inp *prompt.Input, text string) {
		cmd := exec.Command("bash", "-c", text)
		err := cmd.Start()
		if err != nil {
			log.Println(err)
		}
		go func() {
			cmd.Wait()
		}()
		ctx.SplitPrompt.Destroy()
		ctx.SplitPrompt = nil
	}

	ctx.SplitPrompt.Show(ctx.Screens[0].ToXRect(), "Command:", resp, canc)
	return attachF
}

func RegisterSplitHooks(ctx *frame.Context) error {

	var err error
	// Builting shortcuts
	for k, v := range ctx.Config.BuiltinCommands {
		ncmd := v // force to not be a reference
		err = keybind.KeyReleaseFun(
			func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
				if ctx.Locked {
					return
				}
				cmd := exec.Command("bash", "-c", ncmd)
				err := cmd.Start()
				if err != nil {
					log.Println(err)
				}
				go func() {
					cmd.Wait()
				}()
			}).Connect(ctx.X, ctx.X.RootWin(), k.Data, true)
		if err != nil {
			return err
		}
	}

	// Launch help
	err = keybind.KeyReleaseFun(
		func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
			msgPrompt := prompt.NewMessage(ctx.X, prompt.DefaultMessageTheme, prompt.DefaultMessageConfig)
		timeout := 4 * time.Second
		msgPrompt.Show(ctx.Screens[0].ToXRect(), GenerateHelp(ctx), timeout, func(msg *prompt.Message) {})

	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.LaunchHelp, true)


	// Standalone launchers, not great not terrible
	err = keybind.KeyReleaseFun(func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
		if ctx.Locked {
			return
		}

		inPrompt := prompt.NewInput(X, prompt.DefaultInputTheme, prompt.DefaultInputConfig)

		canc := func(inp *prompt.Input) {
			// Chill
		}

		resp := func(inp *prompt.Input, text string) {
			cmd := exec.Command("bash", "-c", text)
			err = cmd.Start()
			if err != nil {
				log.Println(err)
			}
			go func() {
				cmd.Wait()
			}()
			inPrompt.Destroy()
		}

		inPrompt.Show(ctx.Screens[0].ToXRect(), "Command:", resp, canc)

	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.RunCmd.Data, true)
	if err != nil {
		return err
	}

	// Split launchers
	err = keybind.KeyReleaseFun(func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
		if ctx.Locked {
			return
		}

		fr := Split(ctx)
		if fr == nil {
			return
		}
		ctx.AttachPoint = &frame.AttachTarget{
			Target: fr,
			Type:   frame.HORIZONTAL,
		}
	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.SplitHorizontal.Data, true)
	if err != nil {
		return err
	}

	err = keybind.KeyReleaseFun(func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
		if ctx.Locked {
			return
		}

		fr := Split(ctx)
		if fr == nil {
			return
		}
		ctx.AttachPoint = &frame.AttachTarget{
			Target: fr,
			Type:   frame.VERTICAL,
		}
	}).Connect(ctx.X, ctx.X.RootWin(), ctx.Config.SplitVertical.Data, true)
	if err != nil {
		return err
	}

	for _, v := range ctx.Config.GotoKeys {
		ref := v // capture separately so we can use in closure
		err = keybind.KeyReleaseFun(func(X *xgbutil.XUtil, e xevent.KeyReleaseEvent) {
			if ctx.Locked {
				return
			}

			winId, _ := ctx.Gotos[ref]
			f := ctx.Get(winId)
			if f == nil || f.Container == nil {
				return
			}

			if !f.Container.Hidden {
				ffoc := ctx.GetFocusedFrame()
				if ffoc == nil || ffoc.Container != f.Container {
					f.Container.Raise(ctx)
					f.Focus(ctx)
					return
				}
			}

			f.Container.ChangeMinimizationState(ctx)
		}).Connect(ctx.X, ctx.X.RootWin(), ref, true)
		if err != nil {
			return err
		}
	}

	return err
}
