package files

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func renameNoReplace(from, to string) error {
	err := unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, unix.EINVAL), errors.Is(err, unix.ENOSYS):
		// EINVAL here means the filesystem lacks RENAME_NOREPLACE.
		return renameIfAbsent(from, to)
	default:
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
}
