package main

import (
	_ "embed"
	"fmt"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"os"
)

//go:embed ui/main.ui
var mainUi string

//go:embed ui/error.ui
var errorUi string

var applicationWindow *adw.ApplicationWindow
var alertDialog *adw.AlertDialog
var errorBox *gtk.Label

func showError(err string) {
	errorBox.SetText(err)
	alertDialog.Present(applicationWindow)
}

func main() {
	app := adw.NewApplication("com.github.arzumify.waymux", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() {
		activate(app)

		err := ConnectToSocket()
		if err != nil {
			showError("Could not connect to socket: " + err.Error())
		}
	})

	code := app.Run(os.Args)
	if code != 0 {
		os.Exit(code)
	}
}

func activate(app *adw.Application) {
	builder := gtk.NewBuilder()
	err := builder.AddFromString(mainUi)
	if err != nil {
		panic(fmt.Errorf("failed to load main window: %v", err))
	}

	var ok bool
	applicationWindow, ok = builder.GetObject("MainWindow").Cast().(*adw.ApplicationWindow)
	if !ok {
		panic("Could not find main window")
	}

	app.AddWindow(&applicationWindow.Window)
	applicationWindow.SetDefaultSize(512, 128)
	applicationWindow.Widget.SetVisible(true)

	builder = gtk.NewBuilder()
	err = builder.AddFromString(errorUi)
	if err != nil {
		panic(fmt.Errorf("failed to load error dialog: %v", err))
	}

	alertDialog, ok = builder.GetObject("MainWindow").Cast().(*adw.AlertDialog)
	if !ok {
		panic("Could not find error dialog")
	}

	errorBox, ok = builder.GetObject("Error").Cast().(*gtk.Label)
	if !ok {
		panic("Could not find error box")
	}

	dismiss, ok := builder.GetObject("Dismiss").Cast().(*gtk.Button)
	if !ok {
		panic("Could not find dismiss box")
	}

	dismiss.Connect("clicked", func() {
		alertDialog.Close()
	})
}
