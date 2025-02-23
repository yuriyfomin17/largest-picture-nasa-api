package domain

import "errors"

var (
	ErrNotFound                  = errors.New("not found")
	ErrCalculationLargestPicture = errors.New("error calculating largest picture")
)
