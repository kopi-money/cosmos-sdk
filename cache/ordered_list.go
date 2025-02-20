package cache

import "golang.org/x/exp/slices"

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type KeyValue[K ordered, V any] struct {
	key   K
	value V
}

func (kv KeyValue[K, V]) Key() K {
	return kv.key
}

func (kv KeyValue[K, V]) Value() V {
	return kv.value
}

type OrderedList[K ordered, V any] struct {
	list []KeyValue[K, V]
}

func (ol *OrderedList[K, V]) Size() int {
	return len(ol.list)
}

func (ol *OrderedList[K, V]) Clear() {
	ol.list = nil
}

func (ol *OrderedList[K, V]) GetByIndex(index int) KeyValue[K, V] {
	return ol.list[index]
}

func (ol *OrderedList[K, V]) Has(key K) bool {
	_, has := ol.GetIndex(key)
	return has
}

func (ol *OrderedList[K, V]) Get(key K) (V, bool) {
	listIndex, has := ol.GetIndex(key)
	if !has {
		var v V
		return v, false
	}

	return ol.list[listIndex].value, true
}

func (ol *OrderedList[K, V]) GetAll() (filtered []KeyValue[K, V]) {
	return ol.list
}

func (ol *OrderedList[K, V]) GetKeys() (keys []K) {
	for _, entry := range ol.list {
		keys = append(keys, entry.key)
	}

	return
}

func (ol *OrderedList[K, V]) GetIndex(key K) (int, bool) {
	return getIndex(ol.list, key)
}

func (ol *OrderedList[K, V]) Remove(key K) {
	listIndex, has := ol.GetIndex(key)
	if has {
		ol.list = append(ol.list[:listIndex], ol.list[listIndex+1:]...)
	}

	ol.list = slices.Clip(ol.list)
}

func (ol *OrderedList[K, V]) Set(keyValue KeyValue[K, V]) {
	listIndex, has := ol.GetIndex(keyValue.key)
	entry := KeyValue[K, V]{keyValue.key, keyValue.value}

	if has {
		ol.list[listIndex] = entry
	} else {
		if listIndex == len(ol.list) {
			ol.list = append(ol.list, entry)
		} else {
			ol.list = append(ol.list[:listIndex+1], ol.list[listIndex:]...)
			ol.list[listIndex] = entry
		}
	}
}

func (ol *OrderedList[K, V]) set(keyValues []KeyValue[K, V]) {
	ol.list = nil
	for _, keyValue := range keyValues {
		ol.Set(keyValue)
	}
}

func getIndex[K ordered, V any](list []KeyValue[K, V], key K) (int, bool) {
	low, high := 0, len(list)-1

	for low <= high {
		if high-low <= 8 {
			for index := low; index <= high; index++ {
				if list[index].key == key {
					return index, true
				}

				if list[index].key > key {
					return index, false
				}
			}

			return high + 1, false
		} else {
			mid := low + (high-low)/2

			if list[mid].key == key {
				return mid, true
			} else if list[mid].key < key {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}

	return low, false
}
