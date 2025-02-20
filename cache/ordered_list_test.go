package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedList1(t *testing.T) {
	ol := OrderedList[string, int]{}
	ol.Set(KeyValue[string, int]{"a", 0})
	ol.Set(KeyValue[string, int]{"c", 0})

	require.Equal(t, 2, len(ol.list))
	require.Equal(t, "a", ol.list[0].key)
	require.Equal(t, "c", ol.list[1].key)
}

func TestOrderedList2(t *testing.T) {
	ol := OrderedList[string, int]{}
	ol.Set(KeyValue[string, int]{"c", 0})
	ol.Set(KeyValue[string, int]{"a", 0})

	require.Equal(t, 2, len(ol.list))
	require.Equal(t, "a", ol.list[0].key)
	require.Equal(t, "c", ol.list[1].key)
}

func TestOrderedList3(t *testing.T) {
	ol := OrderedList[string, int]{}
	ol.Set(KeyValue[string, int]{"c", 0})
	ol.Set(KeyValue[string, int]{"a", 0})
	ol.Set(KeyValue[string, int]{"b", 0})

	require.Equal(t, 3, len(ol.list))
	require.Equal(t, "a", ol.list[0].key)
	require.Equal(t, "b", ol.list[1].key)
	require.Equal(t, "c", ol.list[2].key)
}

func TestOrderedList4(t *testing.T) {
	ol := OrderedList[string, int]{}
	ol.Set(KeyValue[string, int]{"c", 0})
	ol.Set(KeyValue[string, int]{"b", 0})
	ol.Set(KeyValue[string, int]{"a", 0})

	require.Equal(t, 3, len(ol.list))
	require.Equal(t, "a", ol.list[0].key)
	require.Equal(t, "b", ol.list[1].key)
	require.Equal(t, "c", ol.list[2].key)
}

func TestOrderedList5(t *testing.T) {
	ol := OrderedList[string, int]{}
	ol.Set(KeyValue[string, int]{"c", 0})
	ol.Set(KeyValue[string, int]{"b", 0})
	ol.Set(KeyValue[string, int]{"a", 0})
	ol.Set(KeyValue[string, int]{"d", 0})

	require.Equal(t, 4, len(ol.list))
	require.Equal(t, "a", ol.list[0].key)
	require.Equal(t, "b", ol.list[1].key)
	require.Equal(t, "c", ol.list[2].key)
	require.Equal(t, "d", ol.list[3].key)
}
