package profilegen

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/appengine-ltd/survive-it/internal/game"
)

const defaultSource = "Mapzen Terrarium elevation tiles (AWS public), distilled to compact profile stats"

type Options struct {
	ID         string
	Name       string
	BBox       [4]float64 // minLon, minLat, maxLon, maxLat
	CellMeters int
	CacheRoot  string
	Source     string
	SampleMax  int
}

func ParseBBox(raw string) ([4]float64, error) {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) != 4 {
		return [4]float64{}, fmt.Errorf("bbox must be minLon,minLat,maxLon,maxLat")
	}
	var bbox [4]float64
	for i := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return [4]float64{}, fmt.Errorf("parse bbox value %d: %w", i, err)
		}
		bbox[i] = v
	}
	return normalizeBBox(bbox)
}

func GenerateProfile(ctx context.Context, opts Options) (game.GenProfile, error) {
	bbox, err := normalizeBBox(opts.BBox)
	if err != nil {
		return game.GenProfile{}, err
	}
	if strings.TrimSpace(opts.ID) == "" {
		return game.GenProfile{}, errors.New("id is required")
	}
	if opts.CellMeters <= 0 {
		opts.CellMeters = 100
	}
	if opts.SampleMax <= 0 {
		opts.SampleMax = 34
	}
	if opts.CacheRoot == "" {
		opts.CacheRoot = filepath.Join(".cache", "genprofile")
	}
	if strings.TrimSpace(opts.Source) == "" {
		opts.Source = defaultSource
	}
	if strings.TrimSpace(opts.Name) == "" {
		opts.Name = opts.ID
	}

	lonDistM, latDistM := bboxDistanceMeters(bbox)
	sampleW := clampInt(int(math.Round(lonDistM/float64(opts.CellMeters))), 14, opts.SampleMax)
	sampleH := clampInt(int(math.Round(latDistM/float64(opts.CellMeters))), 14, opts.SampleMax)

	gridKey := hashString(fmt.Sprintf("%.6f:%.6f:%.6f:%.6f:%d:%d", bbox[0], bbox[1], bbox[2], bbox[3], sampleW, sampleH))
	scenarioCacheDir := filepath.Join(opts.CacheRoot, gridKey)
	if err := os.MkdirAll(scenarioCacheDir, 0o755); err != nil {
		return game.GenProfile{}, fmt.Errorf("mkdir cache: %w", err)
	}
	tileCacheDir := filepath.Join(opts.CacheRoot, "tiles")
	if err := os.MkdirAll(tileCacheDir, 0o755); err != nil {
		return game.GenProfile{}, fmt.Errorf("mkdir tile cache: %w", err)
	}

	lats, lons := sampleLatLonGrid(bbox, sampleW, sampleH)
	sampleCachePath := filepath.Join(scenarioCacheDir, "elevation_samples.json")
	elevMeters, loaded := loadSampleCache(sampleCachePath, sampleW*sampleH)
	if !loaded {
		elevMeters, err = fetchElevations(ctx, lats, lons, tileCacheDir, opts.CellMeters)
		if err != nil {
			return game.GenProfile{}, err
		}
		_ = writeSampleCache(sampleCachePath, elevMeters)
	}
	if len(elevMeters) != sampleW*sampleH {
		return game.GenProfile{}, fmt.Errorf("elevation sample size mismatch: got %d want %d", len(elevMeters), sampleW*sampleH)
	}

	elevSorted := append([]float64(nil), elevMeters...)
	sort.Float64s(elevSorted)
	elevP10M := percentileSorted(elevSorted, 0.10)
	elevP50M := percentileSorted(elevSorted, 0.50)
	elevP90M := percentileSorted(elevSorted, 0.90)

	stepX := lonDistM / float64(maxInt(1, sampleW-1))
	stepY := latDistM / float64(maxInt(1, sampleH-1))
	slopes := computeSlopeDegrees(elevMeters, sampleW, sampleH, stepX, stepY)
	slopeP50 := percentileSlice(slopes, 0.50)
	slopeP90 := percentileSlice(slopes, 0.90)

	ruggedness := computeRuggedness(elevMeters)
	riverDensity := computeRiverDensityProxy(elevMeters, sampleW, sampleH)
	lakeCoverage := computeLakeCoverageProxy(elevMeters, sampleW, sampleH)

	profile := game.GenProfile{
		ID:           opts.ID,
		Name:         opts.Name,
		CellMeters:   opts.CellMeters,
		ElevP10:      roundFloat(metersToElevUnits(elevP10M), 3),
		ElevP50:      roundFloat(metersToElevUnits(elevP50M), 3),
		ElevP90:      roundFloat(metersToElevUnits(elevP90M), 3),
		SlopeP50:     roundFloat(slopeP50, 3),
		SlopeP90:     roundFloat(slopeP90, 3),
		Ruggedness:   roundFloat(ruggedness, 4),
		RiverDensity: roundFloat(riverDensity, 4),
		LakeCoverage: roundFloat(lakeCoverage, 4),
		Notes:        fmt.Sprintf("Derived from %dx%d elevation samples in bbox [%.4f, %.4f, %.4f, %.4f]", sampleW, sampleH, bbox[0], bbox[1], bbox[2], bbox[3]),
		Source:       opts.Source,
	}
	return profile, nil
}

