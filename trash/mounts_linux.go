package trash

import (
	"os"
	"strings"
)

// readMountPoints returns user-visible mount points from /proc/self/mountinfo,
// skipping pseudo and system filesystems.
func readMountPoints() []string {
	data, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mp := fields[4]
		if skipMountPoint(mp, seen) {
			continue
		}
		seen[mp] = true
		out = append(out, mp)
	}
	return out
}
