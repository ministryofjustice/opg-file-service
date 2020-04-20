package storage

type Reader interface {
	getEntry(ref string) *Entry
}
