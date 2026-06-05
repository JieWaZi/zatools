package updateapp

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultOwner = "JieWaZi"
	defaultRepo  = "zatools"
	binaryBase   = "zatools"
)

// ServiceOptions configures the self-update service.
type ServiceOptions struct {
	// BaseURL overrides the GitHub release download base URL.
	BaseURL string
	// ExecutablePath overrides the current executable path.
	ExecutablePath string
	// GOOS overrides the target operating system.
	GOOS string
	// GOARCH overrides the target architecture.
	GOARCH string
	// HTTPClient overrides the HTTP client used for release downloads.
	HTTPClient *http.Client
	// RunDetached overrides detached process startup for deferred Windows replacement.
	RunDetached func(name string, args ...string) error
}

// Result describes a completed self-update attempt.
type Result struct {
	// Asset is the release archive selected for this platform.
	Asset string
	// Path is the executable path being updated.
	Path string
	// Deferred reports whether replacement will happen after this process exits.
	Deferred bool
}

// Service updates the currently running zatools executable from GitHub releases.
type Service struct {
	baseURL        string
	executablePath string
	goos           string
	goarch         string
	httpClient     *http.Client
	runDetached    func(name string, args ...string) error
}

// NewService constructs a self-update service.
func NewService(opts ServiceOptions) *Service {
	baseURL := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if baseURL == "" {
		owner := strings.TrimSpace(os.Getenv("ZATOOLS_OWNER"))
		if owner == "" {
			owner = defaultOwner
		}
		repo := strings.TrimSpace(os.Getenv("ZATOOLS_REPO"))
		if repo == "" {
			repo = defaultRepo
		}
		version := strings.TrimSpace(os.Getenv("VERSION"))
		if version != "" {
			baseURL = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s", owner, repo, version)
		} else {
			baseURL = fmt.Sprintf("https://github.com/%s/%s/releases/latest/download", owner, repo)
		}
	}

	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	runDetached := opts.RunDetached
	if runDetached == nil {
		runDetached = startDetached
	}

	return &Service{
		baseURL:        baseURL,
		executablePath: opts.ExecutablePath,
		goos:           goos,
		goarch:         goarch,
		httpClient:     client,
		runDetached:    runDetached,
	}
}

// Update downloads the latest matching release archive and replaces the current executable.
func (s *Service) Update(ctx context.Context) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	executablePath := s.executablePath
	if strings.TrimSpace(executablePath) == "" {
		path, err := os.Executable()
		if err != nil {
			return Result{}, fmt.Errorf("resolve current executable: %w", err)
		}
		executablePath = path
	}
	executablePath, err := filepath.Abs(executablePath)
	if err != nil {
		return Result{}, fmt.Errorf("resolve executable path: %w", err)
	}

	checksums, err := s.downloadText(ctx, s.baseURL+"/checksums.txt")
	if err != nil {
		return Result{}, err
	}
	asset, expectedHash, err := selectAsset(checksums, s.goos, s.goarch)
	if err != nil {
		return Result{}, err
	}

	tempDir, err := os.MkdirTemp("", "zatools-update-*")
	if err != nil {
		return Result{}, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tempDir)
		}
	}()

	archivePath := filepath.Join(tempDir, asset)
	if err := s.downloadFile(ctx, s.baseURL+"/"+asset, archivePath); err != nil {
		return Result{}, err
	}
	if err := verifySHA256(archivePath, expectedHash); err != nil {
		return Result{}, err
	}

	extractedPath, err := extractBinary(archivePath, tempDir, binaryName(s.goos))
	if err != nil {
		return Result{}, err
	}
	if s.goos == "windows" {
		if err := s.scheduleWindowsReplace(tempDir, extractedPath, executablePath); err != nil {
			return Result{}, err
		}
		cleanup = false
		return Result{Asset: asset, Path: executablePath, Deferred: true}, nil
	}

	if err := replaceExecutable(extractedPath, executablePath); err != nil {
		return Result{}, err
	}
	return Result{Asset: asset, Path: executablePath}, nil
}

