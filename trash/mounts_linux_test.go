package trash

import (
	"slices"
	"testing"
)

func TestParseMountInfoSkipsAutofsPlaceholders(t *testing.T) {
	data := "" +
		"25 1 0:22 / / rw,relatime shared:1 - ext4 /dev/nvme0n1p2 rw\n" +
		"60 25 0:40 / /mnt/TrueNAS rw,relatime shared:30 - autofs systemd-1 rw,fd=50,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=1\n" +
		"61 25 0:41 / /mnt/Windows rw,relatime shared:31 - autofs systemd-1 rw,fd=51,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=2\n" +
		"62 61 0:42 / /mnt/Windows rw,relatime shared:32 - ntfs3 /dev/sda1 rw\n" +
		"70 25 0:43 / /run/media/user/USB rw,nosuid,nodev,relatime shared:40 - vfat /dev/sdb1 rw\n"

	got := parseMountInfo(data)
	want := []string{"/mnt/Windows", "/run/media/user/USB"}
	if !slices.Equal(got, want) {
		t.Fatalf("parseMountInfo = %v, want %v", got, want)
	}
}
