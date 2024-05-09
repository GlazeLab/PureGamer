package relaying

import "github.com/libp2p/go-libp2p/core/protocol"

func match(p protocol.ID) bool {
	if pattern.MatchString(string(p)) {
		// check version
		version := pattern.FindStringSubmatch(string(p))[1]
		if version != "0.0.1" {
			return false
		}
		return true
	}
	return false
}
