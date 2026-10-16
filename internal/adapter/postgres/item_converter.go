package postgres

import (
	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/sqlc"
)

// goverter:converter
// goverter:output:file ./item_converter.gen.go
// goverter:skipCopySameType
type ItemConverter interface {
	ToDomainItem(sqlc.Item) domain.Item
	ToDomainItemList([]sqlc.Item) []domain.Item
}
