package prp

import (
	"sync"
	"time"

	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/common/models"
)

type PRPEntry struct {
	sync.Mutex

	Machine  *models.NetworkNode
	LastSeen time.Time
}

type PRPTableType struct {
	sync.Mutex

	Table           Table
	LastLookupTable TimeTable
	localIP         string
	lastReplySent   time.Time
}

func NewPRPTable() *PRPTableType {
	return &PRPTableType{
		Table:           make(map[string]*PRPEntry, 256),
		LastLookupTable: make(map[string]time.Time, 256),
		lastReplySent:   time.Now().Add(-10 * time.Second),
	}
}

func (e *PRPEntry) UpdateLastSeen() {
	e.Lock()
	defer e.Unlock()
	e.LastSeen = time.Now()
}

// Returns Machine, isFound and if it was Queried Less Than 5 Seconds Ago
func (t *PRPTableType) Lookup(ip string) (*models.NetworkNode, bool, bool) {
	t.Lock()
	defer t.Unlock()
	if e, ok := t.Table.Get(ip); ok {
		e.UpdateLastSeen()
		return e.Machine, ok, false
	}
	if last, ok := t.LastLookupTable.Get(ip); ok {
		if time.Since(last) < 5*time.Second {
			t.LastLookupTable.Put(ip, time.Now())
			return nil, false, true
		}
	}
	return nil, false, false
}

func (t *PRPTableType) Myself() (string, *models.NetworkNode) {
	t.Lock()
	defer t.Unlock()
	if e, ok := t.Table.Get(t.localIP); ok {
		e.UpdateLastSeen()
		return t.localIP, e.Machine
	}
	return t.localIP, nil
}

func (t *PRPTableType) insertEntry(ip string, m *models.NetworkNode) {
	t.Lock()
	defer t.Unlock()
	t.Table.Put(ip, &PRPEntry{Machine: m, LastSeen: time.Now()})
}

func (t *PRPTableType) InsertMyselfEntry(m *models.NetworkNode) {
	t.localIP = m.IP
	t.insertEntry(t.localIP, m)
}

func (t *PRPTableType) PRPReplyMyself(always bool) protocol.Payload {
	// Return nil if we already replied in the last second
	if time.Since(t.lastReplySent) > 1*time.Second || always {
		ip, myself := t.Myself()
		t.Lock()
		defer t.Unlock()
		t.lastReplySent = time.Now()
		return &PRPacket{PRPReply, *myself, ip}
	}
	return nil
}
