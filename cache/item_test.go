package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"cosmossdk.io/collections"
)

func createItemCache() (context.Context, *ItemCache[uint64]) {
	ctx, store := Deps()
	sb := collections.NewSchemaBuilder(store)

	itemCache := NewItemCache[uint64](
		sb,
		collections.NewPrefix(0),
		"testitem",
		collections.Uint64Value,
		&Caches{},
	)

	return ctx, itemCache
}

func TestItem1(t *testing.T) {
	ctx, itemCache := createItemCache()
	AddCache(itemCache)
	defer RemoveCache(itemCache)

	_, has := itemCache.Get(ctx)
	require.False(t, has)

	require.NoError(t, Transact(ctx, func(innerCtx context.Context) error {
		itemCache.Set(innerCtx, 1)
		return nil
	}))

	value, has := itemCache.Get(ctx)
	require.True(t, has)
	require.Equal(t, uint64(1), value)
}

func TestItem2(t *testing.T) {
	ctx, itemCache := createItemCache()
	AddCache(itemCache)
	defer RemoveCache(itemCache)

	_, has := itemCache.Get(ctx)
	require.False(t, has)

	require.NoError(t, Transact(ctx, func(innerCtx context.Context) error {
		itemCache.Set(innerCtx, 1)
		return nil
	}))

	value, has := itemCache.Get(ctx)
	require.True(t, has)
	require.Equal(t, uint64(1), value)

	require.Error(t, Transact(ctx, func(innerCtx context.Context) error {
		itemCache.Set(innerCtx, 2)
		return fmt.Errorf("test error")
	}))

	value, has = itemCache.Get(ctx)
	require.True(t, has)
	require.Equal(t, uint64(1), value)
}
