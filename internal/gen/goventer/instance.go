//go:build !goverter

package goventer

func NewItemConverter() ItemConverter {
	return &ItemConverterImpl{}
}
