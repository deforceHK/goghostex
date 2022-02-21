package goghostex

import (
	"fmt"
	"testing"
)

func TestUtil(t *testing.T) {
	var result = FloatToString1(50000.88, 2, 0.10)
	fmt.Println(result)
}
