package clients

import (
	"gomarketplace_api/internal/wholesaler/pkg/clients"
	"io"
)

// контексты для запросов добавить

type WServiceClient struct {
	*clients.AppellationsClient
	*clients.DescriptionsClient
	*clients.GlobalIDsClient
	*clients.ImageClient
	*clients.PriceClient
	*clients.SizesClient
}

func NewWServiceClient(host string, writer io.Writer) *WServiceClient {
	return &WServiceClient{
		AppellationsClient: clients.NewAppellationsClient(host),
		DescriptionsClient: clients.NewDescriptionsClient(host),
		GlobalIDsClient:    clients.NewGlobalIDsClient(host),
		ImageClient:        clients.NewImageClient(host),
		PriceClient:        clients.NewPriceClient(host),
		SizesClient:        clients.NewSizesClient(host, writer),
	}
}
