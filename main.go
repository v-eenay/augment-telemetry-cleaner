package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"augment-telemetry-cleaner/internal/gui"
	"augment-telemetry-cleaner/internal/resources"
)

func main() {
	myApp := app.NewWithID("com.vinaykoirala.augmenttelemetrycleaner")
	myApp.SetIcon(resources.ResourceIconPng)

	mainWindow := myApp.NewWindow("Augment Telemetry Cleaner v2.0.0")
	mainWindow.SetIcon(resources.ResourceIconPng)
	mainWindow.Resize(fyne.NewSize(800, 700))
	mainWindow.CenterOnScreen()

	// Create the main GUI
	mainGUI := gui.NewMainGUI(mainWindow)
	if mainGUI == nil {
		return // Error already shown in NewMainGUI
	}
	mainWindow.SetContent(mainGUI.BuildUI())

	mainWindow.ShowAndRun()
}
