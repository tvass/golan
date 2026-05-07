package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

var zone string

func extractIP(qname string) net.IP {
	qname = strings.ToLower(qname)
	suffix := "." + zone
	if !strings.HasSuffix(qname, suffix) {
		return nil
	}
	relative := qname[:len(qname)-len(suffix)]
	if relative == "" {
		return nil
	}
	labels := strings.Split(relative, ".")
	if len(labels) < 4 {
		return nil
	}
	return parseOctets(labels[len(labels)-4:])
}

// parseOctets turns exactly four decimal strings into an IPv4 address.
func parseOctets(parts []string) net.IP {
	if len(parts) != 4 {
		return nil
	}
	var b [4]byte
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil
		}
		b[i] = byte(n)
	}
	return net.IPv4(b[0], b[1], b[2], b[3])
}

func handle(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false

	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA, dns.TypeANY:
			ip := extractIP(q.Name)
			if ip == nil {
				ip = net.IPv4(127, 0, 0, 1)
			}
			log.Printf("A  %s → %s", q.Name, ip)
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: ip.To4(),
			})
			// AAAA / other types: NOERROR with empty answer.
		}
	}

	w.WriteMsg(m)
}

func handlePTR(w dns.ResponseWriter, r *dns.Msg) {
	const arpa = "in-addr.arpa."

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false

	for _, q := range r.Question {
		if q.Qtype != dns.TypePTR && q.Qtype != dns.TypeANY {
			continue
		}

		// q.Name example: "107.0.168.192.in-addr.arpa."
		name := strings.ToLower(q.Name)
		suffix := "." + arpa
		if !strings.HasSuffix(name, suffix) {
			continue
		}

		reversed := name[:len(name)-len(suffix)] // "107.0.168.192"
		parts := strings.Split(reversed, ".")
		if len(parts) != 4 {
			continue
		}

		octets := []string{parts[3], parts[2], parts[1], parts[0]}
		ip := parseOctets(octets)
		if ip == nil {
			continue
		}

		b := ip.To4()
		ptrTarget := fmt.Sprintf("%d.%d.%d.%d.%s", b[0], b[1], b[2], b[3], zone)
		log.Printf("PTR %s → %s", q.Name, ptrTarget)
		m.Answer = append(m.Answer, &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			Ptr: ptrTarget,
		})
	}

	w.WriteMsg(m)
}

func main() {
	port := flag.Int("port", 15353, "UDP/TCP port to listen on")
	z := flag.String("zone", "local.lan", "DNS zone to serve")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	zone = dns.Fqdn(*z)

	dns.HandleFunc(zone, handle)
	dns.HandleFunc("in-addr.arpa.", handlePTR)
	dns.HandleFunc(".", handle) // catch-all: unknown zones → 127.0.0.1

	log.Printf("starting DNS server on %s for zone %s + in-addr.arpa", addr, zone)

	errCh := make(chan error, 2)
	for _, proto := range []string{"udp", "tcp"} {
		srv := &dns.Server{Addr: addr, Net: proto}
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Printf("error on %s/%s: %v", addr, srv.Net, err)
				errCh <- err
			}
		}()
	}

	log.Fatal(<-errCh)
}
