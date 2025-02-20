package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNestedOrderedList1(t *testing.T) {
	nestedOrderesList := NestedOrderedList[string, string, int]{}
	nestedOrderesList.Set("a", "a", 0)

	require.Equal(t, 1, len(nestedOrderesList.lists))
	require.Equal(t, 1, len(nestedOrderesList.lists[0].value))

	nestedOrderesList.Remove("a", "a")
	require.Equal(t, 0, len(nestedOrderesList.lists))
}

func TestNestedOrderedList2(t *testing.T) {
	nestedOrderesList := NestedOrderedList[string, string, int]{}
	nestedOrderesList.Set("a", "a", 0)
	nestedOrderesList.Set("a", "b", 0)

	require.Equal(t, 1, len(nestedOrderesList.lists))
	require.Equal(t, 2, len(nestedOrderesList.lists[0].value))

	nestedOrderesList.Remove("a", "b")
	require.Equal(t, 1, len(nestedOrderesList.lists))
	require.Equal(t, 1, len(nestedOrderesList.lists[0].value))
}

func TestNestedOrderedList3(t *testing.T) {
	nestedOrderesList := NestedOrderedList[string, string, int]{}
	nestedOrderesList.Set("a", "a", 0)
	nestedOrderesList.Set("a", "a", 1)

	require.Equal(t, 1, len(nestedOrderesList.lists))
	require.Equal(t, 1, nestedOrderesList.lists[0].value[0].value)
}
