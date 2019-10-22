package content

import (
	"fmt"
	"testing"
)

func TestDependencies(t *testing.T) {
	deps := NewCacheDependencies()
	deps.Set("abc", "123", "de", "live")
	d := deps.Get("123", "de", "live")

	fmt.Println(d)

	if len(d) != 1 {
		t.Fatal("unexpected length")
	}

	if d[0] != "abc" {
		t.Fatal("unexpected dependency")
	}
}
