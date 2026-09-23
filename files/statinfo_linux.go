package files

import (
	"io/fs"
	"syscall"
)

type statExtra struct {
	ctimeMs int64
	atimeMs int64
	uid     uint32
	gid     uint32
}

func extraStat(info fs.FileInfo) (statExtra, bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return statExtra{}, false
	}
	return statExtra{
		ctimeMs: st.Ctim.Sec*1000 + st.Ctim.Nsec/1e6,
		atimeMs: st.Atim.Sec*1000 + st.Atim.Nsec/1e6,
		uid:     st.Uid,
		gid:     st.Gid,
	}, true
}
