package utils

import (
	"io"
	"net/http"
	"strconv"
)

func GetLatestHeight() int {
	url := "https://mempool.fractalbitcoin.io/api/blocks/tip/height"
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	height, err := strconv.Atoi(string(body))
	if err != nil {
		panic(err)
	}

	return height
}
