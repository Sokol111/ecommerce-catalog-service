package attribute

import "errors"

var (
	ErrSlugAlreadyExists    = errors.New("attribute with this slug already exists")
	ErrInvalidAttributeData = errors.New("invalid attribute data")
)
