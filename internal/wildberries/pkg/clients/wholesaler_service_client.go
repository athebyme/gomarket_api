package clients

import "gomarketplace_api/internal/wholesaler/pkg/clients"

type WServiceClient struct {
	*clients.AppellationsClient
	*clients.DescriptionsClient
	*clients.GlobalIDsClient
}

func NewWServiceClient(host string) *WServiceClient {
	return &WServiceClient{
		AppellationsClient: clients.NewAppellationsClient(host),
		DescriptionsClient: clients.NewDescriptionsClient(host),
		GlobalIDsClient:    clients.NewGlobalIDsClient(host),
	}
}
