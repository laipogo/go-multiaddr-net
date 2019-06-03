package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

// flags
var formats = []string{"string", "bytes", "hex", "slice"}
var format string
var hideLoopback bool

func init() {
	fmt.Println("multiaddr init...")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [<multiaddr>]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}

	usage := fmt.Sprintf("output format, one of: %v", formats)
	flag.StringVar(&format, "format", "string", usage)
	flag.StringVar(&format, "f", "string", usage+" (shorthand)")
	flag.BoolVar(&hideLoopback, "hide-loopback", false, "do not display loopback addresses")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		output(localAddresses()...)
	} else {
		output(address(args[0]))
	}
}

func getOutboundIP() (string, error) {
	fmt.Println("get outbound ip")
	resp, err := http.Get("http://ifconfig.me/ip")
	if err == nil {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return string(bytes), nil
		}
	}
	return "", err
}

func localAddresses() []ma.Multiaddr {
	maddrs, err := manet.InterfaceMultiaddrs()
	if err != nil {
		die(err)
	}

	ip, err := getOutboundIP()
	if err == nil {
		addr, err := net.ResolveIPAddr("", ip)
		if err == nil {
			multiaddr, err := manet.FromNetAddr(addr)
			if err == nil {
				maddrs = append(maddrs, multiaddr)
			}
		}
	}

	if !hideLoopback {
		return maddrs
	}

	var maddrs2 []ma.Multiaddr
	for _, a := range maddrs {
		if !manet.IsIPLoopback(a) {
			maddrs2 = append(maddrs2, a)
		}
	}

	return maddrs2
}

func address(addr string) ma.Multiaddr {
	m, err := ma.NewMultiaddr(addr)
	if err != nil {
		die(err)
	}

	return m
}

func output(ms ...ma.Multiaddr) {
	for _, m := range ms {
		fmt.Println(outfmt(m))
	}
}

func outfmt(m ma.Multiaddr) string {
	switch format {
	case "string":
		return m.String()
	case "slice":
		return fmt.Sprintf("%v", m.Bytes())
	case "bytes":
		return string(m.Bytes())
	case "hex":
		return "0x" + hex.EncodeToString(m.Bytes())
	}

	die("error: invalid format", format)
	return ""
}

func die(v ...interface{}) {
	fmt.Fprint(os.Stderr, v...)
	fmt.Fprint(os.Stderr, "\n")
	flag.Usage()
	os.Exit(-1)
}
