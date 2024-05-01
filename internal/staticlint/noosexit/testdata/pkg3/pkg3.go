package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Test os.Exit")

	os.Exit(1) // want "Вызов os.Exit в функции main"
}