func WriteProfile(path string, profile game.GenProfile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	blob, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')
	return os.WriteFile(path, blob, 0o644)
}

func normalizeBBox(bbox [4]float64) ([4]float64, error) {
	minLon, minLat := bbox[0], bbox[1]
	maxLon, maxLat := bbox[2], bbox[3]
	if minLon > maxLon {
		minLon, maxLon = maxLon, minLon
	}
	if minLat > maxLat {
		minLat, maxLat = maxLat, minLat
	}
	if minLon < -180 || maxLon > 180 || minLat < -90 || maxLat > 90 {
		return [4]float64{}, fmt.Errorf("bbox out of WGS84 range")
	}
	if math.Abs(maxLon-minLon) < 0.001 || math.Abs(maxLat-minLat) < 0.001 {
		return [4]float64{}, fmt.Errorf("bbox is too small")
	}
	return [4]float64{minLon, minLat, maxLon, maxLat}, nil
}

func bboxDistanceMeters(bbox [4]float64) (float64, float64) {
	midLatRad := ((bbox[1] + bbox[3]) * 0.5) * math.Pi / 180.0
	latDist := math.Abs(bbox[3]-bbox[1]) * 111_320.0
	lonDist := math.Abs(bbox[2]-bbox[0]) * 111_320.0 * math.Cos(midLatRad)
	return maxFloat(1, lonDist), maxFloat(1, latDist)
}

func sampleLatLonGrid(bbox [4]float64, width, height int) ([]float64, []float64) {
	total := width * height
	lats := make([]float64, 0, total)
	lons := make([]float64, 0, total)
	for y := 0; y < height; y++ {
		fy := 0.0
		if height > 1 {
			fy = float64(y) / float64(height-1)
		}
		lat := bbox[1] + (bbox[3]-bbox[1])*fy
		for x := 0; x < width; x++ {
			fx := 0.0
			if width > 1 {
				fx = float64(x) / float64(width-1)
			}
			lon := bbox[0] + (bbox[2]-bbox[0])*fx
			lats = append(lats, lat)
			lons = append(lons, lon)
		}
	}
	return lats, lons
}

