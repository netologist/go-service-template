package handler

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/netologist/go-service-template/internal/domain"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
)

// goverter:converter
// goverter:output:file ./item_converter.gen.go
// goverter:extend UUIDToString TimeToTimestamp
// goverter:skipCopySameType
// goverter:ignoreUnexported
type ItemConverter interface {
	// goverter:map ID Id
	ToProtoItem(domain.Item) *itemv1.Item
	ToProtoItemList([]domain.Item) []*itemv1.Item
}

func UUIDToString(id uuid.UUID) string {
	return id.String()
}

func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
