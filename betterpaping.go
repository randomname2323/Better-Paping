package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

type PingStats struct {
	Attempts      int
	Successes     int
	Failures      int
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
}

func main() {
	if len(os.Args) < 2 {
		show_usage()
		return
	}

	mode := strings.ToLower(os.Args[1])

	switch mode {
	case "tcp":
		run_tcp_mode()
	case "web":
		run_web_mode()
	case "mc":
		run_mc_mode()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		show_usage()
	}
}

func show_usage() {
	fmt.Println("Usage: better_paping <mode> [options]")
	fmt.Println("\nModes:")
	fmt.Println("  tcp   - Standard TCP port ping")
	fmt.Println("  web   - HTTP/HTTPS website ping")
	fmt.Println("  mc    - Minecraft server ping")
	fmt.Println("\nTry 'better_paping <mode>' for specific options.")
}

// --- TCP MODE ---

func run_tcp_mode() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: better_paping tcp <host> [port] [interval_ms] [duration_sec]")
		return
	}

	host := os.Args[2]
	port_str := "80"
	timeout_ms := 650
	time_sec := 0

	if len(os.Args) > 3 {
		port_str = os.Args[3]
	}
	if len(os.Args) > 4 {
		v, _ := strconv.Atoi(os.Args[4])
		if v > 0 {
			timeout_ms = v
		}
	}
	if len(os.Args) > 5 {
		v, _ := strconv.Atoi(os.Args[5])
		if v >= 0 {
			time_sec = v
		}
	}

	p, err := strconv.Atoi(port_str)
	if err != nil {
		fmt.Println("bad port")
		return
	}

	interval := time.Duration(timeout_ms) * time.Millisecond
	var total_time time.Duration
	if time_sec == 0 {
		total_time = time.Duration(1<<63 - 1)
	} else {
		total_time = time.Duration(time_sec) * time.Second
	}

	data := &PingStats{MinDuration: time.Duration(1<<63 - 1)}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go do_paping(host, p, interval, total_time, data)

	select {
	case <-c:
		fmt.Println("\nStopped.")
		print_tcp_stats(host, p, interval, *data)
	case <-time.After(total_time):
		print_tcp_stats(host, p, interval, *data)
	}
}

func do_paping(ip string, p int, t, d time.Duration, s *PingStats) {
	end := time.Now().Add(d)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for time.Now().Before(end) {
		s.Attempts++
		start := time.Now()
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, p), t)
		if err == nil {
			s.Successes++
			conn.Close()
			dur := time.Since(start)
			s.TotalDuration += dur
			if dur < s.MinDuration {
				s.MinDuration = dur
			}
			if dur > s.MaxDuration {
				s.MaxDuration = dur
			}

			time_color := green
			if dur > t/2 {
				time_color = color.New(color.FgYellow).SprintFunc()
			}
			if dur >= t {
				time_color = red
			}

			fmt.Printf("Connected to %s: time=%s protocol=%s port=%s\n",
				green(ip), time_color(fmt.Sprintf("%.2fms", float64(dur.Microseconds())/1000)), green("TCP"), green(strconv.Itoa(p)))
		} else {
			s.Failures++
			fmt.Printf("%s\n", red("Connection timed out"))
		}
		time.Sleep(t)
	}
}

