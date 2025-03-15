package main

import (
	"fmt"
	"os"

	"io/fs"
)

func main() {
	fsys := os.DirFS(".")
	matches, err := fs.Glob(fsys, "*")
	fmt.Println("Matches:", matches)
	fmt.Println("Error:", err)
}
