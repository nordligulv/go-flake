package flake

import (
	"sort"
	"testing"
)

func TestNewFlake(t *testing.T) {
	f := New(1)

	var ids []string

	for i := 0; i < 4; i++ {
		id := f.NextId()

		ids = append(ids, id.String())
	}

	if !sort.StringsAreSorted(ids) {
		t.Errorf("IDs are not sorted!")
	}
}

func BenchmarkNextId(b *testing.B) {
	f := New(1)

	for i := 0; i < b.N; i++ {
		_ = f.NextId()
	}
}
