package node

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type parseLineError struct {
	arg     string
	message string
}

func (e *parseLineError) Error() string {
	return fmt.Sprintf("%s - %s", e.arg, e.message)
}

type Node struct {
	Name  string
	items map[string]string
	rwmu  *sync.RWMutex
	file  *os.File
}

func (n *Node) Put(key, val string) error {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	encLen := base64.StdEncoding.EncodedLen(len(val))
	buf := make([]byte, 0, 4+len(key)+1+encLen+1)

	// Append akıllı olduğu için kendisi veriyi yazıyor
	// ve len değerini ileriye öteliyor ve artık oraya yazabiliyor düşünelim
	// yani örnek key = foo ,se  aşağıda len=8 oluyor
	buf = append(buf, 'S', 'E', 'T', ',')
	buf = append(buf, key...)
	buf = append(buf, ',')

	// Ne yazıkki base64 append kadar akıllı değil
	// Bunun sebebi base64 yazmak için var append allocate+yazmak
	// o yüzden ona bak bu alanı kullanabilirsin diyoruz
	// [0:buffer_boyutu+base64boyutu] diyerek her yer senin diyerek
	// yazdırıyoruz
	start := len(buf)
	buf = buf[:start+encLen]
	base64.StdEncoding.Encode(buf[start:], []byte(val))

	buf = append(buf, '\n')

	c, err := n.file.Write(buf)
	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("WAL write failed: wrote 0 bytes")
	}
	if err := n.file.Sync(); err != nil {
		return err
	}

	n.items[key] = val
	return nil
}

func (n *Node) Del(key string) error {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()
	buf := make([]byte, 0, 4+len(key)+1)
	buf = append(buf, 'D', 'E', 'L', ',')
	buf = append(buf, key...)
	buf = append(buf, '\n')

	c, err := n.file.Write(buf)
	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("WAL write failed: wrote 0 bytes")
	}

	if err = n.file.Sync(); err != nil {
		return err
	}

	delete(n.items, key)

	return nil
}

func (n *Node) Get(key string) (string, bool) {

	n.rwmu.RLock()
	defer n.rwmu.RUnlock()

	v, exist := n.items[key]
	return v, exist
}

func New(name string) (*Node, error) {
	n := &Node{Name: name, items: make(map[string]string), rwmu: &sync.RWMutex{}}

	err := os.MkdirAll("./wal", 0755)
	path := filepath.Join("./wal", name+".aof")

	if err != nil {
		return nil, err
	}

	readFile, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0755)

	if err != nil {
		return nil, err
	}

	err = setMap(n, readFile)
	if err != nil {
		return nil, err
	}
	readFile.Close()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)

	if err != nil {
		return nil, err
	}

	n.file = f
	return n, nil
}

func setMap(n *Node, f *os.File) error {

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		vals := strings.Split(line, ",")

		if len(vals) == 0 {
			return &parseLineError{arg: "strings.Split(line,\",\")", message: "Failed to extract line"}
		}

		if strings.ToUpper(vals[0]) == "SET" && len(vals) == 3 {
			dval, err := base64.StdEncoding.DecodeString(vals[2])

			if err != nil {
				return err
			}
			n.items[vals[1]] = string(dval)

		} else if strings.ToUpper(vals[0]) == "DEL" && len(vals) == 2 {
			delete(n.items, vals[1])
		} else {
			return &parseLineError{arg: "SET or DEL Insufficient val", message: "Failed to parse line"}

		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
