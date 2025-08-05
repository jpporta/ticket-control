package utils

import (
	"fmt"
	"math/rand/v2"
)

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		r := rand.Int()
		fmt.Println(r)
		b[i] = letters[r%len(letters)]
	}
	return string(b)
}
