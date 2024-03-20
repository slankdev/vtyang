package util

import "sort"

func GetSortedKeys(m map[string]interface{}) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
