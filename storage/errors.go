package storage

type NotFoundError struct {
	Ref string
}

func (err NotFoundError) Error() string {
	return "Could not find entry with reference: " + err.Ref
}
