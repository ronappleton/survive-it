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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Hard-wire your repo here
	defaultRepo = "ronappleton/survive-it"

	// GitHub API base
	githubAPI = "https://api.github.com"
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

	// Replace current executable
	if err := replaceSelf(newBinPath); err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated to %s (%s).", rel.TagName, assetName), nil
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

func fetchLatestRelease(ctx context.Context, repo string) (*githubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPI, repo)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
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
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download %s: %s: %s", url, resp.Status, strings.TrimSpace(string(b)))
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func downloadAndFindChecksum(ctx context.Context, checksumsURL, assetName string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download checksums: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
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

		out := filepath.Join(tmpDir, "survive-it.new")
		of, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(of, tr); err != nil {
			of.Close()
			return "", err
		}
		of.Close()
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

		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		out := filepath.Join(tmpDir, "survive-it.new.exe")
		of, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(of, rc); err != nil {
			of.Close()
			return "", err
		}
		of.Close()
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
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
