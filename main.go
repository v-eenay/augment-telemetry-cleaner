package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"augment-telemetry-cleaner/internal/assets"
	"augment-telemetry-cleaner/internal/gui"
)

func main() {
	myApp := app.NewWithID("com.vinaykoirala.augmenttelemetrycleaner")

	// Set application icon
	myApp.SetIcon(assets.GetAppIcon())

	mainWindow := myApp.NewWindow("Augment Telemetry Cleaner v1.0.0")
	mainWindow.Resize(fyne.NewSize(1000, 700))
	mainWindow.CenterOnScreen()

	// Set window icon
	mainWindow.SetIcon(assets.GetAppIcon())

	// Create the main GUI
	mainGUI := gui.NewMainGUI(mainWindow)
	if mainGUI == nil {
		return // Error already shown in NewMainGUI
	}
	mainWindow.SetContent(mainGUI.BuildUI())

	mainWindow.ShowAndRun()
}
