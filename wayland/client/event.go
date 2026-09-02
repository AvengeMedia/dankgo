package client

import (
	"bytes"
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"

	_ "unsafe"
)

// maxFdsOut matches libwayland MAX_FDS_OUT
const maxFdsOut = 28

var oobSpace = unix.CmsgSpace(4 * maxFdsOut)
var ErrCmsgTruncated = errors.New("wayland: control message truncated")

type eventTakesFder interface {
	EventTakesFd(opcode uint32) bool
}

func (ctx *Context) ReadMsg() (senderID uint32, opcode uint32, fd int, msg []byte, err error) {
	fd = -1

	oob := make([]byte, oobSpace)
	header := make([]byte, 8)

	n, oobn, flags, _, err := ctx.conn.ReadMsgUnix(header, oob)
	if err != nil {
		return senderID, opcode, fd, msg, err
	}
	if err := ctx.takeAncillary(oob, oobn, flags, "header"); err != nil {
		return senderID, opcode, fd, msg, fmt.Errorf("ctx.ReadMsg: %w", err)
	}
	if n != 8 {
		return senderID, opcode, fd, msg, fmt.Errorf("ctx.ReadMsg: incorrect number of bytes read for header (n=%d)", n)
	}

	senderID = Uint32(header[:4])
	opcodeAndSize := Uint32(header[4:8])
	opcode = opcodeAndSize & 0xffff
	size := opcodeAndSize >> 16

	msgSize := int(size) - 8
	if msgSize > 0 {
		msg = make([]byte, msgSize)
		if oobn > 0 {
			n, err = ctx.conn.Read(msg)
		} else {
			oob = make([]byte, oobSpace)
			n, oobn, flags, _, err = ctx.conn.ReadMsgUnix(msg, oob)
			if err == nil {
				err = ctx.takeAncillary(oob, oobn, flags, "msg")
			}
		}
		if err != nil {
			return senderID, opcode, fd, msg, fmt.Errorf("ctx.ReadMsg: %w", err)
		}
		if n != msgSize {
			return senderID, opcode, fd, msg, fmt.Errorf("ctx.ReadMsg: incorrect number of bytes read for msg (n=%d, msgSize=%d)", n, msgSize)
		}
	}

	if p, ok := ctx.objects.Load(senderID); ok {
		if e, ok := p.(eventTakesFder); ok && e.EventTakesFd(opcode) {
			fd = ctx.popFd()
		}
	}

	return senderID, opcode, fd, msg, nil
}

func (ctx *Context) takeAncillary(oob []byte, oobn, flags int, source string) error {
	if flags&unix.MSG_CTRUNC != 0 {
		fds, _ := getFdsFromOob(oob, oobn, source)
		ctx.closeFdList(fds)
		return ErrCmsgTruncated
	}
	if oobn == 0 {
		return nil
	}
	fds, err := getFdsFromOob(oob, oobn, source)
	if err != nil {
		return err
	}
	ctx.fds = append(ctx.fds, fds...)
	return nil
}

func (ctx *Context) popFd() int {
	if len(ctx.fds) == 0 {
		return -1
	}
	fd := ctx.fds[0]
	ctx.fds = ctx.fds[1:]
	if len(ctx.fds) == 0 {
		ctx.fds = nil
	}
	return fd
}

func (ctx *Context) closeFds() {
	ctx.closeFdList(ctx.fds)
	ctx.fds = nil
}

func (ctx *Context) closeFdList(fds []int) {
	for _, fd := range fds {
		_ = unix.Close(fd)
	}
}

func getFdsFromOob(oob []byte, oobn int, source string) ([]int, error) {
	if oobn == 0 {
		return nil, nil
	}
	if oobn > len(oob) {
		return nil, fmt.Errorf("getFdsFromOob: incorrect number of bytes read from %s for oob (oobn=%d)", source, oobn)
	}
	scms, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, fmt.Errorf("getFdsFromOob: unable to parse control message from %s: %w", source, err)
	}

	var fdsRet []int
	for _, scm := range scms {
		fds, err := unix.ParseUnixRights(&scm)
		if err != nil {
			for _, fd := range fdsRet {
				_ = unix.Close(fd)
			}
			return nil, fmt.Errorf("getFdsFromOob: unable to parse unix rights from %s: %w", source, err)
		}
		fdsRet = append(fdsRet, fds...)
	}

	return fdsRet, nil
}

func Uint32(src []byte) uint32 {
	_ = src[3]
	return *(*uint32)(unsafe.Pointer(&src[0]))
}

func String(src []byte) string {
	idx := bytes.IndexByte(src, 0)
	src = src[:idx:idx]
	return *(*string)(unsafe.Pointer(&src))
}

func Fixed(src []byte) float64 {
	_ = src[3]
	fx := *(*int32)(unsafe.Pointer(&src[0]))
	return fixedToFloat64(fx)
}
