// This program advertises a mDNS service and proxies ports in order to make a Unistellar telescope available on another network.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/grandcat/zeroconf"
)

var (
	name       = flag.String("name", "unistellar-proxy", "The name for the service.")
	port       = flag.Int("port", 13012, "The port advertised for the service.")
	proxyTo    = flag.String("proxy-to", "192.168.100.1", "The target address for the service.")
	proxyPorts = flag.String("proxy-ports", "13007,13009,13012", "The ports for the proxy service.")
)

func parsePortList(s string) []int {
	var ports []int
	for _, port := range strings.Split(s, ",") {
		p, err := strconv.Atoi(strings.TrimSpace(port))
		if err != nil {
			log.Fatalf("Invalid port number: %s", port)
		}
		ports = append(ports, p)
	}
	return ports
}

// proxy listens on the specified port and forwards connections to the target address.
func proxy(port int, target string, stop chan struct{}) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}
	defer listener.Close()

	log.Printf("Proxying :%d to %s", port, target)
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stop:
				return
			default:
				log.Printf("Failed to accept connection on port %d: %v", port, err)
				continue
			}
		}
		go handleProxyConnection(conn, target)
	}
}

func handleProxyConnection(conn net.Conn, target string) {
	defer conn.Close()
	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", target, err)
		return
	}
	defer targetConn.Close()

	go func() {
		defer conn.Close()
		defer targetConn.Close()
		io.Copy(targetConn, conn)
	}()
	io.Copy(conn, targetConn)
}

func main() {
	flag.Parse()
	ports := parsePortList(*proxyPorts)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for _, p := range ports {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			proxy(p, fmt.Sprintf("%s:%d", *proxyTo, p), stop)
		}()
	}
	server, err := zeroconf.Register(
		*name,
		"_evsoft-control._tcp",
		"local.",
		*port,
		nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Shutdown()

	log.Printf("Advertising mDNS service %q at port %d with target %s", *name, *port, *proxyTo)
	wg.Wait()
}
