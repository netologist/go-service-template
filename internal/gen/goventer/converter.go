package goventer

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/netologist/go-service-template/internal/domain"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
	"github.com/netologist/go-service-template/internal/gen/openapi"
	"github.com/netologist/go-service-template/internal/gen/sqlc"
)

// goverter:converter
// goverter:output:file ./item_converter.gen.go
// goverter:output:package github.com/netologist/go-service-template/internal/gen/goventer
// goverter:extend UUIDToString TimeToTimestamp
// goverter:skipCopySameType
// goverter:ignoreUnexported
type ItemConverter interface {
	// HTTP
	// goverter:map ID Id
	ToOpenAPIItem(domain.Item) openapi.Item
	ToOpenAPIItemList([]domain.Item) []openapi.Item

	// gRPC
	// goverter:map ID Id
	ToProtoItem(domain.Item) *itemv1.Item
	ToProtoItemList([]domain.Item) []*itemv1.Item

	// SQLC / Postgres
	ToDomainItem(sqlc.Item) domain.Item
	ToDomainItemList([]sqlc.Item) []domain.Item
}

func UUIDToString(id uuid.UUID) string {
	return id.String()
}

func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