func (s *Service) downloadText(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", url, err)
	}
	return string(data), nil
}

func (s *Service) downloadFile(ctx context.Context, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func selectAsset(checksums, goos, goarch string) (string, string, error) {
	suffix := fmt.Sprintf("_%s_%s.tar.gz", goos, goarch)
	var selectedAsset string
	var selectedHash string
	count := 0
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		candidate := strings.TrimPrefix(fields[1], "./")
		if strings.HasPrefix(candidate, binaryBase+"_") && strings.HasSuffix(candidate, suffix) {
			count++
			selectedAsset = candidate
			selectedHash = strings.ToLower(fields[0])
		}
	}
	switch count {
	case 0:
		return "", "", fmt.Errorf("unable to find a matching release asset in checksums.txt")
	case 1:
		return selectedAsset, selectedHash, nil
	default:
		return "", "", fmt.Errorf("multiple matching release assets found in checksums.txt")
	}
}

func verifySHA256(path, expectedHash string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash %s: %w", path, err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != strings.ToLower(expectedHash) {
		return fmt.Errorf("checksum verification failed for %s", filepath.Base(path))
	}
	return nil
}

func extractBinary(archivePath, targetDir, name string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("read gzip archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != name {
			continue
		}
		outPath := filepath.Join(targetDir, name)
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", fmt.Errorf("create extracted binary: %w", err)
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return "", fmt.Errorf("extract binary: %w", copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("close extracted binary: %w", closeErr)
		}
		return outPath, nil
	}
	return "", fmt.Errorf("archive did not contain %s", name)
}

func replaceExecutable(sourcePath, executablePath string) error {
	targetDir := filepath.Dir(executablePath)
	tempTarget, err := os.CreateTemp(targetDir, ".zatools-update-*")
	if err != nil {
		return fmt.Errorf("create replacement file: %w", err)
	}
	tempTargetPath := tempTarget.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempTargetPath)
		}
	}()

	source, err := os.Open(sourcePath)
	if err != nil {
		_ = tempTarget.Close()
		return fmt.Errorf("open extracted binary: %w", err)
	}
	if _, err := io.Copy(tempTarget, source); err != nil {
		_ = source.Close()
		_ = tempTarget.Close()
		return fmt.Errorf("write replacement file: %w", err)
	}
	if err := source.Close(); err != nil {
		_ = tempTarget.Close()
		return fmt.Errorf("close extracted binary: %w", err)
	}
	if err := tempTarget.Chmod(0o755); err != nil {
		_ = tempTarget.Close()
		return fmt.Errorf("chmod replacement file: %w", err)
	}
	if err := tempTarget.Close(); err != nil {
		return fmt.Errorf("close replacement file: %w", err)
	}
	if err := os.Rename(tempTargetPath, executablePath); err != nil {
		return fmt.Errorf("replace %s: %w", executablePath, err)
	}
	cleanup = false
	return nil
}

func (s *Service) scheduleWindowsReplace(tempDir, sourcePath, executablePath string) error {
	scriptPath := filepath.Join(tempDir, "replace-zatools.ps1")
	script := fmt.Sprintf(`$ErrorActionPreference = "SilentlyContinue"
$Source = %s
$Target = %s
$TempDir = %s
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Milliseconds 500
    Move-Item -Force -Path $Source -Destination $Target
    if ($?) {
        Remove-Item -Recurse -Force $TempDir
        exit 0
    }
}
exit 1
`, psQuote(sourcePath), psQuote(executablePath), psQuote(tempDir))
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return fmt.Errorf("write replacement script: %w", err)
	}
	return s.runDetached("cmd", "/C", "start", "", "/B", "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
}

func binaryName(goos string) string {
	if goos == "windows" {
		return binaryBase + ".exe"
	}
	return binaryBase
}

func psQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func startDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
