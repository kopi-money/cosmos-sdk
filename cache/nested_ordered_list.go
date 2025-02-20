package cache

import "golang.org/x/exp/slices"

type NestedOrderedList[K1, K2 ordered, V any] struct {
	lists []KeyValue[K1, []KeyValue[K2, V]]
}

func (nol *NestedOrderedList[K1, K2, V]) Size() (sum int) {
	for _, list := range nol.lists {
		sum += len(list.value)
	}

	return sum
}

func (nol *NestedOrderedList[K1, K2, V]) Has(key1 K1, key2 K2) bool {
	_, _, has := nol.GetIndex(key1, key2)
	return has
}

func (nol *NestedOrderedList[K1, K2, V]) GetInnerOrderedList(key1 K1) *OrderedList[K2, V] {
	index1, has := getIndex(nol.lists, key1)
	if !has {
		return &OrderedList[K2, V]{}
	}

	return &OrderedList[K2, V]{
		list: nol.lists[index1].value,
	}
}

func (nol *NestedOrderedList[K1, K2, V]) Get(key1 K1, key2 K2) (V, bool) {
	index1, index2, has := nol.GetIndex(key1, key2)
	if !has {
		var v V
		return v, false
	}

	return nol.lists[index1].value[index2].value, true
}

func (nol *NestedOrderedList[K1, K2, V]) GetIndex(key1 K1, key2 K2) (int, int, bool) {
	index1, has := getIndex(nol.lists, key1)
	if !has {
		return index1, 0, false
	}

	index2, has := getIndex(nol.lists[index1].value, key2)
	return index1, index2, has
}

func (nol *NestedOrderedList[K1, K2, V]) Remove(key1 K1, key2 K2) {
	index1, index2, has := nol.GetIndex(key1, key2)
	if !has {
		return
	}

	innerList := nol.lists[index1].value
	innerList = append(innerList[:index2], innerList[index2+1:]...)
	innerList = slices.Clip(innerList)

	if len(innerList) > 0 {
		nol.lists[index1].value = innerList
	} else {
		nol.lists = append(nol.lists[:index1], nol.lists[index1+1:]...)
		nol.lists = slices.Clip(nol.lists)
	}
}

func (nol *NestedOrderedList[K1, K2, V]) Set(key1 K1, key2 K2, value V) {
	index1, has := getIndex(nol.lists, key1)

	var innerList []KeyValue[K2, V]
	if has {
		innerList = nol.lists[index1].value
	} else {
		nol.addList(index1, key1)
	}

	listIndex, has := getIndex(innerList, key2)
	entry := KeyValue[K2, V]{key2, value}

	if has {
		innerList[listIndex] = entry
	} else {
		if listIndex == len(innerList) {
			innerList = append(innerList, entry)
		} else {
			innerList = append(innerList[:listIndex+1], innerList[listIndex:]...)
			innerList[listIndex] = entry
		}
	}

	nol.lists[index1].value = innerList
}

func (nol *NestedOrderedList[K1, K2, V]) addList(index1 int, key K1) {
	newList := KeyValue[K1, []KeyValue[K2, V]]{
		key: key,
	}

	if index1 == len(nol.lists) {
		nol.lists = append(nol.lists, newList)
	} else {
		nol.lists = append(nol.lists[:index1+1], nol.lists[index1:]...)
		nol.lists[index1] = newList
	}
}
