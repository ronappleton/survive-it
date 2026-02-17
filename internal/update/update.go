package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	// Hard-wire your repo here
	defaultRepo = "ronappleton/survive-it"

	// GitHub API base
	githubAPI = "https://api.github.com"

	maxArchiveBytes   = 200 << 20
	maxExtractedBytes = 200 << 20
	maxChecksumsBytes = 1 << 20
)

var (
	allowedAssetHosts = map[string]struct{}{
		"api.github.com":                        {},
		"github.com":                            {},
		"objects.githubusercontent.com":         {},
		"github-releases.githubusercontent.com": {},
	}
	repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
)

type CheckParams struct {
	CurrentVersion string
}

type UpdateParams struct {
	CurrentVersion string
	// If true, actually downloads + replaces executable. If false, just checks.
	Apply bool
}

func Check(p CheckParams) (string, error) {
	res, err := checkAndMaybeUpdate(UpdateParams{
		CurrentVersion: p.CurrentVersion,
		Apply:          false,
	})
	return res, err
}

func Apply(currentVersion string) (string, error) {
	return checkAndMaybeUpdate(UpdateParams{
		CurrentVersion: currentVersion,
		Apply:          true,
	})
}

func checkAndMaybeUpdate(p UpdateParams) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rel, err := fetchLatestRelease(ctx, defaultRepo)
	if err != nil {
		return "", err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(p.CurrentVersion, "v")

	// Very lightweight compare: if equal -> up to date.
	// (Good enough for alpha; you can swap to semver later.)
	if latest == current || current == "dev" || current == "" {
		if latest == current {
			return fmt.Sprintf("Up to date (v%s).", latest), nil
		}
		// dev builds: always say what the latest is
		return fmt.Sprintf("Latest release is v%s.", latest), nil
	}

	assetName := expectedArchiveName("survive-it", rel.TagName, runtime.GOOS, runtime.GOARCH)
	assetURL, err := findAssetURL(rel, assetName)
	if err != nil {
		return "", err
	}

	checksumsURL, err := findAssetURL(rel, "checksums.txt")
	if err != nil {
		return "", err
	}

	if !p.Apply {
		return fmt.Sprintf("Update available: v%s → v%s. Run update to install.", current, latest), nil
	}

	// Download checksums
	wantSHA, err := downloadAndFindChecksum(ctx, checksumsURL, assetName)
	if err != nil {
		return "", err
	}

	// Download archive
	tmpDir, err := os.MkdirTemp("", "survive-it-update-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(ctx, assetURL, archivePath); err != nil {
		return "", err
	}

	// Verify SHA
	gotSHA, err := sha256File(archivePath)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(gotSHA, wantSHA) {
		return "", fmt.Errorf("checksum mismatch for %s: got %s want %s", assetName, gotSHA, wantSHA)
	}

	// Extract binary
	newBinPath, err := extractBinary(tmpDir, archivePath, runtime.GOOS)
	if err != nil {
		return "", err
	}

	if err := replaceSelf(newBinPath); err != nil {
		return "", err
	}

	return "", Restart()
}

// ---- GitHub release API ----

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func safeHTTPClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Second}
}

func validateRepo(repo string) error {
	if !repoPattern.MatchString(repo) {
		return fmt.Errorf("invalid repository format: %q", repo)
	}
	return nil
}

func validateHTTPSURL(raw string, allowedHosts map[string]struct{}) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	if _, ok := allowedHosts[host]; !ok {
		return fmt.Errorf("unsupported URL host: %s", host)
	}
	return nil
}

func fetchLatestRelease(ctx context.Context, repo string) (*githubRelease, error) {
	if err := validateRepo(repo); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPI, repo)
	if err := validateHTTPSURL(url, map[string]struct{}{"api.github.com": {}}); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")

	// #nosec G704 -- URL is fixed to api.github.com and validated above.
	resp, err := safeHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github latest release: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	if rel.TagName == "" {
		return nil, errors.New("latest release has no tag_name")
	}

	for _, asset := range rel.Assets {
		if err := validateHTTPSURL(asset.BrowserDownloadURL, allowedAssetHosts); err != nil {
			return nil, fmt.Errorf("invalid asset URL for %s: %w", asset.Name, err)
		}
	}

	return &rel, nil
}

func findAssetURL(rel *githubRelease, name string) (string, error) {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("release asset not found: %s", name)
}

// ---- naming ----

func expectedArchiveName(project, tag, goos, goarch string) string {
	ver := strings.TrimPrefix(tag, "v")
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("%s_%s_%s_%s.%s", project, ver, goos, goarch, ext)
}

// ---- download + checksum ----

