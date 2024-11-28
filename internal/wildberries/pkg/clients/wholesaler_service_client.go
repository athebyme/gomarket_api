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
	*clients.BrandsClient
	*clients.BarcodesClient
}

func NewWServiceClient(host string, writer io.Writer) *WServiceClient {
	return &WServiceClient{
		AppellationsClient: clients.NewAppellationsClient(host, writer),
		DescriptionsClient: clients.NewDescriptionsClient(host, writer),
		GlobalIDsClient:    clients.NewGlobalIDsClient(host, writer),
		ImageClient:        clients.NewImageClient(host, writer),
		PriceClient:        clients.NewPriceClient(host, writer),
		SizesClient:        clients.NewSizesClient(host, writer),
		BrandsClient:       clients.NewBrandsClient(host, writer),
		BarcodesClient:     clients.NewBarcodesClient(host, writer),
	}
}
