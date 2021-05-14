package utils

import (
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type BidiPipe struct {
	r io.Reader
	w io.Writer
}

func NewBidiPipe(r io.Reader, w io.Writer) *BidiPipe {
	return &BidiPipe{
		w: w,
		r: r,
	}
}

func (b *BidiPipe) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

func (b *BidiPipe) Write(p []byte) (n int, err error) {
	return b.w.Write(p)
}

func NewTestConnection() (*BidiPipe, *BidiPipe) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	dataStream1 := &BidiPipe{
		r: r1,
		w: w2,
	}
	dataStream2 := &BidiPipe{
		r: r2,
		w: w1,
	}
	return dataStream1, dataStream2
}

type FakeStream struct {
	id string
}

func NewFakeStream(id string) *FakeStream {
	return &FakeStream{id: id}
}

func (f *FakeStream) Read([]byte) (int, error)         { return 0, nil }
func (f *FakeStream) Write([]byte) (int, error)        { return 0, nil }
func (f *FakeStream) Close() error                     { return nil }
func (f *FakeStream) CloseWrite() error                { return nil }
func (f *FakeStream) CloseRead() error                 { return nil }
func (f *FakeStream) Reset() error                     { return nil }
func (f *FakeStream) SetDeadline(time.Time) error      { return nil }
func (f *FakeStream) SetReadDeadline(time.Time) error  { return nil }
func (f *FakeStream) SetWriteDeadline(time.Time) error { return nil }
func (f *FakeStream) ID() string                       { return f.id }
func (f *FakeStream) Protocol() protocol.ID            { return "ALLEIN" }
func (f *FakeStream) SetProtocol(id protocol.ID) error { return nil }
func (f *FakeStream) Stat() network.Stats              { return network.Stats{} }
func (f *FakeStream) Conn() network.Conn               { return nil }
func (f *FakeStream) Scope() network.StreamScope       { return nil }
