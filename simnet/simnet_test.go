package simnet

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

func ifNameOf(ipAddr string) string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.String() == ipAddr {
				return iface.Name
			}
		}
	}

	return ""
}

var loName = ifNameOf("127.0.0.1")

func TestListenHTTP(t *testing.T) {
	port, err := ListenHTTP()
	if err != nil {
		t.Fatal("listens an HTTP:", err)
	}

	deviceFlag := fmt.Sprintf("--device=%v", loName)
	b, err := exec.Command("comcast", deviceFlag, "--target-bw=8", "--default-bw=5", "--target-proto=tcp", "--target-addr=127.0.0.1", fmt.Sprintf("--target-port=%v", port)).CombinedOutput()
	if b != nil {
		t.Log(string(b))
	}
	if err != nil {
		t.Fatal(err)
	} else {
		defer exec.Command("comcast", deviceFlag, "--stop").Run()
	}

	since := time.Now()
	url := fmt.Sprintf("http://127.0.0.1:%v/4k", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal("requests", url, ":", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected HTTP code 200, got %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	n, err := io.Copy(ioutil.Discard, resp.Body)
	if elapsed := time.Since(since); elapsed < 4*time.Second {
		t.Errorf("expected elapse >= 4s, got %v", elapsed)
	}

	if err != nil {
		t.Errorf("reading from %v: %v", url, err)
	} else if n != 4*1024 {
		t.Errorf("expected content size 4k, got %v", n)
	}
}