func fetchElevations(ctx context.Context, lats, lons []float64, cacheDir string, cellMeters int) ([]float64, error) {
	if len(lats) != len(lons) {
		return nil, fmt.Errorf("lat/lon length mismatch")
	}
	zoom := chooseTerrariumZoom(lats, lons, cellMeters)
	client := &http.Client{Timeout: 20 * time.Second}
	cache := map[string]image.Image{}
	values := make([]float64, len(lats))
	for i := range lats {
		tileXf, tileYf := lonLatToTile(lons[i], lats[i], zoom)
		n := 1 << zoom
		tileX := wrapTileX(int(math.Floor(tileXf)), n)
		tileY := clampInt(int(math.Floor(tileYf)), 0, n-1)
		px := clampInt(int(math.Floor((tileXf-float64(tileX))*256.0)), 0, 255)
		py := clampInt(int(math.Floor((tileYf-float64(tileY))*256.0)), 0, 255)
		key := fmt.Sprintf("%d/%d/%d", zoom, tileX, tileY)
		img, ok := cache[key]
		if !ok {
			var err error
			img, err = loadTerrariumTile(ctx, client, cacheDir, zoom, tileX, tileY)
			if err != nil {
				return nil, err
			}
			cache[key] = img
		}
		values[i] = terrariumElevationAt(img, px, py)
	}
	return values, nil
}

func chooseTerrariumZoom(lats, lons []float64, cellMeters int) int {
	if len(lats) == 0 || len(lons) == 0 {
		return 8
	}
	minLat := lats[0]
	maxLat := lats[0]
	for _, lat := range lats {
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}
	midLatRad := ((minLat + maxLat) * 0.5) * math.Pi / 180.0
	targetRes := clampFloat(float64(cellMeters)*6.0, 260, 2600)
	for z := 9; z >= 5; z-- {
		res := 156543.03392 * math.Cos(midLatRad) / math.Pow(2, float64(z))
		if res <= targetRes {
			return z
		}
	}
	return 5
}

func lonLatToTile(lon, lat float64, zoom int) (float64, float64) {
	lat = clampFloat(lat, -85.0511, 85.0511)
	n := math.Pow(2, float64(zoom))
	x := (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y := (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return x, y
}

func loadTerrariumTile(ctx context.Context, client *http.Client, cacheDir string, z, x, y int) (image.Image, error) {
	path := filepath.Join(cacheDir, "terrarium", strconv.Itoa(z), strconv.Itoa(x), fmt.Sprintf("%d.png", y))
	if blob, err := os.ReadFile(path); err == nil {
		img, err := png.Decode(bytes.NewReader(blob))
		if err == nil {
			return img, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://s3.amazonaws.com/elevation-tiles-prod/terrarium/%d/%d/%d.png", z, x, y)
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "survive-it-genprofile/1.0")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
			continue
		}
		blob, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("terrarium tile %d/%d/%d status %d", z, x, y, resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * 350 * time.Millisecond)
			continue
		}
		if err := os.WriteFile(path, blob, 0o644); err != nil {
			return nil, err
		}
		img, err := png.Decode(bytes.NewReader(blob))
		if err != nil {
			return nil, err
		}
		return img, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to fetch terrarium tile %d/%d/%d", z, x, y)
	}
	return nil, lastErr
}

func terrariumElevationAt(img image.Image, px, py int) float64 {
	r, g, b, _ := img.At(px, py).RGBA()
	r8 := float64(r >> 8)
	g8 := float64(g >> 8)
	b8 := float64(b >> 8)
	return (r8*256.0 + g8 + b8/256.0) - 32768.0
}

func wrapTileX(x, n int) int {
	if n <= 0 {
		return x
	}
	x %= n
	if x < 0 {
		x += n
	}
	return x
}

func loadSampleCache(path string, want int) ([]float64, bool) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var values []float64
	if err := json.Unmarshal(blob, &values); err != nil {
		return nil, false
	}
	if len(values) != want {
		return nil, false
	}
	return values, true
}

func writeSampleCache(path string, values []float64) error {
	blob, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return os.WriteFile(path, blob, 0o644)
}

