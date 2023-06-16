package goghostex

import (
	"fmt"
	"testing"
)

func TestUtil(t *testing.T) {
	var result = FloatToPrice(487.7777, 2, 0.05)
	fmt.Println(result)
}