func print_tcp_stats(ip string, p int, t time.Duration, s PingStats) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	fmt.Println("\n══════════════════════ TCP Paping Statistics ══════════════════════")
	fmt.Printf("🎯 Target: %s:%s\n", green(ip), green(strconv.Itoa(p)))
	fmt.Printf("🔄 Attempts: %s\n", green(strconv.Itoa(s.Attempts)))
	fmt.Printf("✅ Successful: %s\n", green(strconv.Itoa(s.Successes)))
	fmt.Printf("❌ Failed: %s\n", green(strconv.Itoa(s.Failures)))
	rate := 0.0
	if s.Attempts > 0 {
		rate = float64(s.Successes) / float64(s.Attempts) * 100
	}
	fmt.Printf("📊 Success rate: %s%%\n", green(fmt.Sprintf("%.2f", rate)))
	if s.Successes > 0 {
		avg := s.TotalDuration / time.Duration(s.Successes)
		fmt.Printf("⏱ Average response time: %s\n", green(fmt.Sprintf("%.2fms", float64(avg.Microseconds())/1000)))
		fmt.Printf("⚡ Fastest response time: %s\n", green(fmt.Sprintf("%.2fms", float64(s.MinDuration.Microseconds())/1000)))
		if s.MaxDuration >= t {
			fmt.Printf("🐢 Slowest response time: %s\n", red(fmt.Sprintf("%.2fms", float64(s.MaxDuration.Microseconds())/1000)))
		} else {
			fmt.Printf("🐢 Slowest response time: %s\n", green(fmt.Sprintf("%.2fms", float64(s.MaxDuration.Microseconds())/1000)))
		}
	}
	fmt.Println("══════════════════════ 🔔 TCP Paping Complete ══════════════════════")
}

// --- WEB MODE ---

func run_web_mode() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: better_paping web <url> [interval_ms] [duration_sec]")
		return
	}

	target_url := os.Args[2]
	timeout_ms := 650
	time_sec := 0

	if len(os.Args) > 3 {
		v, _ := strconv.Atoi(os.Args[3])
		if v > 0 {
			timeout_ms = v
		}
	}
	if len(os.Args) > 4 {
		v, _ := strconv.Atoi(os.Args[4])
		if v >= 0 {
			time_sec = v
		}
	}

	wait := time.Duration(timeout_ms) * time.Millisecond
	var max_time time.Duration
	if time_sec == 0 {
		max_time = time.Duration(1<<63 - 1)
	} else {
		max_time = time.Duration(time_sec) * time.Second
	}

	data := &PingStats{MinDuration: time.Duration(1<<63 - 1)}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go run_web_stats(target_url, wait, max_time, data)

	select {
	case <-c:
		fmt.Println("\nStopped.")
		print_web_stats(target_url, wait, *data)
	case <-time.After(max_time):
		print_web_stats(target_url, wait, *data)
	}
}

func run_web_stats(url string, interval, duration time.Duration, s *PingStats) {
	deadline := time.Now().Add(duration)
	client := &http.Client{Timeout: interval}
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for time.Now().Before(deadline) {
		s.Attempts++
		now := time.Now()
		resp, err := client.Get(url)
		took := time.Since(now)
		if err == nil {
			s.Successes++
			s.TotalDuration += took
			if took < s.MinDuration {
				s.MinDuration = took
			}
			if took > s.MaxDuration {
				s.MaxDuration = took
			}

			time_color := green
			if took > interval/2 {
				time_color = color.New(color.FgYellow).SprintFunc()
			}
			if took >= interval {
				time_color = red
			}

			fmt.Printf("Connected to %s: time=%s status=%s\n",
				green(url), time_color(fmt.Sprintf("%.2fms", float64(took.Microseconds())/1000)), green(resp.Status))
			resp.Body.Close()
		} else {
			s.Failures++
			fmt.Printf("%s: %v\n", red("Connection failed to "+url), err)
		}
		time.Sleep(interval)
	}
}

func print_web_stats(url string, t time.Duration, s PingStats) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	fmt.Println("\n\033[1m══════════════════════ Web Paping Statistics ══════════════════════\033[0m")
	fmt.Printf("🎯 Target: %s\n", green(url))
	fmt.Printf("🔄 Attempts: %s\n", green(strconv.Itoa(s.Attempts)))
	fmt.Printf("✅ Successful: %s\n", green(strconv.Itoa(s.Successes)))
	fmt.Printf("❌ Failed: %s\n", green(strconv.Itoa(s.Failures)))
	rate := 0.0
	if s.Attempts > 0 {
		rate = float64(s.Successes) / float64(s.Attempts) * 100
	}
	fmt.Printf("📊 Success rate: %s%%\n", green(fmt.Sprintf("%.2f", rate)))
	if s.Successes > 0 {
		avg := s.TotalDuration / time.Duration(s.Successes)
		fmt.Printf("⏱ Average response time: %s\n", green(fmt.Sprintf("%.2fms", float64(avg.Microseconds())/1000)))
		fmt.Printf("⚡ Fastest response time: %s\n", green(fmt.Sprintf("%.2fms", float64(s.MinDuration.Microseconds())/1000)))
		if s.MaxDuration >= t {
			fmt.Printf("🐢 Slowest response time: %s\n", red(fmt.Sprintf("%.2fms", float64(s.MaxDuration.Microseconds())/1000)))
		} else {
			fmt.Printf("🐢 Slowest response time: %s\n", green(fmt.Sprintf("%.2fms", float64(s.MaxDuration.Microseconds())/1000)))
		}
	}
	fmt.Println("\033[1m══════════════════════ 🔔 Web Paping Complete ══════════════════════\033[0m")
}

