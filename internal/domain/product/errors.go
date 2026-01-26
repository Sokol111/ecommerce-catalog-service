package product

import "errors"

var (
	ErrInvalidProductData = errors.New("invalid product data")
	ErrCategoryNotFound   = errors.New("category not found")
)
