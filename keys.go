package main

import "sort"

type keys []string

//NewKeys returns an empty keys vector with capacity c.
func NewKeys(c int) keys {
	return make(keys, 0, c)
}

func (k *keys) Add(key string) {
	*k = append(*k, key)
}

func (k *keys) Sort() []string {
	ss := []string(*k)
	sort.Strings(ss)
	return ss
}
