package service

import (
	"context"

	"vitek/internal/tokens"
)

// SimilarListing is a single search hit returned by a listing processor.
type SimilarListing struct {
	AvitoID string
	Title   string
}

// ListingProcessor finds listings similar to a source Avito URL.
type ListingProcessor interface {
	FindSimilar(ctx context.Context, listingURL string) ([]SimilarListing, error)
}

// StubListingProcessor is the Day-1 stub (no Avito network).
type StubListingProcessor struct{}

func NewStubListingProcessor() *StubListingProcessor {
	return &StubListingProcessor{}
}

func (p *StubListingProcessor) FindSimilar(_ context.Context, listingURL string) ([]SimilarListing, error) {
	id := tokens.ListingIDFromURL(listingURL)
	out := make([]SimilarListing, 0, tokens.ListingSearchStubResultCount)
	for i := 1; i <= tokens.ListingSearchStubResultCount; i++ {
		out = append(out, SimilarListing{
			AvitoID: tokens.StubSimilarAvitoID(id, i),
			Title:   tokens.StubSimilarTitle(id, i),
		})
	}
	return out, nil
}