func downloadFile(ctx context.Context, url, dest string) error {
	if err := validateHTTPSURL(url, allowedAssetHosts); err != nil {
		return err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// #nosec G704 -- URL host and scheme are validated above.
	resp, err := safeHTTPClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download %s: %s: %s", url, resp.Status, strings.TrimSpace(string(b)))
	}

	// #nosec G304 -- destination path is controlled by updater internals.
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, io.LimitReader(resp.Body, maxArchiveBytes+1))
	if err != nil {
		return err
	}
	if n > maxArchiveBytes {
		return fmt.Errorf("download exceeded max size (%d bytes)", maxArchiveBytes)
	}
	return nil
}

func downloadAndFindChecksum(ctx context.Context, checksumsURL, assetName string) (string, error) {
	if err := validateHTTPSURL(checksumsURL, allowedAssetHosts); err != nil {
		return "", err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
	// #nosec G704 -- URL host and scheme are validated above.
	resp, err := safeHTTPClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download checksums: %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxChecksumsBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxChecksumsBytes {
		return "", fmt.Errorf("checksums file exceeded max size (%d bytes)", maxChecksumsBytes)
	}

	// checksums.txt format: "<sha256>  <filename>"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		sha := parts[0]
		name := parts[len(parts)-1]
		if name == assetName {
			return sha, nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s", assetName)
}

func sha256File(path string) (string, error) {
	// #nosec G304 -- path is generated by updater internals.
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ---- extract ----

func extractBinary(tmpDir, archivePath, goos string) (string, error) {
	if goos == "windows" {
		return extractFromZip(tmpDir, archivePath)
	}
	return extractFromTarGz(tmpDir, archivePath)
}

func extractFromTarGz(tmpDir, archivePath string) (string, error) {
	// #nosec G304 -- archive path is generated by updater internals.
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// In our GoReleaser config, binary is named "survive-it"
		base := filepath.Base(hdr.Name)
		if base != "survive-it" {
			continue
		}
		if hdr.Size < 0 || hdr.Size > maxExtractedBytes {
			return "", fmt.Errorf("archive binary size out of bounds: %d", hdr.Size)
		}

		out := filepath.Join(tmpDir, "survive-it.new")
		// #nosec G302,G304 -- binary must be executable; path is updater-controlled.
		of, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		written, err := io.Copy(of, io.LimitReader(tr, maxExtractedBytes+1))
		if err != nil {
			_ = of.Close()
			return "", err
		}
		if written > maxExtractedBytes {
			_ = of.Close()
			return "", fmt.Errorf("extracted binary exceeded max size (%d bytes)", maxExtractedBytes)
		}
		if err := of.Close(); err != nil {
			return "", err
		}
		return out, nil
	}

	return "", errors.New("binary not found in tar.gz")
}

func extractFromZip(tmpDir, archivePath string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		base := filepath.Base(f.Name)
		if base != "survive-it.exe" && base != "survive-it" {
			continue
		}
		if f.UncompressedSize64 > maxExtractedBytes {
			return "", fmt.Errorf("zip binary size out of bounds: %d", f.UncompressedSize64)
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		out := filepath.Join(tmpDir, "survive-it.new.exe")
		// #nosec G302,G304 -- binary must be executable; path is updater-controlled.
		of, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			_ = rc.Close()
			return "", err
		}
		written, err := io.Copy(of, io.LimitReader(rc, maxExtractedBytes+1))
		if err != nil {
			_ = of.Close()
			_ = rc.Close()
			return "", err
		}
		if written > maxExtractedBytes {
			_ = of.Close()
			_ = rc.Close()
			return "", fmt.Errorf("extracted binary exceeded max size (%d bytes)", maxExtractedBytes)
		}
		if err := of.Close(); err != nil {
			_ = rc.Close()
			return "", err
		}
		if err := rc.Close(); err != nil {
			return "", err
		}
		return out, nil
	}

	return "", errors.New("binary not found in zip")
}

// ---- replace ----

func replaceSelf(newBinPath string) error {
	current, err := os.Executable()
	if err != nil {
		return err
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		return err
	}

	dir := filepath.Dir(current)
	tmp := filepath.Join(dir, ".survive-it.tmp")

	// Copy new binary beside current
	if err := copyFile(newBinPath, tmp, 0o755); err != nil {
		return err
	}

	// Backup current
	backup := current + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(current, backup); err != nil {
		return fmt.Errorf("backup current: %w", err)
	}

	// Move tmp into place
	if err := os.Rename(tmp, current); err != nil {
		// rollback
		_ = os.Rename(backup, current)
		return fmt.Errorf("replace current: %w", err)
	}

	_ = os.Remove(backup)
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	// #nosec G304 -- source path is generated by updater internals.
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// #nosec G304 -- destination path is generated by updater internals.
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return nil
}

func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		// #nosec G204,G702 -- restarting current executable with inherited args.
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return err
		}
		os.Exit(0)
		return nil
	}

	// #nosec G204,G702 -- restarting current executable with inherited args.
	return syscall.Exec(exe, os.Args, os.Environ())
}
