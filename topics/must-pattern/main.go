package main

import (
	"io"
	"os"
)

func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

func main() {
	src := "./template.txt"
	dst := "./out/template.txt"

	r := Must(os.Open(src))
	defer r.Close()

	w := Must(os.Create(dst))
	defer w.Close()

	Must(io.Copy(w, r))

	if err := w.Close(); err != nil {
		panic(err)
	}
}
