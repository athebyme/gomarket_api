package clients

import (
	"fmt"
	pkg2 "gomarketplace_api/internal/suppliers/wholesaler/pkg"
	clients2 "gomarketplace_api/internal/suppliers/wholesaler/pkg/clients"
	"gomarketplace_api/pkg/logger"
	"io"
)

type WServiceClient struct {
	FetcherChain *pkg2.FetcherChain
}

func NewWServiceClient(host string, writer io.Writer) (*WServiceClient, error) {
	log := logger.NewLogger(writer, "[WServiceClient]")

	fetcherChain := pkg2.NewFetcherChain(log)

	if err := registerClients(fetcherChain, host, writer); err != nil {
		return nil, fmt.Errorf("failed to register clients: %w", err)
	}

	return &WServiceClient{
		FetcherChain: fetcherChain,
	}, nil
}

func registerClients(fetcherChain *pkg2.FetcherChain, host string, writer io.Writer) error {
	clientsToRegister := []struct {
		name    string
		fetcher pkg2.Fetcher
	}{
		{"appellations", clients2.NewAppellationsFetcher(host, writer)},
		{"descriptions", clients2.NewDescriptionsClient(host, writer)},
		{"globalIDs", clients2.NewGlobalIDsClient(host, writer)},
		{"images", clients2.NewImageClient(host, writer)},
		{"prices", clients2.NewPriceClient(host, writer)},
		{"sizes", clients2.NewSizesClient(host, writer)},
		{"brands", clients2.NewBrandsClient(host, writer)},
		{"barcodes", clients2.NewBarcodesClient(host, writer)},
		{"media", clients2.NewImageClient(host, writer)},
	}

	for _, client := range clientsToRegister {
		if err := fetcherChain.Register(client.name, client.fetcher); err != nil {
			return fmt.Errorf("failed to register client '%s': %w", client.name, err)
		}
	}

	return nil
}
