package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func buildSIPOptions(dstIP string, dstPort int, srcIP string, srcPort int, callID, branchID string, sipUser, toUser string) []byte {
	crlf := "\r\n"
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("OPTIONS sip:%s:%d SIP/2.0%s", dstIP, dstPort, crlf))
	b.WriteString(fmt.Sprintf("Via: SIP/2.0/UDP %s:%d;branch=%s%s", srcIP, srcPort, branchID, crlf))
	b.WriteString(fmt.Sprintf("From: <sip:%s@%s>;tag=monitor%s", sipUser, srcIP, crlf))
	b.WriteString(fmt.Sprintf("To: <sip:%s@%s>%s", toUser, dstIP, crlf))
	b.WriteString(fmt.Sprintf("Contact: <sip:%s@%s:%d>%s", sipUser, srcIP, srcPort, crlf))
	b.WriteString(fmt.Sprintf("Call-ID: %s%s", callID, crlf))
	b.WriteString("CSeq: 1 OPTIONS" + crlf)
	b.WriteString("Max-Forwards: 70" + crlf)
	b.WriteString("User-Agent: SIP Monitor" + crlf)
	b.WriteString("Content-Length: 0" + crlf)
	b.WriteString(crlf)
	return b.Bytes()
}

func parseStatusAndHeaders(payload []byte) (status string, callID string) {
	lines := bytes.Split(payload, []byte{'\n'})
	for i, raw := range lines {
		ln := strings.TrimRight(string(raw), "\r")
		if i == 0 && strings.HasPrefix(ln, "SIP/2.0") {
			status = strings.TrimSpace(ln)
			continue
		}
		if strings.TrimSpace(ln) == "" {
			break
		}
		if k, v, ok := splitHeader(ln); ok && strings.EqualFold(k, "Call-ID") {
			callID = strings.TrimSpace(v)
		}
	}
	return
}

func splitHeader(line string) (key, val string, ok bool) {
	i := strings.Index(line, ":")
	if i <= 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:]), true
}

func dialWithBind(dstIP string, dstPort int, srcIP string, srcPort int) (*net.UDPConn, error) {
	raddr := &net.UDPAddr{IP: net.ParseIP(dstIP), Port: dstPort}

	// Tenta bindar em srcIP:srcPort
	if ip := net.ParseIP(srcIP); ip != nil {
		laddr := &net.UDPAddr{IP: ip, Port: srcPort}
		if conn, err := net.DialUDP("udp", laddr, raddr); err == nil {
			return conn, nil
		}
		// fallback: 0.0.0.0:srcPort (preserva a porta)
		laddr2 := &net.UDPAddr{IP: net.IPv4zero, Port: srcPort}
		if conn, err := net.DialUDP("udp", laddr2, raddr); err == nil {
			return conn, nil
		}
	}
	// último fallback: sem bind explícito
	return net.DialUDP("udp", nil, raddr)
}

func sendOnce(dstIP string, kamPort int, srcIP string, srcPort int, timeout time.Duration, sipUser, toUser string) string {
	callID := uuid.NewString()
	branchID := "z9hG4bK-" + uuid.NewString()
	payload := buildSIPOptions(dstIP, kamPort, srcIP, srcPort, callID, branchID, sipUser, toUser)

	conn, err := dialWithBind(dstIP, kamPort, srcIP, srcPort)
	if err != nil {
		return fmt.Sprintf("%s - ERROR dial: %v", dstIP, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(payload); err != nil {
		return fmt.Sprintf("%s - ERROR write: %v", dstIP, err)
	}

	buf := make([]byte, 65535)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return fmt.Sprintf("%s - TIMEOUT", dstIP)
		}
		return fmt.Sprintf("%s - ERROR read: %v", dstIP, err)
	}

	status, rCallID := parseStatusAndHeaders(buf[:n])
	if status == "" {
		return fmt.Sprintf("%s - no SIP status line", dstIP)
	}
	if strings.HasPrefix(status, "SIP/2.0 200") {
		if rCallID == callID {
			return fmt.Sprintf("%s - 200 OK (Call-ID OK)", dstIP)
		}
		return fmt.Sprintf("%s - 200 OK (Call-ID mismatch: expected %s got %s)", dstIP, callID, rCallID)
	}
	return fmt.Sprintf("%s - unexpected: %s", dstIP, status)
}

func main() {
	var (
		srcIP   string
		srcPort int
		kamPort int
		ipsCSV  string
		timeout int
		sipUser string
		toUser  string
	)

	flag.StringVar(&srcIP, "src-ip", "", "IP de origem para o pacote UDP (obrigatório)")
	flag.IntVar(&srcPort, "src-port", 0, "Porta de origem UDP (obrigatório)")
	flag.IntVar(&kamPort, "kam-port", 5060, "Porta do destino (Kamailio/SBC), padrão 5060")
	flag.StringVar(&ipsCSV, "ips", "", "Lista de IPs separada por vírgula (ex: 187.60.51.10,187.60.51.12) (obrigatório)")
	flag.IntVar(&timeout, "timeout", 2, "Timeout em segundos para aguardar resposta (padrão 2)")
	flag.StringVar(&sipUser, "sip-user", "SIPMonitor", "Usuário SIP para From/Contact")
	flag.StringVar(&toUser, "to-user", "SIPMonitor", "Usuário SIP para To")
	flag.Parse()

	if srcIP == "" || srcPort == 0 || ipsCSV == "" {
		fmt.Println("uso:")
		fmt.Println("  ./sip-options-once -src-ip <IP> -src-port <PORTA> -ips <ip1,ip2,...> [-kam-port 5060] [-timeout 2] [-sip-user SIPMonitor] [-to-user SIPMonitor]")
		os.Exit(2)
	}

	ips := make([]string, 0, 8)
	for _, p := range strings.Split(ipsCSV, ",") {
		ip := strings.TrimSpace(p)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		fmt.Println("lista de IPs vazia em -ips")
		os.Exit(2)
	}

	for _, ip := range ips {
		res := sendOnce(ip, kamPort, srcIP, srcPort, time.Duration(timeout)*time.Second, sipUser, toUser)
		fmt.Println(res)
	}
}
