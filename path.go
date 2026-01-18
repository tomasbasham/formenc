package formenc

import (
	"fmt"
	"strings"
)

type pathSegment struct {
	Key   string
	Index bool // true for []
}

func parseKey(key string) ([]pathSegment, error) {
	var path []pathSegment
	for len(key) > 0 {
		i := strings.IndexByte(key, '[')
		if i == -1 {
			path = append(path, pathSegment{Key: key})
			break
		}

		if i > 0 {
			path = append(path, pathSegment{Key: key[:i]})
		}

		key = key[i+1:]
		j := strings.IndexByte(key, ']')
		if j == -1 {
			return nil, fmt.Errorf("form: invalid key syntax")
		}

		part := key[:j]
		if part == "" {
			path = append(path, pathSegment{Index: true})
		} else {
			path = append(path, pathSegment{Key: part})
		}
		key = key[j+1:]
	}
	return path, nil
}
