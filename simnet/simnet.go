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
