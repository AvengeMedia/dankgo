package trash

import "golang.org/x/sys/unix"

// readMountPoints returns user-visible mount points via getfsstat(2),
// skipping pseudo and system filesystems.
func readMountPoints() []string {
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil || n == 0 {
		return nil
	}
	stats := make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(stats, unix.MNT_NOWAIT)
	if err != nil {
		return nil
	}

	var out []string
	seen := map[string]bool{}
	for _, st := range stats[:n] {
		mp := unix.ByteSliceToString(st.Mntonname[:])
		if mp == "" || skipMountPoint(mp, seen) {
			continue
		}
		seen[mp] = true
		out = append(out, mp)
	}
	return out
}
