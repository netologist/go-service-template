package e2e_test

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/netologist/go-service-template/pkg/client"
	itemv1 "github.com/netologist/go-service-template/pkg/proto/item/v1"
)

var _ = Describe("gRPC ItemService and Protocol Interoperability", func() {
	Context("grpc.health.v1 Health Check", func() {
		It("reports SERVING status for the service", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := grpcHealthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
				Service: itemv1.ItemService_ServiceDesc.ServiceName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Status).To(Equal(grpc_health_v1.HealthCheckResponse_SERVING))
		})
	})

	Context("ItemService.CreateItem RPC", func() {
		It("registers a new item over gRPC and persists entity correctly", func() {
			itemName := "gRPC Item " + uuid.NewString()[:8]
			desc := "Created via gRPC client"
			price := int64(7500)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:        itemName,
				Description: desc,
				Price:       price,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Item).NotTo(BeNil())
			Expect(resp.Item.Id).NotTo(BeEmpty())
			Expect(resp.Item.Name).To(Equal(itemName))
			Expect(resp.Item.Description).To(Equal(desc))
			Expect(resp.Item.Price).To(Equal(price))
			Expect(resp.Item.CreatedAt).NotTo(BeNil())
		})

		It("returns InvalidArgument when name is too short", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:  "X",
				Price: 1000,
			})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

		It("returns InvalidArgument when price is negative", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:  "Valid Name",
				Price: -100,
			})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

		It("returns AlreadyExists when creating duplicate item name", func() {
			duplicateName := "Duplicate gRPC " + uuid.NewString()[:8]

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// First registration
			_, err := grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:  duplicateName,
				Price: 1000,
			})
			Expect(err).NotTo(HaveOccurred())

			// Second registration should conflict
			_, err = grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:  duplicateName,
				Price: 2000,
			})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.AlreadyExists))
		})
	})

	Context("Cross-Protocol Dual-Serving Interoperability", func() {
		It("retrieves item created via gRPC when listing over REST HTTP", func() {
			itemName := "CrossProto gRPC->REST " + uuid.NewString()[:8]

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			created, err := grpcItemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
				Name:        itemName,
				Description: "Created via gRPC port",
				Price:       4200,
			})
			Expect(err).NotTo(HaveOccurred())

			// Fetch via REST HTTP
			resp, err := apiClient.FindItemsWithResponse(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			Expect(resp.JSON200).NotTo(BeNil())

			var found bool
			for _, it := range *resp.JSON200 {
				if it.Id.String() == created.Item.Id {
					Expect(it.Name).To(Equal(itemName))
					Expect(it.Price).To(Equal(int64(4200)))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Item created via gRPC should be found via HTTP REST")
		})

		It("retrieves item created via REST HTTP when listing over gRPC", func() {
			itemName := "CrossProto REST->gRPC " + uuid.NewString()[:8]
			desc := "Created via REST port"

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			postResp, err := apiClient.CreateItemWithResponse(ctx, client.CreateItemJSONRequestBody{
				Name:        itemName,
				Description: &desc,
				Price:       8800,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode()).To(Equal(http.StatusCreated))
			Expect(postResp.JSON201).NotTo(BeNil())

			createdRest := postResp.JSON201
			// Query via gRPC

			grpcListResp, err := grpcItemClient.FindItems(ctx, &itemv1.FindItemsRequest{
				Name: itemName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(grpcListResp.Items).NotTo(BeEmpty())

			var found bool
			for _, it := range grpcListResp.Items {
				if it.Id == createdRest.Id.String() {
					Expect(it.Name).To(Equal(itemName))
					Expect(it.Price).To(Equal(int64(8800)))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Item created via REST should be found via gRPC FindItems")
		})
	})
})
