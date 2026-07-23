// Package portal provides small clients for the XDG desktop portal
// interfaces (org.freedesktop.portal.Desktop) used by sandboxed dank apps.
package portal

import "os"

const (
	busName    = "org.freedesktop.portal.Desktop"
	objectPath = "/org/freedesktop/portal/desktop"
)

// InFlatpak reports whether the process runs inside a Flatpak sandbox.
func InFlatpak() bool {
	if os.Getenv("FLATPAK_ID") != "" {
		return true
	}
	_, err := os.Stat("/.flatpak-info")
	return err == nil
}
