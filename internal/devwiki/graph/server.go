package graph

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ServerOptions describes the static graph server.
type ServerOptions struct {
	Dir  string
	Root string
	Host string
	Port int
}

// Serve starts a static HTTP server for the graph directory.
func Serve(ctx context.Context, opts ServerOptions) (string, error) {
	host := opts.Host
	if host == "" {
		host = "127.0.0.1"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, opts.Port))
	if err != nil {
		return "", err
	}
	server := &http.Server{Handler: graphHandler(opts)}
	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	url := "http://" + listener.Addr().String() + "/"
	select {
	case err := <-errCh:
		return "", err
	default:
		return url, nil
	}
}

func graphHandler(opts ServerOptions) http.Handler {
	static := http.FileServer(http.Dir(opts.Dir))
	root := opts.Root
	if root == "" {
		root = opts.Dir
	}
	root = filepath.Clean(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if strings.HasPrefix(cleanPath, "wiki/") && strings.HasSuffix(cleanPath, ".md") {
			http.ServeFile(w, r, filepath.Join(root, filepath.FromSlash(cleanPath)))
			return
		}
		static.ServeHTTP(w, r)
	})
}

// OpenBrowser opens a URL with the OS default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
