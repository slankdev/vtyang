package util

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func StringInArray(s string, slice []string) bool {
	for _, tmp := range slice {
		if tmp == s {
			return true
		}
	}
	return false
}

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

func DeepEqualJSON(j1, j2 string) (bool, error) {
	var d1 interface{}
	if err := json.Unmarshal([]byte(j1), &d1); err != nil {
		return false, errors.Wrap(err, "json.Unmarshal(j1)")
	}
	var d2 interface{}
	if err := json.Unmarshal([]byte(j2), &d2); err != nil {
		return false, errors.Wrap(err, "json.Unmarshal(j2)")
	}
	return reflect.DeepEqual(d1, d2), nil
}
