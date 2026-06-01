package graph

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	devwikipage "zatools/internal/devwiki/page"
	"zatools/internal/devwiki/retrieval"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultAPIUsername is the built-in username for the DevWiki HTTP API.
	DefaultAPIUsername = "devwiki"
	// DefaultAPIPassword is the built-in password for the DevWiki HTTP API.
	DefaultAPIPassword = "T19xwxc3n2I38F1A"
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
	return serveHTTP(ctx, host, opts.Port, graphHandler(opts))
}

// ServeAPI starts the read-only DevWiki HTTP API server.
func ServeAPI(ctx context.Context, opts ServerOptions) (string, error) {
	host := opts.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return serveHTTP(ctx, host, opts.Port, APIHandlerWithContext(ctx, opts.Root))
}

func serveHTTP(ctx context.Context, host string, port int, handler http.Handler) (string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return "", err
	}
	server := &http.Server{Handler: handler}
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
	actualPort := port
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		actualPort = tcpAddr.Port
	}
	url := "http://" + net.JoinHostPort(host, fmt.Sprintf("%d", actualPort)) + "/"
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

// APIHandler returns the read-only DevWiki HTTP API handler with Basic Auth.
func APIHandler(root string) http.Handler {
	return APIHandlerWithContext(context.Background(), root)
}

// APIHandlerWithContext returns the read-only DevWiki HTTP API handler with Basic Auth and shared command context.
func APIHandlerWithContext(ctx context.Context, root string) http.Handler {
	root = filepath.Clean(root)
	return requireAPIBasicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/devwiki/") {
			handleAPI(ctx, w, r, root)
			return
		}
		http.NotFound(w, r)
	}))
}

func requireAPIBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(DefaultAPIUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(password), []byte(DefaultAPIPassword)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="DevWiki"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type apiReadRequest struct {
	Kind string `json:"kind"`
	Slug string `json:"slug"`
	View string `json:"view"`
}

type apiSearchRequest struct {
	Kind  string   `json:"kind"`
	Query []string `json:"query"`
}

type apiTextResponse struct {
	Text string `json:"text"`
}

type apiProjectInfo struct {
	ProjectSlug string `json:"project_slug"`
	ProjectName string `json:"project_name"`
	Language    string `json:"language"`
}

func handleAPI(ctx context.Context, w http.ResponseWriter, r *http.Request, root string) {
	if r.URL.Path == "/api/devwiki/project" {
		handleAPIProject(w, r, root)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch r.URL.Path {
	case "/api/devwiki/read":
		handleAPIRead(w, r, root)
	case "/api/devwiki/search":
		handleAPISearch(ctx, w, r, root)
	default:
		http.NotFound(w, r)
	}
}

func handleAPIProject(w http.ResponseWriter, r *http.Request, root string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	info := apiProjectInfo{
		ProjectSlug: filepath.Base(root),
		ProjectName: filepath.Base(root),
	}
	data, err := os.ReadFile(filepath.Join(root, "config", "project.yaml"))
	if err == nil {
		var parsed struct {
			ProjectSlug string `yaml:"project_slug"`
			ProjectName string `yaml:"project_name"`
			Language    string `yaml:"language"`
		}
		if yaml.Unmarshal(data, &parsed) == nil {
			if strings.TrimSpace(parsed.ProjectSlug) != "" {
				info.ProjectSlug = strings.TrimSpace(parsed.ProjectSlug)
			}
			if strings.TrimSpace(parsed.ProjectName) != "" {
				info.ProjectName = strings.TrimSpace(parsed.ProjectName)
			}
			info.Language = strings.TrimSpace(parsed.Language)
		}
	}
	writeAPIJSON(w, info)
}

func handleAPIRead(w http.ResponseWriter, r *http.Request, root string) {
	var request apiReadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.Kind != devwikipage.KindTopic && request.Kind != devwikipage.KindWorkflow {
		http.Error(w, "unsupported devwiki read kind", http.StatusBadRequest)
		return
	}
	view := strings.TrimSpace(request.View)
	if view == "" {
		view = "card"
	}
	switch view {
	case "card", "core", "explain":
	default:
		http.Error(w, "unsupported devwiki read view", http.StatusBadRequest)
		return
	}

	text, err := retrieval.ReadText(root, request.Kind, request.Slug, view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeAPIJSON(w, apiTextResponse{Text: text})
}

func handleAPISearch(ctx context.Context, w http.ResponseWriter, r *http.Request, root string) {
	var request apiSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	queries := retrieval.NormalizeQueries(request.Query)
	if len(queries) == 0 {
		http.Error(w, "devwiki search query cannot be empty", http.StatusBadRequest)
		return
	}
	switch request.Kind {
	case "index":
		results, err := retrieval.SearchIndexTable(root, queries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeAPIJSON(w, results)
	case "glossary":
		results, err := retrieval.SearchGlossaryTable(root, queries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeAPIJSON(w, results)
	case devwikipage.KindTopic, devwikipage.KindWorkflow:
		results, err := retrieval.SearchPages(ctx, root, request.Kind, queries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeAPIJSON(w, results)
	default:
		http.Error(w, "unsupported devwiki search kind", http.StatusBadRequest)
	}
}

func writeAPIJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
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
