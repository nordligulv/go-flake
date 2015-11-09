// Flake generates unique identifiers that are roughly sortable by time. Flake can
// run on a cluster of machines and still generate unique IDs without requiring
// worker coordination.
//
// A Flake ID is a 64-bit integer will the following components:
//  - 41 bits is the timestamp with millisecond precision
//  - 10 bits is the host id (uses IP modulo 2^10)
//  - 13 bits is an auto-incrementing sequence for ID requests within the same millisecond
//
// Note: In order to make a millisecond timestamp fit within 41 bits, a custom
// epoch of Jan 1, 2015 00:00:00 is used.

package flake

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	HostBits     = 10
	SequenceBits = 13
)

var (
	// Custom Epoch so the timestamp can fit into 41 bits.
	// Jan 1, 2015 00:00:00 UTC
	Epoch       time.Time = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	MaxWorkerId uint64    = (1 << HostBits) - 1
	MaxSequence uint64    = (1 << SequenceBits) - 1
)

// Id represents a unique k-ordered Id
type Id uint64

// String formats the Id as a base36 string
func (id Id) String() string {
	return strconv.FormatUint(uint64(id), 36)
}

// Uint64 formats the Id as an unsigned integer
func (id Id) Uint64() uint64 {
	return uint64(id)
}

// Flake is a unique Id generator
type Flake struct {
	prevTime uint64
	workerId uint64
	sequence uint64
	mu       sync.Mutex
}

// New returns new Id generator
func New(workerId uint64) *Flake {
	return &Flake{
		sequence: 0,
		prevTime: getTimestamp(),
		workerId: workerId % MaxWorkerId,
	}
}

// WithHostId creates new Id generator with host machine address as worker id
func WithHostId() (*Flake, error) {
	workerID, err := getHostId()
	if err != nil {
		return nil, err
	}
	return New(workerID), nil
}

// WithRandomId creates new Id generator with random worker id
func WithRandomId() (*Flake, error) {
	workerID, err := getRandomId()
	if err != nil {
		return nil, err
	}
	return New(workerID), nil
}

// NextId returns a new Id from the generator
func (f *Flake) NextId() Id {
	now := getTimestamp()

	f.mu.Lock()
	sequence := f.sequence

	// Use the sequence number if the id request is in the same millisecond as
	// the previous request.
	if now <= f.prevTime {
		now = f.prevTime
		sequence++
	} else {
		sequence = 0
	}

	// Bump the timestamp by 1ms if we run out of sequence bits.
	if sequence > MaxSequence {
		now++
		sequence = 0
	}

	f.prevTime = now
	f.sequence = sequence
	f.mu.Unlock()

	timestamp := now << (HostBits + SequenceBits)
	workerId := f.workerId << SequenceBits
	return Id(timestamp | workerId | sequence)
}

// getTimestamp returns the timestamp in milliseconds adjusted for the custom
// epoch
func getTimestamp() uint64 {
	return uint64(time.Since(Epoch).Nanoseconds() / 1e6)
}

// getHostId returns the host id using the IP address of the machine
func getHostId() (uint64, error) {
	h, err := os.Hostname()
	if err != nil {
		return 0, err
	}

	addrs, err := net.LookupIP(h)
	if err != nil {
		return 0, err
	}

	a := addrs[0].To4()
	if len(a) < 4 {
		return 0, errors.New("failed to resolve hostname")
	}

	ip := binary.BigEndian.Uint32(a)
	return uint64(ip), nil
}

// getRandomId generates random worker id
func getRandomId() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b[:]), nil
}
