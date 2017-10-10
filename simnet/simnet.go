package simnet

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"syscall"
)

var (
	b8k          = make([]byte, 8192)
	speedPattern = regexp.MustCompile(`^([0-9]+)([kKmM])$`)
)

// Handler implements the http.HandlerFunc interface.
//
// Supported requests:
//   GET /[0-9]+[kKmM] - downloads a file of arbitrary size with random data
func Handler(w http.ResponseWriter, r *http.Request) {
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

		return
	}
}

// ListenHTTP creates a PORT-unspecified HTTP server.
// If success it returns the underlying port and a nil error.
func ListenHTTP(ctx context.Context) (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	server := new(http.Server)
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r.WithContext(ctx))
	})
	go server.Serve(l)
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
		server.Close()
	}()

	_, port, _ := net.SplitHostPort(l.Addr().String())
	n, _ := strconv.Atoi(port)
	return n, nil
}

// CreateServers creates specific number of HTTP servers.
func CreateServers(ctx context.Context, n int) ([]int, error) {
	var ports []int

	localCtx, cancel := context.WithCancel(ctx)
	for i := 0; i < n; i++ {
		port, err := ListenHTTP(localCtx)
		if err != nil {
			cancel()
			return nil, err
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func getFileLimit() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, err
	}
	return rLimit.Cur, nil
}

// func setFileLimit(n uint64) error {
// 	var rLimit syscall.Rlimit
// 	rLimit.Max = n
// 	rLimit.Cur = n
// 	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
// 	if err != nil {
// 		return err
// 	}
// 	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
// 	if err != nil {
// 		return err
// 	}
// 	if rLimit.Cur < n {
// 		return fmt.Errorf("the file limit is up to %v", rLimit.Cur)
// 	}
// 	return nil
// }
