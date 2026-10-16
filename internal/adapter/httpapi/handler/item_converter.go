package handler

import (
	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/openapi"
)

// goverter:converter
// goverter:output:file ./item_converter.gen.go
// goverter:skipCopySameType
type ItemConverter interface {
	// goverter:map ID Id
	ToOpenAPIItem(domain.Item) openapi.Item
	ToOpenAPIItemList([]domain.Item) []openapi.Item
}
