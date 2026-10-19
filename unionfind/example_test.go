package unionfind_test

import (
	"fmt"

	"github.com/hmdsefi/gograph/unionfind"
)

func Example() {
	// Two addresses that paid into the same transaction belong to one owner.
	s := unionfind.New[string]()
	s.Union("addr1", "addr2")
	s.Union("addr3", "addr4")
	s.Union("addr2", "addr5")

	fmt.Println(s.Connected("addr1", "addr5"))
	fmt.Println(s.Connected("addr1", "addr3"))
	fmt.Println(s.Groups())

	// Output:
	// true
	// false
	// [[addr1 addr2 addr5] [addr3 addr4]]
}
