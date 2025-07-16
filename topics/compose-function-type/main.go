package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type TrasformFunc func(string) string

type Server struct {
	filenameTransformFunc TrasformFunc
}

func (s *Server) handleRequest(filename string) error {
	newFileName := s.filenameTransformFunc(filename)
	fmt.Println("New filename is ", newFileName)
	return nil
}

func hashFilename(filename string) string {
	hash := sha256.Sum256([]byte(filename))
	newFileName := hex.EncodeToString(hash[:])
	return newFileName
}

func prefixFilename(prefix string) TrasformFunc {
	return func(filename string) string {
		return prefix + filename
	}
}

func main() {
	s := &Server{
		filenameTransformFunc: hashFilename,
	}
	s.handleRequest("cool_pic.jpg")

	s = &Server{
		filenameTransformFunc: prefixFilename("BOB_"),
	}
	s.handleRequest("cool_pic.jpg")
}
