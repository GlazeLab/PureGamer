package pinging

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	icmping "github.com/go-ping/ping"
	pool "github.com/libp2p/go-buffer-pool"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"io"
	mrand "math/rand"
	"net"
	"time"
)

func Ping(ctx context.Context, h host.Host, p peer.ID) (time.Duration, error) {
	var err error
	var s network.Stream
	s, err = h.NewStream(network.WithUseTransient(ctx, "ping"), p, protocol)
	if err != nil {
		return 0, err
	}

	if err = s.Scope().SetService(ServiceName); err != nil {
		log.Debugf("error attaching stream to ping service: %s", err)
		s.Reset()
		return 0, err
	}

	b := make([]byte, 8)
	if _, err = rand.Read(b); err != nil {
		log.Errorf("failed to get cryptographic random: %s", err)
		s.Reset()
		return 0, err
	}

	ra := mrand.New(mrand.NewSource(int64(binary.BigEndian.Uint64(b))))

	var duration time.Duration
	duration, err = ping(s, ra)
	if err != nil {
		log.Debugf("error pinging peer %s: %s", p, err)
		s.Reset()
		return 0, err
	}
	return duration, nil
}

func ping(s network.Stream, randReader io.Reader) (time.Duration, error) {
	if err := s.Scope().ReserveMemory(2*PingSize, network.ReservationPriorityAlways); err != nil {
		log.Debugf("error reserving memory for ping stream: %s", err)
		s.Reset()
		return 0, err
	}
	defer s.Scope().ReleaseMemory(2 * PingSize)

	buf := pool.Get(PingSize)
	defer pool.Put(buf)

	if _, err := io.ReadFull(randReader, buf); err != nil {
		return 0, err
	}

	before := time.Now()
	if _, err := s.Write(buf); err != nil {
		return 0, err
	}

	rbuf := pool.Get(PingSize)
	defer pool.Put(rbuf)

	if _, err := io.ReadFull(s, rbuf); err != nil {
		return 0, err
	}

	if !bytes.Equal(buf, rbuf) {
		return 0, errors.New("ping packet was incorrect")
	}

	return time.Since(before), nil
}

func PingICMP(host string, port uint64) (time.Duration, error) {
	pinger, err := icmping.NewPinger(host)
	if err != nil {
		return 0, err
	}
	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics().AvgRtt
	return stats, err
}

func PingTCP(host string, port uint64) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(10)*time.Second)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	return time.Since(start), nil
}