func computeSlopeDegrees(elev []float64, width, height int, stepX, stepY float64) []float64 {
	if width < 3 || height < 3 {
		return []float64{0}
	}
	if stepX <= 0 {
		stepX = 1
	}
	if stepY <= 0 {
		stepY = 1
	}
	slopes := make([]float64, 0, width*height)
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			idx := y*width + x
			dzdx := (elev[idx+1] - elev[idx-1]) / (2 * stepX)
			dzdy := (elev[idx+width] - elev[idx-width]) / (2 * stepY)
			g := math.Sqrt(dzdx*dzdx + dzdy*dzdy)
			slopes = append(slopes, math.Atan(g)*180/math.Pi)
		}
	}
	if len(slopes) == 0 {
		return []float64{0}
	}
	return slopes
}

func computeRuggedness(elev []float64) float64 {
	if len(elev) == 0 {
		return 0.5
	}
	sum := 0.0
	for _, v := range elev {
		sum += v
	}
	mean := sum / float64(len(elev))
	ss := 0.0
	for _, v := range elev {
		d := v - mean
		ss += d * d
	}
	stdDev := math.Sqrt(ss / float64(len(elev)))
	return clampFloat(stdDev/220.0, 0.08, 2.4)
}

func computeRiverDensityProxy(elev []float64, width, height int) float64 {
	if len(elev) == 0 {
		return 0.05
	}
	n := len(elev)
	flowTo := make([]int, n)
	accum := make([]int, n)
	for i := range flowTo {
		flowTo[i] = -1
		accum[i] = 1
	}
	neigh := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			bestIdx := -1
			bestElev := elev[idx]
			for _, off := range neigh {
				nx := x + off[0]
				ny := y + off[1]
				if nx < 0 || ny < 0 || nx >= width || ny >= height {
					continue
				}
				nIdx := ny*width + nx
				if elev[nIdx] < bestElev {
					bestElev = elev[nIdx]
					bestIdx = nIdx
				}
			}
			flowTo[idx] = bestIdx
		}
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return elev[order[i]] > elev[order[j]] })
	for _, idx := range order {
		next := flowTo[idx]
		if next >= 0 {
			accum[next] += accum[idx]
		}
	}
	vals := make([]float64, n)
	for i, v := range accum {
		vals[i] = float64(v)
	}
	threshold := percentileSlice(vals, 0.94)
	hits := 0
	for _, v := range vals {
		if v >= threshold {
			hits++
		}
	}
	return clampFloat(float64(hits)/float64(n), 0.01, 0.22)
}

func computeLakeCoverageProxy(elev []float64, width, height int) float64 {
	if width < 3 || height < 3 {
		return 0.02
	}
	neigh := [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	lowQ := percentileSlice(elev, 0.24)
	minima := 0
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			idx := y*width + x
			e := elev[idx]
			if e > lowQ {
				continue
			}
			isMin := true
			for _, off := range neigh {
				nIdx := (y+off[1])*width + (x + off[0])
				if elev[nIdx] < e {
					isMin = false
					break
				}
			}
			if isMin {
				minima++
			}
		}
	}
	total := float64(width * height)
	return clampFloat(float64(minima)/total*1.8, 0.003, 0.14)
}

func metersToElevUnits(m float64) float64 {
	return clampFloat(m/40.0, -90, 90)
}

func percentileSlice(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copyVals := append([]float64(nil), values...)
	sort.Float64s(copyVals)
	return percentileSorted(copyVals, q)
}

func percentileSorted(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if q <= 0 {
		return sorted[0]
	}
	if q >= 1 {
		return sorted[len(sorted)-1]
	}
	pos := q * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}
	frac := pos - float64(lo)
	return sorted[lo] + (sorted[hi]-sorted[lo])*frac
}

func hashString(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func roundFloat(v float64, precision int) float64 {
	if precision < 0 {
		return v
	}
	factor := math.Pow(10, float64(precision))
	return math.Round(v*factor) / factor
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
