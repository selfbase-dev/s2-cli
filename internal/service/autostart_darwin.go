package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const launchAgentLabel = "dev.selfbase.s2sync"

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist"), nil
}

// IsAutostartEnabled reports whether a launch agent plist is in place.
func IsAutostartEnabled() bool {
	path, err := launchAgentPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// SetAutostart installs or removes the launch agent plist for the
// given executable path. When exePath lives inside a .app bundle, the
// agent invokes `open -a <bundle>` so macOS treats it as a real app
// launch (Dock / LSUIElement / proper NSApplication setup). Falls back
// to launching the raw binary for non-bundled cases.
func SetAutostart(enabled bool, exePath string) error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	if !enabled {
		_ = exec.Command("launchctl", "unload", path).Run()
		return removeIfExists(path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var argsXML string
	if bundle := appBundlePath(exePath); bundle != "" {
		argsXML = fmt.Sprintf("        <string>/usr/bin/open</string>\n        <string>-a</string>\n        <string>%s</string>", bundle)
	} else {
		argsXML = fmt.Sprintf("        <string>%s</string>", exePath)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
%s
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>ProcessType</key>
    <string>Interactive</string>
</dict>
</plist>
`, launchAgentLabel, argsXML)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	// Best-effort unload first (to replace any prior plist), then load.
	_ = exec.Command("launchctl", "unload", path).Run()
	_ = exec.Command("launchctl", "load", path).Run()
	return nil
}

// appBundlePath returns the .app bundle path when exePath is its
// Contents/MacOS/<binary>, otherwise "".
func appBundlePath(exePath string) string {
	macosDir := filepath.Dir(exePath)
	if filepath.Base(macosDir) != "MacOS" {
		return ""
	}
	contentsDir := filepath.Dir(macosDir)
	if filepath.Base(contentsDir) != "Contents" {
		return ""
	}
	return filepath.Dir(contentsDir)
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
