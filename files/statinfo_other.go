//go:build !linux

package files

import "io/fs"

type statExtra struct {
	ctimeMs int64
	atimeMs int64
	uid     uint32
	gid     uint32
}

func extraStat(fs.FileInfo) (statExtra, bool) { return statExtra{}, false }