// --- MC MODE ---

func run_mc_mode() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: better_paping mc <host> [port] [interval_ms] [duration_sec]")
		return
	}

	host := os.Args[2]
	port := 25565
	wait_ms := 650
	limit_sec := 0

	if len(os.Args) > 3 {
		v, _ := strconv.Atoi(os.Args[3])
		if v > 0 {
			port = v
		}
	}
	if len(os.Args) > 4 {
		v, _ := strconv.Atoi(os.Args[4])
		if v > 0 {
			wait_ms = v
		}
	}
	if len(os.Args) > 5 {
		v, _ := strconv.Atoi(os.Args[5])
		if v >= 0 {
			limit_sec = v
		}
	}

	wait := time.Duration(wait_ms) * time.Millisecond
	var dur time.Duration
	if limit_sec == 0 {
		dur = time.Duration(1<<63 - 1)
	} else {
		dur = time.Duration(limit_sec) * time.Second
	}

	data := &PingStats{MinDuration: time.Duration(1<<63 - 1)}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go do_mc_ping(host, port, wait, dur, data)

	select {
	case <-c:
		fmt.Println("\nStopped.")
		show_mc_results(host, port, *data)
	case <-time.After(dur):
		show_mc_results(host, port, *data)
	}
}

func do_mc_ping(host string, port int, wait, total time.Duration, s *PingStats) {
	end := time.Now().Add(total)
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	target, p := resolve_srv(host, port)
	ips := get_ips(target)

	if len(ips) == 0 {
		fmt.Printf("%s Cannot resolve %s\n", red("Error:"), host)
		return
	}

	idx := 0
	for time.Now().Before(end) {
		s.Attempts++
		now := time.Now()
		var usedIP string
		var info string
		var pingErr error

		for i := 0; i < len(ips); i++ {
			ip := ips[idx%len(ips)]
			idx++
			addr := net.JoinHostPort(ip, strconv.Itoa(p))
			_, res, err := test_mc(addr, target, 2500*time.Millisecond)
			if err == nil && !strings.HasPrefix(res, "Offline") {
				usedIP = ip
				info = res
				pingErr = nil
				break
			}
			pingErr = err
		}

		took := time.Since(now)
		s.TotalDuration += took
		if took < s.MinDuration {
			s.MinDuration = took
		}
		if took > s.MaxDuration {
			s.MaxDuration = took
		}

		if pingErr == nil && info != "" {
			s.Successes++
			ts := fmt.Sprintf("%.2fms", float64(took.Microseconds())/1000)
			colored := green(ts)
			if took > 300*time.Millisecond {
				colored = yellow(ts)
			}
			if took > 700*time.Millisecond {
				colored = red(ts)
			}

			fmt.Printf("Connected to %s (%s): time=%s proto=%s port=%d\n", green(host), usedIP, colored, green(info), p)
		} else {
			s.Failures++
			fmt.Println(red("MC Server Down"))
		}
		time.Sleep(wait)
	}
}

func resolve_srv(domain string, def_port int) (string, int) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", domain)
	if err == nil && len(addrs) > 0 {
		srv := addrs[0]
		target := strings.TrimSuffix(srv.Target, ".")
		if target != "" {
			return target, int(srv.Port)
		}
	}
	return domain, def_port
}

