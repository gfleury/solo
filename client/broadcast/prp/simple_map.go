package prp

import "time"

type Table map[string]*PRPEntry
type TimeTable map[string]time.Time

func (t Table) Lookup(k string) (*PRPEntry, bool) {
	v, found := t[k]
	return v, found
}

func (t Table) Get(k string) (*PRPEntry, bool) {
	v, found := t[k]
	return v, found
}

func (t Table) Put(k string, v *PRPEntry) {
	t[k] = v
}

func (t TimeTable) Lookup(k string) (time.Time, bool) {
	v, found := t[k]
	return v, found
}

func (t TimeTable) Get(k string) (time.Time, bool) {
	v, found := t[k]
	return v, found
}

func (t TimeTable) Put(k string, v time.Time) {
	t[k] = v
}
