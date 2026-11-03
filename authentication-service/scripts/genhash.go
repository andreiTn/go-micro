//go:build ignore

package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("verysecret"), 12)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(hash))
}
