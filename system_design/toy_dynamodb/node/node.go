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

	v64 := base64.StdEncoding.EncodeToString([]byte(val))

	c, err := n.file.WriteString(fmt.Sprintf("SET,%s,%s\n", key, v64))

	if err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("WAL write failed: wrote 0 bytes")
	}

	if err = n.file.Sync(); err != nil {
		return err
	}
	n.items[key] = val

	return nil
}

func (n *Node) Del(key string) error {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()

	c, err := n.file.WriteString(fmt.Sprintf("DEL,%s\n", key))
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
