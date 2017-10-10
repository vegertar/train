<!-- TOC -->

- [Train](#train)
    - [ISP & Region](#isp--region)
        - [First-Class Core](#first-class-core)
        - [Second-Class Influx](#second-class-influx)
        - [Third-Class Capital](#third-class-capital)
    - [Reachable Restriction](#reachable-restriction)
    - [Simulation](#simulation)
        - [Basic HTTP Server](#basic-http-server)
        - [Speed Measurement](#speed-measurement)
        - [Isolation](#isolation)
        - [Connectivity](#connectivity)

<!-- /TOC -->
# Train

This document briefly described the minimum restriction of interconnect ability between two IDCs and how to build connections above these restrictions cross all over network (currently only covering China mainland).

## ISP & Region

The Internet of China (mainland) has been working with 3 biggest backbone networks, CTC, CNC, and CMCC. Further more, there has been also some significant ISPs like GWBN, CNNET, etc. Different ISP configuring different routing policy in order to purchase traffic within their own network, which resulting some typical issues like IP hijacking, network unreachable, high latency, and so on.

Typically, internet company use CDN to hide underlying latency, provide ISP transparent service to their users. But the point is, most CDN providers would rather use cache technology, which makes users *avoid* latency, but the CDN itself *doesn't*. So, let's face it.

The most ideal solution is we are an ISP, we have our own IDCs, cables, AS, and BGP announcements, all internal packets are routing under our switches, and all rest ISPs are peering with us. For too much reasons we have no plans to become a new ISP, so skip to next solution.

The second and cheapest idea is an overlay network, like CDN does, but this time we focus on latency itself, not the edge users concerned resource latency.

Suppose we have hundreds of IDCs located in hundreds of cities serve whole country with tens of ISPs. We have no idea that how these ISPs built their network bone and how they peering with others and how they discarding packets from or to other ISPs or even other regions. Fortunately, we could be inspiring from geography.

China (mainland) is separated into 7 districts. Most of citizens are living in southeast, especially, there are 3 super cities, Beijing, Shanghai, and Guangzhou, which are also the heart of China-North, China-East and China-South respectively. So China (mainland) is supposed to be organized as a 3-layer tree in geography.

### First-Class Core

Obviously, first-class cores consist of all 3 super cities: Beijing, Shanghai, Guangzhou. Not only these cities are the actual cores of administration, economy and high-tech, but also Bei-Shang-Guang is the most incredible population influx, any company who want to serve whole country must take meaningful projects on these places, as we have seen, the 3 biggest ISPs (CNC, CTC, CMCC) are built on Beijing, the largest third-party DC provider (TianDiXiangYun) of China-South is based on Beijing, too. In other words, if we are going to communicate with the rest of China, we need be able to communicate with Bei-Shang-Guang directly or indirectly.

### Second-Class Influx

China is big enough that no single region can provide low-latency and low-price services simultaneously for all citizens, that's why we use a CDN. As the most other companies did, for example, China Railway built branch railways on Shenyang, Wuhan, Chengdu, etc., let's order every administrative center city of 7 districts as the second-class convergences, they are responsible for taking communicates between Bei-Shang-Guang and local district, or communicating with other districts.

### Third-Class Capital

All province capitals except of mentioned cities early are considered as the third-class capitals, they are responsible for taking communicates between second-class influx of present district and local region, or communicating with other capitals within same district.

From now on, any province capitals in China can talk to each other at least in logical organization chart. For example, if Kunming would like to reach to Hangzhou, the routing path might be `Kunming->Chengdu->Shanghai->Hangzhou`.

Taking a logical path is really far away from building an actually working route, as such similar structures are what the traditional CDN does, which doesn't accomplish hiding the network differences when routing through multiple ISPs. And additionally, *why we define a 3-layer tree structure?* We will explain these problems at later sections.

## Reachable Restriction

Usually, latency is restricted by distance if determines medium as copper or fiber. But in real internet, medium is not only limited by material, the major limitation is companies, i.e. ISPs.

By qualifying on a specific ISP, let's define the *province* as the minimum communication cell in geography space. Why don't we use a smaller region like a city or town instead? Since [fibers are widely used in China](http://news.163.com/16/0323/10/BIR9PA4Q00014JB5.html), it's no big deal to treat them equally.

So, the first restriction can be given as below.

> Any IDC can communicate with another IDC within same province, same ISP.

This restriction is intuitive, easily understandable. If we draw it as a graph, then which can be expressed in what any two (IDC) vertices are connectable (within same province, same ISP), that means it must be a **strongly connected graph**.

In contrast, is it actually possible that there has been an IDC existing as an alone vertex that has no communication with outside? Yes, definitely. But no communication with outside doesn't means completely unreachable between inside and outside, otherwise how the company manages this IDC? So the definition of the first restriction could be given again in more general.

> All IDCs within same province and same ISP construct a **connected graph**.

How about an ISP services over multiple provinces under a district?

We have already defined a province as a minimal cell, so for multiple ones which can be shaped as either peering or layering. On layering, formally define the second-class influx as a *gateway*, which in charge of the communication among attached provinces. This relationship are easily keeping even zooming the service out over multiple districts, then the *gateway* is the first-class core. So giving a new restriction over multiple regions for an ISP.

> Existing an IDC on capital of down region can reach to an IDC on capital of up region with same ISP.

For instance, if `Dali` reaches to `Ningbo`, then all the capitals are *Kunming*, **Chengdu**, ***Shanghai***, *Hangzhou*.

Because of business competitions and government policies, no single ISP is large enough to serve everywhere, so as we mentioned early, there are multiple ISPs eventually in the same province.

Can multiple IDCs from different ISPs be interconnected? There is no promise, furthermore, this is why the multi-room exists. However, the worse is, a multi-room cannot swear for communicating with any unknown ISP, otherwise it will become the incredible and impossible single ISP covering everywhere.

But we are working on *manageable* IDCs. Which means any IDC is finally reachable from or to our central management machine, the routing paths, though hidden, are living. To reduce the number of ISPs we have to directly consider of, it is necessary to define a **base** of ISPs that our management machine built on. For instance, a typical ISP base consists of CNC, CTC, CMCC and any one tagged *multi-room*.

After defining base, we can give the last restriction.

> Existing an IDC can reach to one of the base ISPs for any ISP within same province.

Considering of any ISP unable to reach from or to base ISP, we can fix the unmanageable one by extending which into base.

Therefore, by defining 3 restrictions, we can connect any two IDC from everywhere. Let's list all items together.
- All IDCs within same province and same ISP construct a connected graph.
- Existing an IDC on capital of down region can reach to an IDC on capital of up region with same ISP.
- Existing an IDC can reach to one of the base ISPs for any ISP within same province. Base is the minimal set of ISPs that our management machine built on.

But finding above items aren't enough, the challenge is how we confirm if a given routing path between any two IDCs is usable, optimal as well, not only in geography but also in mathematics. Also, it seems like that we are still not given the answer that why we must define such restrictions or classification. Let's take these questions into next laboratory, a simulated internet.

## Simulation

Real Internet is very complicated, and relative to LAN is much slower. We are going to run 10,000 HTTP servers with different latency and firewall configs to simulate a small internet on LAN.

For concurrently running 10 thousands of HTTP instances, we would rather use `golang` as the main language, however, other additional "pressure" simulations are setting up by [comcast](https://github.com/tylertreat/comcast), which also implied that the laboratory is based on Linux or Darwin. Let's get started.

### Basic HTTP Server

After [installing](https://golang.org/doc/install) the golang environment, goes into a working directory and makes a project, such as `simnet`, by `mkdir simnet && cd simnet`. We have `simnet.go` and `simnet_test.go` respectively shown as below.

```go
package simnet

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// ListenHTTP creates a PORT-unspecified HTTP server.
// If success, it returns the underlying port and a nil error.
func ListenHTTP() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	server := new(http.Server)
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	})
	go server.Serve(l)

	_, port, _ := net.SplitHostPort(l.Addr().String())
	n, _ := strconv.Atoi(port)
	return n, nil
}
```

```go
package simnet

import (
	"fmt"
	"net/http"
	"testing"
)

func TestListenHTTP(t *testing.T) {
	port, err := ListenHTTP()
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
}
```

Running `go test` should print `PASS`.

### Speed Measurement

Since the environment is built on LAN, so simply calculates speed values from downloading size divide by its elapse isn't enough, it still needs to operates the LAN QoS. Fortunately, thanks to [comcast](https://github.com/tylertreat/comcast) which provides a really easy way to do the stuff.

Firstly, updating the `simnet.go` so that which could serve a speed test.

```go
package simnet

import (
	"net"
	"net/http"
	"regexp"
	"strconv"
)

var (
	b8k          = make([]byte, 8192)
	speedPattern = regexp.MustCompile(`^([0-9]+)([kKmM])$`)
)

// ListenHTTP creates a PORT-unspecified HTTP server.
// If success it returns the underlying port and a nil error.
//
// Supported requests:
//   GET /[0-9]+[kKmM] - downloads a file of arbitrary size with random data
func ListenHTTP() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	server := new(http.Server)
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m := speedPattern.FindAllStringSubmatch(r.URL.Path[1:], -1); len(m) > 0 {
			size, _ := strconv.Atoi(m[0][1])
			switch m[0][2] {
			case "k", "K":
				size *= 1024
			case "m", "M":
				size *= 1024 * 1024
			}

			sent := 0
			for sent+len(b8k) < size {
				n, err := w.Write(b8k)
				if err != nil {
					panic(err)
				}
				sent += n
			}
			if sent < size {
				w.Write(b8k[:size-sent])
			}
		}
	})
	go server.Serve(l)

	_, port, _ := net.SplitHostPort(l.Addr().String())
	n, _ := strconv.Atoi(port)
	return n, nil
}
```

Then, downloading a file of `4k` size with bandwidth `8kbit/s`, the result of elapsed time shall approximately be `4s`. Please remember installing the `comcast` from terminal, by `go get github.com/tylertreat/comcast`.

```go
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
```

In the end, running `go test` should print `PASS`. Readers might notice that `comcast` flag `--default-bw=5` appeared as well, because of on Linux, the traffic control is implemented by command `tc`, which isn't as accurate as `pfctl` on Darwin. But it doesn't matter that how accurate `comcast` can do, the point is there has been a programmatic way that we can simulate the internet.

### Isolation

At previous section, we have created a HTTP server, however, there was no further information except a port to identify a host. Right here, we're gonna be citylizing a HTTP server by associating a port with a 4-digit [city code](https://raw.githubusercontent.com/modood/Administrative-divisions-of-China/master/dist/pc-code.json). Actually, there are 338 cities in total, it is easy to use a hash way or directly lookup a table.

Next, it is necessary to define an affinity between cities. Let's use 1-digit numbers as the affinity constant, 0 means a city is completely isolate to another, a greater number means a more open isolation, until 9 which means a city is thoroughly open to another. For simplicity, considering of a demonstrated function below.

```go
func CityAffinity(codeA, codeB string) int {
	// for same cities
	if codeA == codeA { return 9 }
	// for within a same province
	if codeA[:2] == codeB[:2] {	return 5 }
	// for within a same district
	if codeA[:1] == codeB[:1] { return 1 }
	// others
	return 0
}
```

In more ordinary scenes, affinity constant is defined between hosts, in which the isolation is not limited by city, but also, as we said, by ISP. So, a 3-D coordinate system can be given as below, A and B are a host within a specific city and ISP, where in plane y-z and x-z respectively, then the distance of segment AB is used as value of affinity between two hosts. Obviously, less distance stands for less isolation.

```
      | z (ISP)
      |
      |       .
      |      (A)
      |
    O /------------------- y (city)
 .   /
(B) /
   /
  / x (city)
```

After calculating a value of affinity between two hosts, we eventually need to map it into latency and apply which by `comcast`.

### Connectivity

TODO: check if this graph is connected
