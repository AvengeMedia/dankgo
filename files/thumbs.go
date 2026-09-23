package files

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"image"
	"image/draw"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	xdraw "golang.org/x/image/draw"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	imageSizeCap  = 64 << 20
	pixelCap      = 100 << 20
	externalLimit = 20 * time.Second
	failDir       = "fail/dankfiles"
)

var thumbSizes = map[string]int{
	"normal":   128,
	"large":    256,
	"x-large":  512,
	"xx-large": 1024,
}

func ParseThumbSize(raw string) string {
	name := strings.ToLower(strings.TrimSpace(raw))
	if _, ok := thumbSizes[name]; ok {
		return name
	}
	return "normal"
}

type Thumbs struct {
	root  string
	video string
	pdf   string
}

func NewThumbs(cacheHome string) *Thumbs {
	video, _ := exec.LookPath("ffmpegthumbnailer")
	pdf, _ := exec.LookPath("pdftoppm")
	return &Thumbs{root: filepath.Join(cacheHome, "thumbnails"), video: video, pdf: pdf}
}

func (t *Thumbs) Capabilities() []string {
	caps := []string{"thumbnail.image"}
	if t.video != "" {
		caps = append(caps, "thumbnail.video")
	}
	if t.pdf != "" {
		caps = append(caps, "thumbnail.pdf")
	}
	return caps
}

func (t *Thumbs) Supports(mimeType string) bool {
	switch {
	case mimeType == "image/svg+xml" || mimeType == "image/svg":
		return false
	case mediaType(mimeType) == "image":
		return true
	case mediaType(mimeType) == "video":
		return t.video != ""
	case mimeType == "application/pdf":
		return t.pdf != ""
	default:
		return false
	}
}

func (t *Thumbs) Lookup(path string, mtimeMs int64, size string) (string, bool) {
	cached := t.cachePath(path, size)
	if t.valid(cached, path, mtimeMs) {
		return cached, true
	}
	return "", false
}

func (t *Thumbs) Failed(path string, mtimeMs int64) bool {
	return t.valid(t.failPath(path), path, mtimeMs)
}

func (t *Thumbs) Generate(e Entry, size string) (string, error) {
	if !t.Supports(e.Mime) {
		return "", errors.New("no thumbnailer for " + e.Mime)
	}
	if cached, ok := t.Lookup(e.Path, e.MtimeMs, size); ok {
		return cached, nil
	}

	img, err := t.render(e, thumbSizes[size])
	if err != nil {
		t.markFailed(e)
		return "", err
	}

	dst := t.cachePath(e.Path, size)
	if err := writePNGWithText(dst, img, t.metadata(e)); err != nil {
		return "", err
	}
	return dst, nil
}

func (t *Thumbs) render(e Entry, bound int) (image.Image, error) {
	switch {
	case mediaType(e.Mime) == "image":
		if e.Size > imageSizeCap {
			return nil, errors.New("image over the thumbnail size cap")
		}
		return decodeScaled(e.Path, bound)
	case mediaType(e.Mime) == "video":
		return t.external(bound, "", func(dst string) []string {
			return []string{t.video, "-i", e.Path, "-o", dst, "-s", strconv.Itoa(bound)}
		})
	case e.Mime == "application/pdf":
		return t.external(bound, "-1.png", func(dst string) []string {
			return []string{t.pdf, "-png", "-f", "1", "-l", "1", "-scale-to", strconv.Itoa(bound), e.Path, strings.TrimSuffix(dst, ".png")}
		})
	}
	return nil, errors.New("no thumbnailer for " + e.Mime)
}

func (t *Thumbs) external(bound int, producedSuffix string, build func(dst string) []string) (image.Image, error) {
	tmp, err := os.MkdirTemp("", "dfiles-thumb-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	ctx, cancel := context.WithTimeout(context.Background(), externalLimit)
	defer cancel()

	dst := filepath.Join(tmp, "out.png")
	argv := build(dst)
	if err := exec.CommandContext(ctx, argv[0], argv[1:]...).Run(); err != nil {
		return nil, err
	}
	return decodeScaled(strings.TrimSuffix(dst, ".png")+cmp.Or(producedSuffix, ".png"), bound)
}

func decodeScaled(path string, bound int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}
	if config.Width*config.Height > pixelCap {
		return nil, errors.New("image over the pixel cap")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return scaleToBound(src, bound), nil
}

func scaleToBound(src image.Image, bound int) image.Image {
	size := src.Bounds().Size()
	if size.X <= bound && size.Y <= bound {
		return src
	}
	width, height := size.X, size.Y
	switch {
	case width >= height:
		height = max(1, height*bound/width)
		width = bound
	default:
		width = max(1, width*bound/height)
		height = bound
	}
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func (t *Thumbs) metadata(e Entry) map[string]string {
	return map[string]string{
		"Thumb::URI":      fileURI(e.Path),
		"Thumb::MTime":    strconv.FormatInt(e.MtimeMs/1000, 10),
		"Thumb::Size":     strconv.FormatInt(max(e.Size, 0), 10),
		"Thumb::Mimetype": e.Mime,
	}
}

func (t *Thumbs) markFailed(e Entry) {
	marker := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	_ = writePNGWithText(t.failPath(e.Path), marker, t.metadata(e))
}

func (t *Thumbs) valid(cached, path string, mtimeMs int64) bool {
	text, err := readPNGText(cached)
	if err != nil {
		return false
	}
	if text["Thumb::URI"] != fileURI(path) {
		return false
	}
	return text["Thumb::MTime"] == strconv.FormatInt(mtimeMs/1000, 10)
}

func (t *Thumbs) cachePath(path, size string) string {
	return filepath.Join(t.root, size, uriDigest(path)+".png")
}

func (t *Thumbs) failPath(path string) string {
	return filepath.Join(t.root, failDir, uriDigest(path)+".png")
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func uriDigest(path string) string {
	sum := md5.Sum([]byte(fileURI(path)))
	return hex.EncodeToString(sum[:])
}
