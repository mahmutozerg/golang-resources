package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strconv"
)

func main() {
	const (
		input1 = "The tunneling gopher digs downwards, "
		input2 = "unaware of what he will find."
	)

	servers := make(map[int][]string)
	var serverCount uint64 = 3

	shaSource := sha1.New()

	datas := func() []string {

		out := make([]string, 0, 10)
		for i := 0; i < 10; i++ {
			out = append(out, "user_"+strconv.FormatInt(int64(i), 10))
		}
		return out
	}()

	fmt.Printf("*******SERVER COUNT = %v**********\n", serverCount)

	for _, v := range datas {
		shaSource.Reset()
		shaSource.Write([]byte(v))
		sum := shaSource.Sum(nil)
		servers[int(binary.BigEndian.Uint64(sum)%serverCount)] = append(servers[int(binary.BigEndian.Uint64(sum)%serverCount)], v)
		fmt.Printf("%v added into %v\n", v, binary.BigEndian.Uint64(sum)%serverCount)
	}

	for k := range servers {
		delete(servers, k)
	}

	servers = make(map[int][]string)
	serverCount++
	fmt.Printf("*******SERVER COUNT = %v**********\n", serverCount)

	for _, v := range datas {
		shaSource.Reset()
		shaSource.Write([]byte(v))
		sum := shaSource.Sum(nil)
		servers[int(binary.BigEndian.Uint64(sum)%serverCount)] = append(servers[int(binary.BigEndian.Uint64(sum)%serverCount)], v)
		fmt.Printf("%v added into %v\n", v, binary.BigEndian.Uint64(sum)%serverCount)
	}

}
