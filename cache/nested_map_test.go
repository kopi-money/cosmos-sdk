package cache

import (
	"context"
	"testing"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"
)

func createNestedMapCache() (context.Context, *NestedMapCache[string, string, uint64]) {
	store, ctx := testutil.CacheDeps()
	sb := collections.NewSchemaBuilder(store)

	mapCache := NewNestedMapCache[string, string, uint64](
		sb,
		collections.NewPrefix(0),
		"testmap",
		collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		collections.Uint64Value,
		&Caches{},
	)

	return ctx, mapCache
}

func TestNestedMap0(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(ctx, "a", "a", 0)

	iterator := nestedMapCache.Iterator(ctx, nil, "a")
	require.Equal(t, 1, len(iterator.GetAll()))
}

func TestNestedMap1(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(ctx, "a", "a", 0)
	nestedMapCache.Remove(ctx, "a", "a")

	iterator := nestedMapCache.Iterator(ctx, nil, "a")
	require.Equal(t, 0, len(iterator.GetAll()))
}

func TestNestedMap2(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(ctx, "a", "a", 0)
	nestedMapCache.Set(ctx, "a", "b", 0)
	nestedMapCache.Remove(ctx, "a", "a")

	iterator := nestedMapCache.Iterator(ctx, nil, "a")
	require.Equal(t, 1, len(iterator.GetAll()))
}

func TestNestedMap3(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(ctx, "a", "a", 0)
	nestedMapCache.Set(ctx, "a", "a", 1)

	iterator := nestedMapCache.Iterator(ctx, nil, "a")

	values := iterator.GetAll()
	require.Equal(t, 1, len(values))
	require.Equal(t, uint64(1), values[0])
}

func TestNestedMap4(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(ctx, "a", "a", 0)
	nestedMapCache.Remove(ctx, "a", "a")
	nestedMapCache.Set(ctx, "a", "a", 1)

	iterator := nestedMapCache.Iterator(ctx, nil, "a")

	values := iterator.GetAll()
	require.Equal(t, 1, len(values))
	require.Equal(t, uint64(1), values[0])
}

func TestNestedMap5(t *testing.T) {
	ctx, nestedMapCache := createNestedMapCache()
	tx1 := NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	nestedMapCache.Set(tx1, "a", "a", 0)
	TransactionHandler.CommitToCache(tx1)
	TransactionHandler.Clear(tx1)

	tx2 := NewCacheContext(ctx, 1, true)
	nestedMapCache.Remove(tx2, "a", "a")

	iterator := nestedMapCache.Iterator(tx2, nil, "a")
	values := iterator.GetAll()
	require.Equal(t, 0, len(values))
}
