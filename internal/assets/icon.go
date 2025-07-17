package assets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// AppIcon contains the embedded application icon
var AppIcon fyne.Resource

func init() {
	// Embed the SVG icon data
	AppIcon = &fyne.StaticResource{
		StaticName: "icon.svg",
		StaticContent: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="256" height="256" viewBox="0 0 256 256" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <!-- Modern gradient background -->
    <linearGradient id="bgGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#4A90E2;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#357ABD;stop-opacity:1" />
    </linearGradient>

    <!-- Clean icon gradient -->
    <linearGradient id="iconGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#FFFFFF;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#F8F9FA;stop-opacity:1" />
    </linearGradient>
  </defs>

  <!-- Background circle with modern gradient -->
  <circle cx="128" cy="128" r="120" fill="url(#bgGradient)" stroke="#2C5282" stroke-width="3"/>

  <!-- Central cleaning symbol - simplified broom/brush -->
  <g transform="translate(128, 128)">
    <!-- Broom handle -->
    <rect x="-3" y="-60" width="6" height="80" fill="url(#iconGradient)" rx="3"/>

    <!-- Broom head -->
    <ellipse cx="0" cy="25" rx="20" ry="10" fill="url(#iconGradient)"/>

    <!-- Bristles -->
    <g stroke="url(#iconGradient)" stroke-width="2" opacity="0.9">
      <line x1="-15" y1="30" x2="-15" y2="40"/>
      <line x1="-8" y1="29" x2="-8" y2="42"/>
      <line x1="0" y1="28" x2="0" y2="43"/>
      <line x1="8" y1="29" x2="8" y2="42"/>
      <line x1="15" y1="30" x2="15" y2="40"/>
    </g>
  </g>

  <!-- Data/telemetry symbols being cleaned -->
  <g opacity="0.6">
    <!-- Data dots -->
    <circle cx="80" cy="80" r="3" fill="#FF6B6B"/>
    <circle cx="90" cy="70" r="2" fill="#FF6B6B"/>
    <circle cx="70" cy="90" r="2.5" fill="#FF6B6B"/>

    <!-- More data dots on the right -->
    <circle cx="180" cy="180" r="3" fill="#FF6B6B"/>
    <circle cx="170" cy="190" r="2" fill="#FF6B6B"/>
    <circle cx="190" cy="170" r="2.5" fill="#FF6B6B"/>
  </g>

  <!-- Cleaning motion lines -->
  <g stroke="rgba(255,255,255,0.7)" stroke-width="2" opacity="0.8" fill="none">
    <path d="M100 100 Q120 90 140 100"/>
    <path d="M110 180 Q130 170 150 180"/>
  </g>

  <!-- Simple "C" for Cleaner in bottom right -->
  <text x="200" y="220" font-family="Arial, sans-serif" font-size="24" font-weight="bold" fill="rgba(255,255,255,0.8)">C</text>
</svg>`),
	}
}

// GetAppIcon returns the application icon resource
func GetAppIcon() fyne.Resource {
	return AppIcon
}

// GetIconURI returns the icon as a URI for use in various contexts
func GetIconURI() fyne.URI {
	return storage.NewFileURI("assets/icon.svg")
}
