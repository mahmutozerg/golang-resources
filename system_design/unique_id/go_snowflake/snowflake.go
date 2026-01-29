package snowflake

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	Epoch          int64 = 1016060400000 // 13-03-2002 23:00:00 UTC
	timestampBits        = 41
	datacenterBits       = 5
	machineBits          = 5
	sequenceBits         = 12

	datacenterMax = -1 ^ (-1 << datacenterBits)
	machineMax    = -1 ^ (-1 << machineBits)
	sequenceMask  = -1 ^ (-1 << sequenceBits) // 4095

	machineShift    = sequenceBits
	datacenterShift = sequenceBits + machineBits
	timestampShift  = sequenceBits + machineBits + datacenterBits
)

type Node struct {
	mu           sync.Mutex
	timestamp    int64
	datacenterID int64
	machineID    int64
	sequence     int64
}

func NewNode(datacenterID, machineID int64) (*Node, error) {
	if datacenterID < 0 || datacenterID > datacenterMax {
		return nil, fmt.Errorf("datacenter ID must be between 0 and %d", datacenterMax)
	}
	if machineID < 0 || machineID > machineMax {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", machineMax)
	}

	return &Node{
		timestamp:    0,
		datacenterID: datacenterID,
		machineID:    machineID,
		sequence:     0,
	}, nil
}

func (n *Node) NextId() (int64, error) {

	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < n.timestamp {
		return 0, fmt.Errorf("NTP Err now is'in past of Now: %v , Min:%v ", now, n.timestamp)
	}

	if now == n.timestamp {
		n.sequence = (n.sequence + 1) & sequenceMask

		for n.sequence == 0 && now <= n.timestamp {
			runtime.Gosched()
			now = time.Now().UnixMilli()
		}
	}

	if now > n.timestamp {
		n.sequence = 0
		n.timestamp = now
	}

	return ((now-Epoch)<<timestampShift |
		(n.datacenterID<<datacenterShift |
			(n.machineID << machineShift) | n.sequence)), nil

}
