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
	return parseMountInfo(string(data))
}

func parseMountInfo(data string) []string {
	var out []string
	seen := map[string]bool{}
	for line := range strings.SplitSeq(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mp := fields[4]
		if skipMountPoint(mp, seen) || skipFsType(mountInfoFsType(fields)) {
			continue
		}
		seen[mp] = true
		out = append(out, mp)
	}
	return out
}

func mountInfoFsType(fields []string) string {
	for i := 6; i+1 < len(fields); i++ {
		if fields[i] == "-" {
			return fields[i+1]
		}
	}
	return ""
}
