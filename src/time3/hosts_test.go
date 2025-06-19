package main

import (
	_ "embed"
	"fmt"
	"sync"
	"testing"
	"time"
)

//go:embed hosts_test_expected_1.txt
var expected1 string

func Test_Hosts_toString(t *testing.T) {
	var hosts = RemoteHosts{
		set: make(map[HostInfo]uint64),
	}

	host1 := HostInfo{ip: "ip1", userAgent: "agent1"}
	host2 := HostInfo{ip: "ip2", userAgent: "agent2"}

	hosts.add(host1)
	hosts.add(host2)
	hosts.add(host1)

	got := hosts.asTable()
	if got != expected1 {
		t.Errorf("hosts.asTable(), want:\n[%s]\n, got:\n[%s]\n", expected1, got)
	}
}

//go:embed hosts_test_expected_2.txt
var expected2 string

func Test_Hosts_concurrent(t *testing.T) {
	var hosts = RemoteHosts{
		set: make(map[HostInfo]uint64),
	}

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// "Client" number N connects 10*N+1 times.
			for range 10*n + 1 {
				hosts.add(HostInfo{
					ip:        fmt.Sprintf("ip-%d", n),
					userAgent: fmt.Sprintf("agent-%d", n),
				})
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}
	wg.Wait()

	got := hosts.asTable()
	if got != expected2 {
		t.Errorf("hosts.asTable(), want:\n[%s]\n, got:\n[%s]\n", expected2, got)
	}
}
