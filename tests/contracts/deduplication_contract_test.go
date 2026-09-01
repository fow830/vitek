package contracts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"vitek/internal/domain"
	"vitek/internal/service"
)

// CONTRACT-DEDUP-001: second insert with same avito_id is rejected.
func TestContract_ItemDeduplicationByAvitoID(t *testing.T) {
	ctx := context.Background()
	_, q := queries(t)
	items := service.NewItems(q)

	_, err := items.Record(ctx, "avito-123", "iPhone 15")
	require.NoError(t, err)

	_, err = items.Record(ctx, "avito-123", "iPhone 15 again")
	require.ErrorIs(t, err, domain.ErrDuplicateAvitoID)
}
