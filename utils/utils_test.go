package utils

import (
	"fmt"
	"testing"
)

func TestGetLatestHeight(t *testing.T) {
	height := GetLatestHeight()
	fmt.Println(height)
}
