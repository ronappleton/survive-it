package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Progress struct {
	DownloadedBytes int64
	TotalBytes      int64
}

type progressWriter struct {
	downloaded int64
	total      int64
	report     func(Progress)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.downloaded += int64(n)
	if w.report != nil {
		w.report(Progress{DownloadedBytes: w.downloaded, TotalBytes: w.total})
	}
	return n, nil
}

func ModelExists() (bool, error) {
	path, err := ModelPath()
	if err != nil {
		return false, err
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func DeleteModel() error {
	path, err := ModelPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func DownloadModel(ctx context.Context, onProgress func(Progress)) error {
	if strings.TrimSpace(ModelURL) == "" {
		return errors.New("model URL is empty")
	}

	modelPath, err := ModelPath()
	if err != nil {
		return err
	}
	modelsDir := filepath.Dir(modelPath)
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ModelURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	if onProgress != nil {
		onProgress(Progress{DownloadedBytes: 0, TotalBytes: resp.ContentLength})
	}

	tmp, err := os.CreateTemp(modelsDir, "ai-pack-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		_ = tmp.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	pw := &progressWriter{total: resp.ContentLength, report: onProgress}
	if _, err := io.Copy(io.MultiWriter(tmp, pw), resp.Body); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, modelPath); err != nil {
		return fmt.Errorf("swap model file: %w", err)
	}
	cleanup = false
	return nil
}