func get_ips(host string) []string {
	var ips []string
	addrs, _ := net.LookupIP(host)
	for _, ip := range addrs {
		ips = append(ips, ip.String())
	}
	if len(ips) == 0 {
		ips = []string{host}
	}
	return ips
}

func test_mc(addr, host string, timeout time.Duration) (string, string, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))
	proto, name, err := mc_handshake(conn, host)
	if err == nil {
		if name == "" {
			name = fmt.Sprintf("proto %d", proto)
		}
		if proto <= 47 || proto == 0 {
			return addr, "Offline", nil
		}
		return addr, name, nil
	}
	return addr, "", fmt.Errorf("fail")
}

func mc_handshake(conn net.Conn, host string) (int, string, error) {
	hs := new(bytes.Buffer)
	hs.WriteByte(0x00)
	put_varint(hs, 768)
	put_str(hs, host)
	binary.Write(hs, binary.BigEndian, uint16(25565))
	put_varint(hs, 1)

	pkt := new(bytes.Buffer)
	put_varint(pkt, hs.Len())
	pkt.Write(hs.Bytes())
	pkt.Write([]byte{0x01, 0x00})

	_, err := conn.Write(pkt.Bytes())
	if err != nil {
		return 0, "", err
	}

	l, err := get_varint(conn)
	if err != nil || l < 3 {
		return 0, "", err
	}
	pid, err := get_varint(conn)
	if err != nil || pid != 0x00 {
		return 0, "", fmt.Errorf("pid")
	}
	js_l, err := get_varint(conn)
	if err != nil || js_l < 10 {
		return 0, "", err
	}
	js_d := make([]byte, js_l)
	_, err = io.ReadFull(conn, js_d)
	if err != nil {
		return 0, "", err
	}

	var status struct {
		Version struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		} `json:"version"`
	}
	if err := json.Unmarshal(js_d, &status); err != nil {
		return 0, "", err
	}
	return status.Version.Protocol, status.Version.Name, nil
}

func show_mc_results(host string, port int, s PingStats) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println("\n══════════════════════ Minecraft Paping Statistics ══════════════════════")
	fmt.Printf("🎯 Target: %s:%d\n", green(host), port)
	fmt.Printf("🔄 Attempts: %s\n", green(strconv.Itoa(s.Attempts)))
	fmt.Printf("✅ Successful: %s\n", green(strconv.Itoa(s.Successes)))
	fmt.Printf("❌ Failed: %s\n", red(strconv.Itoa(s.Failures)))

	rate := 0.0
	if s.Attempts > 0 {
		rate = float64(s.Successes) / float64(s.Attempts) * 100
	}
	fmt.Printf("📊 Success rate: %s%%\n", green(fmt.Sprintf("%.1f", rate)))

	if s.Successes > 0 {
		avg := s.TotalDuration / time.Duration(s.Successes)
		fmt.Printf("⏱ Average response time: %s\n", green(fmt.Sprintf("%.0fms", float64(avg.Microseconds())/1000)))
		fmt.Printf("⚡ Fastest response time: %s\n", green(fmt.Sprintf("%.0fms", float64(s.MinDuration.Microseconds())/1000)))
		sc := green
		if s.MaxDuration > 700*time.Millisecond {
			sc = red
		}
		fmt.Printf("🐢 Slowest response time: %s\n", sc(fmt.Sprintf("%.0fms", float64(s.MaxDuration.Microseconds())/1000)))
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

func put_varint(w io.Writer, val int) {
	for {
		b := byte(val & 0x7F)
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		w.Write([]byte{b})
		if val == 0 {
			break
		}
	}
}

func put_str(w io.Writer, s string) {
	put_varint(w, len(s))
	w.Write([]byte(s))
}

func get_varint(r io.Reader) (int, error) {
	var val uint
	var shift uint
	for i := 0; i < 5; i++ {
		b := make([]byte, 1)
		_, err := r.Read(b)
		if err != nil {
			return 0, err
		}
		val |= uint(b[0]&0x7F) << shift
		if b[0]&0x80 == 0 {
			return int(val), nil
		}
		shift += 7
	}
	return 0, fmt.Errorf("err")
}
