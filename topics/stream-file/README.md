If we look at the below file transfer implementation over the network, we are using buffer size of 2048. If we write file size below
that we get that many bytes were written and that many received.

```go 
package main

import (
	"crypto/rand"
"fmt"
	"io"
	"log"
	"net"
	"time"
)

type FileServer struct{}

func (fs *FileServer) start() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	// Receiving connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		file := buf[:n]
		fmt.Println(file)
		fmt.Printf("received %d bytes over the network\n", n)
	}
}

func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}
	n, err := conn.Write(file)
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes over the network\n", n)
	return nil
}

func main() {
	go func() {
		time.Sleep(4 * time.Second)
		sendFile(4000)
	}()
	server := &FileServer{}
	server.start()
}
```

But if we write byte size above 2048 e.g. in this case 4000 bytes. We get 4000 bytes written, and on receive side we get 2048 bytes 
received i.e. first chunk of bytes and then 1952 bytes received.

It's very hard to handle it on server side because how do you know, how large the file is and in basically in non-streaming way you
are going to get into issues and best way to fix all of this is to stream the file over the network.

So we are going to modify sendFile ad below:

If we look at implementation of Copy, it uses CopyBuffer which keeps gonna copying specified buffer of size 32 * 1024. It's gonna
keep doing it until you stop it or until it reaches end of file.

```go 
package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type FileServer struct{}

func (fs *FileServer) start() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	// Receiving connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	for {
		n, err := io.Copy(buf, conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf.Bytes())
		fmt.Printf("received %d bytes over the network\n", n)
	}
}

func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}

	n, err := io.Copy(conn, bytes.NewReader(file))
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes over the network\n", n)
	return nil
}

func main() {
	go func() {
		time.Sleep(4 * time.Second)
		sendFile(4000)
	}()
	server := &FileServer{}
	server.start()
}
```

Now we see 4000 bytes written but after that we don't see any logs. Our system hangs. We can add a panic in our readLoop to find out:

```go 
func (fs *FileServer) readLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	for {
		n, err := io.Copy(buf, conn)
		if err != nil {
			log.Fatal(err)
		}
		panic("panic")
		fmt.Println(buf.Bytes())
		fmt.Printf("received %d bytes over the network\n", n)
	}
}
```

We now see that we never panic. It's not panicing because we are never reaching that code. Why is that?
Because it's a connection, it's not getting end of file. So it's waiting to receive EOF.

To fix that we can use ```CopyN``` method. Which copies only specified amount of byte.

```go 
func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3001")
	if err != nil {
		return err
	}

	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes over the network\n", n)
	return nil
}
```

It's not going to work if we only add it client side. We also have to use CopyN server side. But now the problem is what size we 
should provide?
So we are going to specify and write filesize over the connection.

```go 
binary.Write(conn, binary.LittleEndian, int64(size))
```

On the server side we first need to read the filesize.

```go 
func (fs *FileServer) readLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	for {
		var size int64
		binary.Read(conn, binary.LittleEndian, &size)
		n, err := io.CopyN(buf, conn, size)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf.Bytes())
		fmt.Printf("received %d bytes over the network\n", n)
	}
}
```

Now everything works fine.
