package files

import (
	"os"
	"path/filepath"
	"strings"
)

var userDirIcons = map[string]string{
	"XDG_DESKTOP_DIR":     "user-desktop",
	"XDG_DOCUMENTS_DIR":   "folder-documents",
	"XDG_DOWNLOAD_DIR":    "folder-download",
	"XDG_MUSIC_DIR":       "folder-music",
	"XDG_PICTURES_DIR":    "folder-pictures",
	"XDG_PUBLICSHARE_DIR": "folder-publicshare",
	"XDG_TEMPLATES_DIR":   "folder-templates",
	"XDG_VIDEOS_DIR":      "folder-videos",
}

var userDirOrder = []string{
	"XDG_DESKTOP_DIR",
	"XDG_DOCUMENTS_DIR",
	"XDG_DOWNLOAD_DIR",
	"XDG_MUSIC_DIR",
	"XDG_PICTURES_DIR",
	"XDG_VIDEOS_DIR",
	"XDG_PUBLICSHARE_DIR",
	"XDG_TEMPLATES_DIR",
}

type UserDir struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	IconName string `json:"iconName"`
}

func loadUserDirPaths() map[string]string {
	paths := map[string]string{}
	for key := range userDirIcons {
		if path := os.Getenv(key); path != "" {
			paths[key] = filepath.Clean(path)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return paths
	}
	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(home, ".config")
	}
	data, err := os.ReadFile(filepath.Join(config, "user-dirs.dirs"))
	if err != nil {
		return paths
	}

	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := parseUserDir(line)
		if !ok {
			continue
		}
		if _, known := userDirIcons[key]; !known {
			continue
		}
		if _, fromEnv := paths[key]; fromEnv {
			continue
		}
		paths[key] = value
	}
	return paths
}

func loadUserDirs() map[string]string {
	dirs := map[string]string{}
	if home, err := os.UserHomeDir(); err == nil {
		dirs[home] = "user-home"
	}
	for key, path := range loadUserDirPaths() {
		dirs[path] = userDirIcons[key]
	}
	return dirs
}

func userDirKey(envKey string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(envKey, "XDG_"), "_DIR"))
}

func listUserDirs() []UserDir {
	paths := loadUserDirPaths()

	out := make([]UserDir, 0, len(userDirOrder)+1)
	if home, err := os.UserHomeDir(); err == nil {
		out = append(out, UserDir{Key: "home", Name: filepath.Base(home), Path: home, IconName: "user-home"})
	}
	for _, envKey := range userDirOrder {
		path, ok := paths[envKey]
		if !ok {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		out = append(out, UserDir{Key: userDirKey(envKey), Name: filepath.Base(path), Path: path, IconName: userDirIcons[envKey]})
	}
	return out
}

func parseUserDir(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") {
		return "", "", false
	}
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if !strings.HasPrefix(value, "$HOME") {
		if !filepath.IsAbs(value) {
			return "", "", false
		}
		return strings.TrimSpace(key), filepath.Clean(value), true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", false
	}
	return strings.TrimSpace(key), filepath.Clean(filepath.Join(home, strings.TrimPrefix(value, "$HOME"))), true
}
