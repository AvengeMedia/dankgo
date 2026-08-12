//go:build linux

package wlclipboard

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AvengeMedia/dankgo/wayland/ext_data_control"
)

// ErrNoData reports an empty selection: no client owns the clipboard or the
// owner advertises no readable mime type.
var ErrNoData = errors.New("no clipboard data")

type Offer struct {
	MimeType string
	Data     []byte
}

// textMimeAliases are offered alongside plain-text content so legacy X11
// clients bridged through XWayland find a target they can convert.
var textMimeAliases = []string{
	"text/plain",
	"text/plain;charset=utf-8",
	"UTF8_STRING",
	"STRING",
	"TEXT",
}

// ExpandOffers turns raw clipboard data into the full offer list to serve,
// adding the standard alias set for text content.
func ExpandOffers(data []byte, mimeType string) []Offer {
	offers := []Offer{{MimeType: mimeType, Data: data}}
	if mimeType != "text/plain" && mimeType != "text/plain;charset=utf-8" {
		return offers
	}
	for _, alias := range textMimeAliases {
		if alias == mimeType {
			continue
		}
		offers = append(offers, Offer{MimeType: alias, Data: data})
	}
	return offers
}

// TextOffers is the offer list for plain text: UTF-8 text/plain plus the
// legacy X11 aliases.
func TextOffers(text string) []Offer {
	return ExpandOffers([]byte(text), "text/plain;charset=utf-8")
}

func Paste() ([]byte, string, error) {
	s, err := connectSession()
	if err != nil {
		return nil, "", err
	}
	defer s.Close()

	dataControlMgr, err := s.requireDataControl()
	if err != nil {
		return nil, "", err
	}

	device, err := dataControlMgr.GetDataDevice(s.seat)
	if err != nil {
		return nil, "", fmt.Errorf("get data device: %w", err)
	}
	defer device.Destroy()

	offerMimeTypes := make(map[*ext_data_control.ExtDataControlOfferV1][]string)

	device.SetDataOfferHandler(func(e ext_data_control.ExtDataControlDeviceV1DataOfferEvent) {
		if e.Id == nil {
			return
		}
		offerMimeTypes[e.Id] = nil
		e.Id.SetOfferHandler(func(me ext_data_control.ExtDataControlOfferV1OfferEvent) {
			offerMimeTypes[e.Id] = append(offerMimeTypes[e.Id], me.MimeType)
		})
	})

	var selectionOffer *ext_data_control.ExtDataControlOfferV1
	gotSelection := false

	device.SetSelectionHandler(func(e ext_data_control.ExtDataControlDeviceV1SelectionEvent) {
		selectionOffer = e.Id
		gotSelection = true
	})

	s.display.Roundtrip()
	s.display.Roundtrip()

	if !gotSelection || selectionOffer == nil {
		return nil, "", ErrNoData
	}

	selectedMime := selectPreferredMimeType(offerMimeTypes[selectionOffer])
	if selectedMime == "" {
		return nil, "", ErrNoData
	}

	r, w, err := os.Pipe()
	if err != nil {
		return nil, "", fmt.Errorf("create pipe: %w", err)
	}
	defer r.Close()

	if err := selectionOffer.Receive(selectedMime, int(w.Fd())); err != nil {
		w.Close()
		return nil, "", fmt.Errorf("receive: %w", err)
	}
	w.Close()

	s.display.Roundtrip()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("read: %w", err)
	}

	return data, selectedMime, nil
}

func PasteText() (string, error) {
	data, _, err := Paste()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func selectPreferredMimeType(mimes []string) string {
	preferred := []string{
		"text/plain;charset=utf-8",
		"text/plain",
		"UTF8_STRING",
		"STRING",
		"TEXT",
		"image/png",
		"image/jpeg",
	}

	for _, pref := range preferred {
		for _, mime := range mimes {
			if mime == pref {
				return mime
			}
		}
	}

	if len(mimes) > 0 {
		return mimes[0]
	}
	return ""
}
