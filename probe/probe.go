package probe

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"net"
	"strconv"
	"time"

	cache "github.com/essentialkaos/ek/v13/cache/memory"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// ports is map with the most popular ports
var ports = []int{
	21,    // ftp
	22,    // ssh
	23,    // telnet
	25,    // smtp
	53,    // dns
	80,    // http
	110,   // pop3
	115,   // sftp
	143,   // imap
	220,   // imap3
	389,   // ldap
	443,   // https
	445,   // smb
	636,   // ldap
	990,   // ftps
	993,   // imaps
	995,   // pop3s
	1434,  // msql
	3000,  // unicorn
	3306,  // mysql or maria
	3389,  // rdp
	3690,  // subversion
	5432,  // postgres
	5800,  // vnc
	6379,  // redis or valkey
	6432,  // pgbouncer
	8080,  // http
	8443,  // https
	9000,  // gunicorn
	9042,  // cassandra
	9464,  // prometheus
	13000, // grafana
	27017, // mongo
}

// ////////////////////////////////////////////////////////////////////////////////// //

// probeCache is in-memory probing cache
var probeCache, _ = cache.New(cache.Config{DefaultExpiration: time.Hour})

// ////////////////////////////////////////////////////////////////////////////////// //

// Probe probes given IPs for accessible ports
func Probe(ips []string) []string {
	if len(ips) == 0 {
		return nil
	}

	foundPorts := map[int]bool{}

	for _, ip := range ips {
		for _, port := range ports {
			addr := fmt.Sprintf("%s:%d", ip, port)

			if !probeCache.Has(addr) {
				_, err := net.DialTimeout("tcp", addr, time.Second/10)

				if err != nil {
					probeCache.Set(addr, false)
					continue
				}

				foundPorts[port] = true
				probeCache.Set(addr, true)

			} else {
				if probeCache.Get(addr).(bool) {
					foundPorts[port] = true
				}
			}
		}
	}

	if len(foundPorts) == 0 {
		return nil
	}

	var result []string

	for _, port := range ports {
		if foundPorts[port] {
			result = append(result, strconv.Itoa(port))
		}
	}

	return result
}
