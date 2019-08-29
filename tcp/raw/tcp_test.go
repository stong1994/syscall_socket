package main

import (
	"fmt"
	"testing"
)

func Test_htons(t *testing.T) {
	fmt.Println(htons(1))
	fmt.Println(htons(2))
	fmt.Println(htons(3))
}
