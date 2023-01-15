package main

import (
	"fmt"
	"log"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(fmt.Sprintf("[ERROR] %v", err))
	}
}

func ConventStringToBytes(str string) []byte {
	return []byte(str)
}

func ConventBytesToString(bytes []byte) string {
	return string(bytes[:])
}
