package cache_test

import (
	"context"
	"github.com/cosmos/cosmos-sdk/cache"
	"testing"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"
)

func createMapCache() (context.Context, *cache.MapCache[string, uint64]) {
	store, ctx := testutil.Deps()
	sb := collections.NewSchemaBuilder(store)

	mapCache := NewMapCache[string, uint64](
		sb,
		collections.NewPrefix(0),
		"testmap",
		collections.StringKey,
		collections.Uint64Value,
		&Caches{},
	)

	return ctx, mapCache
}

func TestMap1(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)

	counter := 0
	iterator := mapCache.Iterator(ctx, nil)
	for iterator.Valid() {
		counter++
		item := iterator.GetNextKeyValue()
		mapCache.Remove(ctx, item.key)
	}

	require.Equal(t, 1, counter)
}

func TestMap2(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)
	mapCache.Set(ctx, "b", 1)

	counter := 0
	iterator := mapCache.Iterator(ctx, nil)
	for iterator.Valid() {
		counter++
		item := iterator.GetNextKeyValue()
		mapCache.Remove(ctx, item.key)
	}

	require.Equal(t, 2, counter)
}

func TestMap3(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)
	mapCache.Set(ctx, "b", 1)
	mapCache.Set(ctx, "c", 1)

	counter := 0
	iterator := mapCache.Iterator(ctx, nil)
	for iterator.Valid() {
		counter++
		item := iterator.GetNextKeyValue()
		mapCache.Remove(ctx, item.key)
	}

	require.Equal(t, 3, counter)
}

func TestMap4(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)
	mapCache.Remove(ctx, "a")

	counter := 0
	iterator := mapCache.Iterator(ctx, nil)
	for iterator.Valid() {
		counter++
		item := iterator.GetNextKeyValue()
		mapCache.Remove(ctx, item.key)
	}

	require.Equal(t, 0, counter)
}

func TestMap5(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)
	mapCache.Set(ctx, "b", 1)
	mapCache.Remove(ctx, "a")

	counter := 0
	iterator := mapCache.Iterator(ctx, nil)
	for iterator.Valid() {
		counter++
		item := iterator.GetNextKeyValue()
		mapCache.Remove(ctx, item.key)
	}

	require.Equal(t, 1, counter)
}

func TestMap6(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	mapCache.Set(ctx, "a", 1)
	iterator := mapCache.Iterator(ctx, nil)
	require.Equal(t, 1, len(iterator.GetAll()))

	mapCache.Set(ctx, "a", 1)
	iterator = mapCache.Iterator(ctx, nil)
	require.Equal(t, 1, len(iterator.GetAll()))

	mapCache.Set(ctx, "b", 2)
	iterator = mapCache.Iterator(ctx, nil)
	require.Equal(t, 2, len(iterator.GetAll()))

	mapCache.Remove(ctx, "a")
	iterator = mapCache.Iterator(ctx, nil)
	require.Equal(t, 1, len(iterator.GetAll()))

	mapCache.Remove(ctx, "b")
	iterator = mapCache.Iterator(ctx, nil)
	require.Equal(t, 0, len(iterator.GetAll()))
}

func TestMap7(t *testing.T) {
	ctx, mapCache := createMapCache()
	ctx = NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()

	_, has := mapCache.Get(ctx, "a")
	require.False(t, has)

	iterator := mapCache.Iterator(ctx, nil)
	require.Equal(t, 0, len(iterator.GetAll()))
}

func TestMap8(t *testing.T) {
	ctx, mapCache := createMapCache()
	tx := NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()
	AddCache(mapCache)

	mapCache.initialized = true
	mapCache.Set(tx, "a", 1)

	require.NoError(t, TransactionHandler.CommitToDB(tx))

	_, has := mapCache.Get(ctx, "a")
	require.False(t, has)

	TransactionHandler.Rollback(tx)
	_, has = mapCache.Get(ctx, "a")
	require.False(t, has)

	iterator := mapCache.Iterator(ctx, nil)
	require.Equal(t, 0, len(iterator.GetAll()))
}

func TestMap9(t *testing.T) {
	ctx, mapCache := createMapCache()
	tx1 := NewCacheContext(ctx, 1, true)
	defer TransactionHandler.ClearTransactions()
	AddCache(mapCache)

	mapCache.initialized = true
	mapCache.Set(tx1, "a", 1)
	require.NoError(t, TransactionHandler.CommitToDB(tx1))
	TransactionHandler.CommitToCache(tx1)
	TransactionHandler.Clear(tx1)

	tx2 := NewCacheContext(ctx, 1, true)
	mapCache.Set(tx1, "a", 2)
	require.NoError(t, TransactionHandler.CommitToDB(tx2))
	TransactionHandler.Rollback(tx2)
	TransactionHandler.Clear(tx2)

	v, _ := mapCache.Get(ctx, "a")
	require.Equal(t, uint64(1), v)
}
