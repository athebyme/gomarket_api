package clients

import (
	"fmt"
	"gomarketplace_api/internal/wholesaler/pkg"
	"gomarketplace_api/internal/wholesaler/pkg/clients"
	"gomarketplace_api/pkg/logger"
	"io"
)

type WServiceClient struct {
	FetcherChain *pkg.FetcherChain
}

func NewWServiceClient(host string, writer io.Writer) (*WServiceClient, error) {
	log := logger.NewLogger(writer, "[WServiceClient]")

	fetcherChain := pkg.NewFetcherChain(log)

	if err := registerClients(fetcherChain, host, writer); err != nil {
		return nil, fmt.Errorf("failed to register clients: %w", err)
	}

	return &WServiceClient{
		FetcherChain: fetcherChain,
	}, nil
}

func registerClients(fetcherChain *pkg.FetcherChain, host string, writer io.Writer) error {
	clientsToRegister := []struct {
		name    string
		fetcher pkg.Fetcher
	}{
		{"appellations", clients.NewAppellationsFetcher(host, writer)},
		{"descriptions", clients.NewDescriptionsClient(host, writer)},
		{"globalIDs", clients.NewGlobalIDsClient(host, writer)},
		{"images", clients.NewImageClient(host, writer)},
		{"prices", clients.NewPriceClient(host, writer)},
		{"sizes", clients.NewSizesClient(host, writer)},
		{"brands", clients.NewBrandsClient(host, writer)},
		{"barcodes", clients.NewBarcodesClient(host, writer)},
		{"media", clients.NewImageClient(host, writer)},
	}

	for _, client := range clientsToRegister {
		if err := fetcherChain.Register(client.name, client.fetcher); err != nil {
			return fmt.Errorf("failed to register client '%s': %w", client.name, err)
		}
	}

	return nil
}
