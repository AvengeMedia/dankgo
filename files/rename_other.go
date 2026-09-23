//go:build !linux

package files

func renameNoReplace(from, to string) error { return renameIfAbsent(from, to) }
