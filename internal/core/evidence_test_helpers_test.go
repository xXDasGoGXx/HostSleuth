package core

import (
	"bytes"
	"sort"
)

func evidenceFileBytes(files map[string][]byte) []byte {
	var b bytes.Buffer
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		b.WriteString(name)
		b.WriteByte('\n')
		b.Write(files[name])
		b.WriteByte('\n')
	}
	return b.Bytes()
}
