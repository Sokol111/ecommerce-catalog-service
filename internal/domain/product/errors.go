package product

import "errors"

var (
	ErrInvalidProductData   = errors.New("invalid product data")
	ErrInsufficientQuantity = errors.New("insufficient quantity")
)
