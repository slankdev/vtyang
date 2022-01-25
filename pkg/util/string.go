package util

import "strings"

func SplitMultiSep(s string, sep []string) []string {
	var ret []string
	ret = strings.Split(s, sep[0])
	if len(sep) > 1 {
		ret2 := []string{}
		for _, r := range ret {
			ret2 = append(ret2, SplitMultiSep(r, sep[1:])...)
		}
		ret = ret2
	}

	ret2 := []string{}
	for _, r := range ret {
		if r != "" {
			ret2 = append(ret2, r)
		}
	}
	return ret2
}
