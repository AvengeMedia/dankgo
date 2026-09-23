package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

func readPNGText(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	signature := make([]byte, len(pngSignature))
	if _, err := io.ReadFull(f, signature); err != nil {
		return nil, err
	}
	if !bytes.Equal(signature, pngSignature) {
		return nil, errors.New("not a png")
	}

	text := map[string]string{}
	header := make([]byte, 8)
	for {
		if _, err := io.ReadFull(f, header); err != nil {
			return text, nil
		}
		length := binary.BigEndian.Uint32(header[:4])
		kind := string(header[4:8])
		if kind == "IDAT" || kind == "IEND" {
			return text, nil
		}
		data := make([]byte, length+4)
		if _, err := io.ReadFull(f, data); err != nil {
			return text, nil
		}
		if kind != "tEXt" {
			continue
		}
		key, value, found := bytes.Cut(data[:length], []byte{0})
		if !found {
			continue
		}
		text[string(key)] = string(value)
	}
}

func writePNGWithText(dst string, img image.Image, text map[string]string) error {
	var encoded bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := encoder.Encode(&encoded, img); err != nil {
		return err
	}

	body := encoded.Bytes()
	headerEnd := len(pngSignature) + 8 + int(binary.BigEndian.Uint32(body[len(pngSignature):])) + 4

	var out bytes.Buffer
	out.Write(body[:headerEnd])
	for key, value := range text {
		writeTextChunk(&out, key, value)
	}
	out.Write(body[headerEnd:])

	return writeFileAtomic(dst, out.Bytes())
}

func writeTextChunk(out *bytes.Buffer, key, value string) {
	payload := make([]byte, 0, len(key)+1+len(value))
	payload = append(payload, key...)
	payload = append(payload, 0)
	payload = append(payload, value...)

	_ = binary.Write(out, binary.BigEndian, uint32(len(payload)))
	chunk := append([]byte("tEXt"), payload...)
	out.Write(chunk)
	_ = binary.Write(out, binary.BigEndian, crc32.ChecksumIEEE(chunk))
}

func writeFileAtomic(dst string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".dfiles-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}
