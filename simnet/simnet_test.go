package simnet

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Handler))
	defer ts.Close()

	t.Run("Banner", func(t *testing.T) {
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != 200 {
			t.Errorf("expected 200, got %v", res.StatusCode)
		}

		defer res.Body.Close()
	})

	t.Run("Download", func(t *testing.T) {
		units := []string{"k", "m"}
		for i := 0; i < 9; i++ {
			for _, unit := range units {
				url := fmt.Sprintf("%s/%d%s", ts.URL, i, unit)
				res, err := http.Get(url)
				if err != nil {
					t.Fatalf("%s: %v", url, err)
				}

				if res.StatusCode != 200 {
					t.Errorf("%s: expected status code 200, got %d", url, res.StatusCode)
				}

				n, err := io.Copy(ioutil.Discard, res.Body)
				if err != nil {
					t.Fatalf("%s: %v", url, err)
				}
				size := int64(i)
				switch unit {
				case "k", "K":
					size *= 1024
				case "m", "M":
					size *= 1024 * 1024
				}
				if size != n {
					t.Errorf("%s: expected content length %d, got %d", url, size, n)
				}
			}
		}
	})
}

func TestListenHTTP(t *testing.T) {
	port, err := ListenHTTP(context.Background())
	if err != nil {
		t.Fatal("listens an HTTP:", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%v", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal("requests", url, ":", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected HTTP code 200, got %v", resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestCreateServers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rlimit, err := getFileLimit()
	if err != nil {
		t.Fatal(err)
	}
	t.Run(fmt.Sprintf("by %v", rlimit), func(t *testing.T) {
		_, err := CreateServers(ctx, int(rlimit))
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	n := int(rlimit / 10)
	t.Run(fmt.Sprintf("by %v", n), func(t *testing.T) {
		ports, err := CreateServers(ctx, n)
		if err != nil {
			t.Fatal(err)
		}

		for _, port := range ports {
			url := fmt.Sprintf("http://127.0.0.1:%v", port)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatal("requests", url, ":", err)
			}
			if resp.StatusCode != 200 {
				t.Errorf("expected HTTP code 200, got %v", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}
