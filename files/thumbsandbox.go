package files

import (
	"os"
	"path/filepath"
)

const (
	sandboxInputDir = "/tmp"
	sandboxOutput   = "/tmp/dfiles-thumbnail.png"
)

func detectBwrap() string {
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		return ""
	}
	return resolveExecutable("bwrap")
}

func sandboxInputPath(hostInput string) string {
	return filepath.Join(sandboxInputDir, "dfiles-input"+filepath.Ext(hostInput))
}

func bwrapArgs(bwrap, hostInput, hostOutDir, cacheHome, toolDir string) []string {
	args := []string{bwrap,
		"--ro-bind", "/usr", "/usr",
		"--ro-bind-try", "/etc/ld.so.cache", "/etc/ld.so.cache",
		"--ro-bind-try", "/etc/alternatives", "/etc/alternatives",
		"--ro-bind-try", "/etc/fonts", "/etc/fonts",
	}
	for _, dir := range []string{"/bin", "/lib", "/lib64", "/sbin"} {
		target, err := os.Readlink(dir)
		switch err {
		case nil:
			args = append(args, "--symlink", target, dir)
		default:
			args = append(args, "--ro-bind-try", dir, dir)
		}
	}
	for _, dir := range []string{"/nix", "/run/current-system", "/etc/static", "/opt"} {
		args = append(args, "--ro-bind-try", dir, dir)
	}
	fontCache := filepath.Join(cacheHome, "fontconfig")
	gstCache := filepath.Join(cacheHome, "gstreamer-1.0")
	return append(args,
		"--ro-bind-try", fontCache, fontCache,
		"--bind-try", gstCache, gstCache,
		"--bind", hostOutDir, sandboxInputDir,
		"--ro-bind", hostInput, sandboxInputPath(hostInput),
		"--proc", "/proc",
		"--dev", "/dev",
		"--chdir", "/",
		"--clearenv",
		"--setenv", "PATH", toolDir+":/usr/local/bin:/usr/bin:/bin",
		"--setenv", "HOME", sandboxInputDir,
		"--setenv", "XDG_CACHE_HOME", cacheHome,
		"--setenv", "GIO_USE_VFS", "local",
		"--unshare-all",
		"--die-with-parent",
		"--new-session",
		"--",
	)
}
