package assets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// AppIcon contains the embedded application icon
var AppIcon fyne.Resource

func init() {
	// Use Fyne's default icon for now to avoid SVG issues
	AppIcon = nil
}

// GetAppIcon returns the application icon resource
func GetAppIcon() fyne.Resource {
	if AppIcon == nil {
		return fyne.CurrentApp().Icon()
	}
	return AppIcon
}

// GetIconURI returns the icon as a URI for use in various contexts
func GetIconURI() fyne.URI {
	return storage.NewFileURI("assets/icon.svg")
}
