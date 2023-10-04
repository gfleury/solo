package stream_map

import (
	"io"
	"sync"

	"github.com/gfleury/solo/client/crypto/noise"
	"github.com/libp2p/go-libp2p/core/network"
)

type AlleinStream struct {
	NoiseStream noise.NoiseStream
	Stream      io.ReadWriter
}

type AlleinStreamMap struct {
	sync.Mutex
	streamMap map[string]*AlleinStream
}

func NewNoiseStreamMap() *AlleinStreamMap {
	return &AlleinStreamMap{
		streamMap: map[string]*AlleinStream{},
	}
}

func (p *AlleinStreamMap) Get(streamID string) (*AlleinStream, bool) {
	p.Lock()
	defer p.Unlock()
	s, found := p.streamMap[streamID]
	return s, found
}

func (p *AlleinStreamMap) New(streamID string, stream network.Stream) {
	p.Lock()
	defer p.Unlock()
	p.streamMap[streamID] = &AlleinStream{Stream: stream}
}

func (p *AlleinStreamMap) NewWithNoise(streamID string, stream io.ReadWriter, noiseStream noise.NoiseStream) {
	p.Lock()
	defer p.Unlock()
	p.streamMap[streamID] = &AlleinStream{Stream: stream, NoiseStream: noiseStream}
}

func (p *AlleinStreamMap) Put(streamID string, stream *AlleinStream) {
	p.Lock()
	defer p.Unlock()
	p.streamMap[streamID] = stream
}

func (p *AlleinStreamMap) Delete(streamID string) {
	if s, found := p.Get(streamID); found {
		s.NoiseStream = nil
		s.Stream = nil
	}
	p.Lock()
	defer p.Unlock()
	delete(p.streamMap, streamID)
}
