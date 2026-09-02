package client

import (
	"errors"
	"io"
	"net"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

type fdEventProxy struct{ BaseProxy }

func (*fdEventProxy) EventTakesFd(uint32) bool { return true }

type noFdProxy struct{ BaseProxy }

func TestCmsgSpaceFitsTwoFds(t *testing.T) {
	// CMSG_SPACE aligns to word size. On 64-bit Linux one-fd and two-fd
	// control messages occupy the same 24 bytes, so CmsgSpace(4) does
	// not truncate a 2-fd batch — both descriptors arrive on the first
	// recvmsg and today's fds[0]-only path leaks the second.
	if unix.CmsgSpace(4) != unix.CmsgSpace(8) {
		t.Skip("this platform's cmsg alignment does not fold 1-fd and 2-fd sizes")
	}
}

func TestReadMsgBatchedFds(t *testing.T) {
	ctx, peer := testConn(t)
	p := &fdEventProxy{}
	ctx.RegisterWithID(p, 1)

	w1, r1 := pipe(t)
	w2, r2 := pipe(t)

	msg1 := wlMsg(1, 0, wlString("image/bmp"))
	msg2 := wlMsg(1, 0, wlString("text/plain"))
	payload := append(msg1, msg2...)
	oob := unix.UnixRights(w1, w2)
	if _, oobn, err := peer.WriteMsgUnix(payload, oob, nil); err != nil {
		t.Fatalf("WriteMsgUnix: %v", err)
	} else if oobn != len(oob) {
		t.Fatalf("WriteMsgUnix oobn=%d want %d", oobn, len(oob))
	}
	unix.Close(w1)
	unix.Close(w2)

	_, _, fd1, data1, err := ctx.ReadMsg()
	if err != nil {
		t.Fatalf("first ReadMsg: %v", err)
	}
	_, _, fd2, data2, err := ctx.ReadMsg()
	if err != nil {
		t.Fatalf("second ReadMsg: %v", err)
	}

	if fd1 < 0 || fd2 < 0 {
		t.Fatalf("batched fds: fd1=%d fd2=%d (second must not be -1)", fd1, fd2)
	}
	if got := String(data1[4:]); got != "image/bmp" {
		t.Fatalf("first mime: %q", got)
	}
	if got := String(data2[4:]); got != "text/plain" {
		t.Fatalf("second mime: %q", got)
	}

	if _, err := unix.Write(fd1, []byte("IMG")); err != nil {
		t.Fatalf("write fd1: %v", err)
	}
	if _, err := unix.Write(fd2, []byte("TXT")); err != nil {
		t.Fatalf("write fd2: %v", err)
	}
	unix.Close(fd1)
	unix.Close(fd2)
	got1 := make([]byte, 3)
	got2 := make([]byte, 3)
	if _, err := io.ReadFull(r1, got1); err != nil {
		t.Fatalf("read r1: %v", err)
	}
	if _, err := io.ReadFull(r2, got2); err != nil {
		t.Fatalf("read r2: %v", err)
	}
	if string(got1) != "IMG" || string(got2) != "TXT" {
		t.Fatalf("pipe contents %q %q", got1, got2)
	}
}

func TestReadMsgDoesNotStealFdForEventWithoutFd(t *testing.T) {
	ctx, peer := testConn(t)
	ctx.RegisterWithID(&noFdProxy{}, 1)
	ctx.RegisterWithID(&fdEventProxy{}, 2)

	w, r := pipe(t)
	// delete_id-shaped message (object 1, no fd) followed by a send
	// (object 2, one fd), one sendmsg — the compositor-flush case.
	msg1 := wlMsg(1, 1, putU32(7))
	msg2 := wlMsg(2, 0, wlString("text/plain"))
	if _, _, err := peer.WriteMsgUnix(append(msg1, msg2...), unix.UnixRights(w), nil); err != nil {
		t.Fatalf("WriteMsgUnix: %v", err)
	}
	unix.Close(w)

	_, _, fd1, _, err := ctx.ReadMsg()
	if err != nil {
		t.Fatalf("first ReadMsg: %v", err)
	}
	_, _, fd2, _, err := ctx.ReadMsg()
	if err != nil {
		t.Fatalf("second ReadMsg: %v", err)
	}
	if fd1 != -1 {
		unix.Close(fd1)
		t.Fatalf("non-fd event stole fd=%d", fd1)
	}
	if fd2 < 0 {
		t.Fatal("fd event got -1 after a preceding non-fd event in the same batch")
	}
	if _, err := unix.Write(fd2, []byte("ok")); err != nil {
		t.Fatalf("write: %v", err)
	}
	unix.Close(fd2)
	got := make([]byte, 2)
	if _, err := io.ReadFull(r, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("pipe %q", got)
	}
}

func TestReadMsgReportsCtrunc(t *testing.T) {
	orig := oobSpace
	oobSpace = unix.CmsgSpace(4) // fits two fds on 64-bit, not three
	t.Cleanup(func() { oobSpace = orig })

	ctx, peer := testConn(t)
	ctx.RegisterWithID(&fdEventProxy{}, 1)

	w1, _ := pipe(t)
	w2, _ := pipe(t)
	w3, _ := pipe(t)
	t.Cleanup(func() {
		unix.Close(w1)
		unix.Close(w2)
		unix.Close(w3)
	})

	payload := append(append(wlMsg(1, 0, nil), wlMsg(1, 0, nil)...), wlMsg(1, 0, nil)...)
	if _, _, err := peer.WriteMsgUnix(payload, unix.UnixRights(w1, w2, w3), nil); err != nil {
		t.Fatalf("WriteMsgUnix: %v", err)
	}

	_, _, fd, _, err := ctx.ReadMsg()
	if fd != -1 {
		unix.Close(fd)
	}
	if !errors.Is(err, ErrCmsgTruncated) {
		t.Fatalf("err=%v, want ErrCmsgTruncated", err)
	}
}

func TestReadMsgClosesQueuedFdsOnClose(t *testing.T) {
	ctx, peer := testConn(t)
	// No proxy registered → event does not pop. The fd must not leak
	// across Close.
	w, _ := pipe(t)
	if _, _, err := peer.WriteMsgUnix(wlMsg(99, 0, nil), unix.UnixRights(w), nil); err != nil {
		t.Fatalf("WriteMsgUnix: %v", err)
	}
	unix.Close(w)

	_, _, fd, _, err := ctx.ReadMsg()
	if err != nil {
		t.Fatalf("ReadMsg: %v", err)
	}
	if fd != -1 {
		unix.Close(fd)
		t.Fatalf("unregistered object popped fd=%d", fd)
	}
	if len(ctx.fds) != 1 {
		t.Fatalf("queued %d fds, want 1", len(ctx.fds))
	}
	queued := ctx.fds[0]
	if err := ctx.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := unix.Close(queued); err == nil {
		t.Fatal("queued fd still open after Context.Close")
	}
}

func testConn(t *testing.T) (*Context, *net.UnixConn) {
	t.Helper()
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	f0 := os.NewFile(uintptr(fds[0]), "client")
	f1 := os.NewFile(uintptr(fds[1]), "server")
	c0, err := net.FileConn(f0)
	if err != nil {
		t.Fatal(err)
	}
	c1, err := net.FileConn(f1)
	if err != nil {
		t.Fatal(err)
	}
	f0.Close()
	f1.Close()
	client := c0.(*net.UnixConn)
	server := c1.(*net.UnixConn)
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	return &Context{conn: client}, server
}

func pipe(t *testing.T) (writeFd int, r *os.File) {
	t.Helper()
	rd, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rd.Close() })
	t.Cleanup(func() { wr.Close() })
	return int(wr.Fd()), rd
}

func wlMsg(id, opcode uint32, payload []byte) []byte {
	size := 8 + len(payload)
	b := make([]byte, size)
	PutUint32(b[0:4], id)
	PutUint32(b[4:8], opcode|uint32(size)<<16)
	copy(b[8:], payload)
	return b
}

func wlString(s string) []byte {
	n := PaddedLen(len(s) + 1)
	b := make([]byte, 4+n)
	PutString(b, s)
	return b
}

func putU32(v uint32) []byte {
	b := make([]byte, 4)
	PutUint32(b, v)
	return b
}
