package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	MimeDirectory = "inode/directory"
	MimeSymlink   = "inode/symlink"
	MimeDesktop   = "application/x-desktop"
)

type Entry struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	IsDir         bool   `json:"isDir"`
	IsSymlink     bool   `json:"isSymlink"`
	SymlinkTarget string `json:"symlinkTarget"`
	SymlinkBroken bool   `json:"symlinkBroken"`
	Size          int64  `json:"size"`
	MtimeMs       int64  `json:"mtimeMs"`
	CtimeMs       int64  `json:"ctimeMs"`
	AtimeMs       int64  `json:"atimeMs"`
	Mode          string `json:"mode"`
	Owner         string `json:"owner"`
	Group         string `json:"group"`
	Hidden        bool   `json:"hidden"`
	Executable    bool   `json:"isExecutable"`
	Extension     string `json:"extension"`
	Mime          string `json:"mime"`
	IconName      string `json:"iconName"`
	Thumbnail     string `json:"thumbnail"`
	Thumbnailable bool   `json:"thumbnailable"`
	Unreadable    bool   `json:"unreadable"`
	DisplayName   string `json:"displayName"`
	Untrusted     bool   `json:"untrusted"`

	nameKey []byte
}

func (b *builder) entry(dir, name string, hidden bool) (Entry, error) {
	full := filepath.Join(dir, name)
	info, err := os.Lstat(full)
	if err != nil {
		return Entry{}, err
	}

	e := Entry{
		Name:      name,
		Path:      full,
		Hidden:    hidden || strings.HasPrefix(name, "."),
		Extension: strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")),
	}
	b.fill(&e, info, full)
	return e, nil
}

func (b *builder) unreadable(dir, name string, hidden bool) Entry {
	return Entry{
		Name:       name,
		Path:       filepath.Join(dir, name),
		Hidden:     hidden || strings.HasPrefix(name, "."),
		Size:       -1,
		Unreadable: true,
		IconName:   "emblem-unreadable",
		nameKey:    b.keys.key(name),
	}
}

func (b *builder) fill(e *Entry, info fs.FileInfo, full string) {
	e.Mode = info.Mode().String()
	e.MtimeMs = info.ModTime().UnixMilli()
	e.CtimeMs, e.AtimeMs, e.Owner, e.Group = b.ownership(info)
	e.Size = info.Size()
	e.IsDir = info.IsDir()

	if info.Mode()&fs.ModeSymlink != 0 {
		e.IsSymlink = true
		e.SymlinkTarget = b.linkTarget(full)
		target, err := os.Stat(full)
		if err != nil {
			e.SymlinkBroken = true
			e.Size = -1
			e.Mime = MimeSymlink
		} else {
			e.IsDir = target.IsDir()
			e.Size = target.Size()
			e.Executable = !target.IsDir() && target.Mode()&0o111 != 0
		}
	} else {
		e.Executable = !e.IsDir && info.Mode()&0o111 != 0
	}

	if e.IsDir {
		e.Size = -1
		e.Mime = MimeDirectory
	}
	if e.Mime == "" {
		e.Mime = b.mime.FromName(e.Name)
	}
	e.Untrusted = e.Mime == MimeDesktop && !e.Executable
	e.Thumbnailable = b.supports(e.Mime)
	e.IconName = b.icons.best(*e)
	e.nameKey = b.keys.key(e.Name)
}

func (b *builder) linkTarget(full string) string {
	target, err := os.Readlink(full)
	if err != nil {
		return ""
	}
	if filepath.IsAbs(target) {
		return target
	}
	return filepath.Join(filepath.Dir(full), target)
}

func (b *builder) ownership(info fs.FileInfo) (ctimeMs, atimeMs int64, owner, group string) {
	extra, ok := extraStat(info)
	if !ok {
		return info.ModTime().UnixMilli(), info.ModTime().UnixMilli(), "", ""
	}
	owner, group = b.owners.names(extra.uid, extra.gid)
	return extra.ctimeMs, extra.atimeMs, owner, group
}
