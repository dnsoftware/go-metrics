package main

import "os"

func main() {
	os.Exit(1) // want "Вызов os.Exit в функции main"
}
