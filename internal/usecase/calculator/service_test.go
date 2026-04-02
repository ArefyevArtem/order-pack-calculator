package calculator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"order-pack-calculator/internal/domain"
	ucmocks "order-pack-calculator/internal/usecase/calculator/mocks"
)

func TestService_Calculate_ItemsNotPositive(t *testing.T) {
	// items <= 0: PackStore.List must not be called.
	svc := NewService(ucmocks.NewMockPackStore(t))
	_, err := svc.Calculate(context.Background(), 0)
	assert.ErrorIs(t, err, ErrItemsNotPositive)
}

func TestService_Calculate_NoPackSizes(t *testing.T) {
	store := ucmocks.NewMockPackStore(t)
	store.EXPECT().List(mock.Anything).Return([]int(nil), nil)

	svc := NewService(store)
	_, err := svc.Calculate(context.Background(), 10)
	assert.ErrorIs(t, err, ErrNoPackSizes)
}

func TestService_ReplaceAndCalculate_MinPacks(t *testing.T) {
	ctx := context.Background()
	store := ucmocks.NewMockPackStore(t)
	store.EXPECT().ReplaceAll(mock.Anything, []int{20, 50, 100}).Return(nil)
	store.EXPECT().List(mock.Anything).Return([]int{20, 50, 100}, nil)

	svc := NewService(store)
	require.NoError(t, svc.ReplacePackSizes(ctx, []int{20, 50, 100}))
	res, err := svc.Calculate(ctx, 100)
	require.NoError(t, err)
	require.Len(t, res.Packs, 1)
	assert.Equal(t, 100, res.Packs[0].Pack)
	assert.Equal(t, 1, res.Packs[0].Quantity)
}

func TestService_Calculate_NoExactPacking(t *testing.T) {
	ctx := context.Background()
	store := ucmocks.NewMockPackStore(t)
	store.EXPECT().ReplaceAll(mock.Anything, []int{4, 6}).Return(nil)
	store.EXPECT().List(mock.Anything).Return([]int{4, 6}, nil)

	svc := NewService(store)
	require.NoError(t, svc.ReplacePackSizes(ctx, []int{4, 6}))
	_, err := svc.Calculate(ctx, 5)
	assert.ErrorIs(t, err, domain.ErrNoExactPacking)
}
