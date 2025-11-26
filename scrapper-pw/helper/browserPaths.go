package helper

import (
	"os"
	"runtime"
)

type executablePaths struct {
	Windows string
	Linux   string
}

type browserPaths struct {
	Name string
	Path executablePaths
}

var allBrowsers = []browserPaths{
	{Name: "Chrome", Path: executablePaths{
		Windows: `C:\Program Files\Google\Chrome\Application\chrome.exe`,
		Linux:   `/usr/bin/google-chrome`,
	}},
	{Name: "Brave", Path: executablePaths{
		Windows: `C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		Linux:   `/usr/bin/brave-browser`,
	}},
	{Name: "Edge", Path: executablePaths{
		Windows: `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		Linux:   `/usr/bin/microsoft-edge`,
	}},
	{Name: "Chromium", Path: executablePaths{
		Windows: `C:\Program Files (x86)\Chromium\Application\chromium.exe`,
		Linux:   `/usr/bin/chromium`,
	}},
	{Name: "Firefox", Path: executablePaths{
		Windows: `C:\Program Files\Mozilla Firefox\firefox.exe`,
		Linux:   `/usr/bin/firefox`,
	}},
}

func GetAvailableBrowsers() []string {
	var available []string
	osName := runtime.GOOS
	var path string

	for _, b := range allBrowsers {

		if osName == "windows" {
			path = b.Path.Windows
		} else {
			path = b.Path.Linux
		}

		if _, err := os.Stat(path); err == nil {
			available = append(available, path)
		}
	}

	return available
}
