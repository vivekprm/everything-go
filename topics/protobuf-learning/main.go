package main

import (
	"io/ioutil"
	"log"
	"protobuf-learning/tutorial1"

	"google.golang.org/protobuf/proto"
)

func main() {
	// writeToFile("out1.txt")
	// writeToFile("out2.txt")
	readFromFile("out.txt")
	readFromFile("out2.txt")
}

func writeToFile(filename string) {
	p1 := &tutorial1.Person{
		Name:        "John Doe",
		Id:          12345,
		ExampleEnum: tutorial1.ExampleEnum1_OPTION_FOUR,
	}

	// Write the new address book back to disk.
	out, err := proto.Marshal(p1)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
	}
	if err := ioutil.WriteFile(filename, out, 0644); err != nil {
		log.Fatalln("Failed to write address book:", err)
	}
}

func readFromFile(filename string) {
	in, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}

	p1 := &tutorial1.Person{}
	if err := proto.Unmarshal(in, p1); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}
	log.Println("Person read from file:", p1)
}
