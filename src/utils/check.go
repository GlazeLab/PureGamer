package utils

import (
	"github.com/GlazeLab/PureGamer/src/model"
	"sort"
)

func in(target string, arr []string) bool {
	// must be sorted sort.Strings(arr)
	index := sort.SearchStrings(arr, target)
	if index < len(arr) && arr[index] == target {
		return true
	}
	return false
}

func IsNotAllowed(peerId string, wblist model.WhiteOrBlackList) bool {
	switch wblist.Type {
	case "all":
		return false
	case "black":
		return in(peerId, wblist.BlackList)
	case "white":
		return !in(peerId, wblist.WhiteList)
	}
	return true
}
