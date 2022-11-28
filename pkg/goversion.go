package pkg

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/goccy/go-json"

	"github.com/Masterminds/semver/v3"
	"github.com/pterm/pterm"
	"github.com/valyala/fasthttp"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

const (
	dLURL = "https://go.dev/dl/?mode=json&include=all"
	iND   = "archive"
	oS    = runtime.GOOS
)

type DownloadSettings struct {
	OutDir string
	KindOSArch
}

func DownloadLatest(outdir ...*DownloadSettings) error {
	versions, err := GetVersions()
	if err != nil {
		return err
	}
	f := versions[0].File(outdir...)
	if f == nil {
		return fmt.Errorf("failed to find file for %s", versions[0].Version)
	}
	return f.Download(outdir...)
}

func GetVersions() (Versions, error) {
	var dst []byte
	statusCode, resp, err := fasthttp.Get(dst, dLURL)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get versions, statuscode: %d", statusCode)
	}
	versions := make(Versions, 0)
	if err := json.Unmarshal(resp, &versions); err != nil {
		return nil, err
	}
	versions = versions.Parse()
	return versions, nil
}

type File struct {
	Filename string
	Os       string
	Arch     string
	Sha256   string
	Size     int
	Kind     string
}

func (f *File) URL() string {
	return "https://go.dev/dl/" + f.Filename
}

func (f *File) Download(outDir ...*DownloadSettings) error {
	dl, err := downloader.New(f.URL(), f.Sha256, nil, &f.Size)
	if err != nil {
		return err
	}
	dled, err := dl.Download(10, 10, true)
	if err != nil {
		return err
	}
	outPath := func() string {
		if len(outDir) > 0 {
			return filepath.Join(outDir[0].OutDir, f.Filename)
		}
		return f.Filename
	}()
	return dled.Save(outPath)
}

type Versions []*GoVersion

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	iParsed := v[i].Parsed()
	jParsed := v[j].Parsed()
	if iParsed == nil || jParsed == nil {
		return false
	}
	return iParsed.LessThan(jParsed)
}

func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Versions) Parse() Versions {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan *GoVersion)
	wg := &sync.WaitGroup{}
	for _, ver := range v {
		wg.Add(1)
		go func(ver *GoVersion) {
			defer wg.Done()
			newVer, err := ver.Parse()
			if err != nil {
				pterm.Debug.Printfln("failed to parse %s version: %s", ver.Version, err)
				return
			}
			resChan <- newVer
		}(ver)
	}
	go func() {
		wg.Wait()
		cancel()
		close(resChan)
	}()
	res := make(Versions, 0, len(v))
	for {
		select {
		case <-ctx.Done():
			v = res
			sort.Sort(v)
			return v
		case r := <-resChan:
			res = append(res, r)
		}
	}
	// cnt := 0
	// for ver := range resChan {
	// 	res[cnt] = ver
	// 	cnt++
	// }
	// *v = res
	// sort.Sort(sort.Reverse(v))
	// return v
}

type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
	parsed  *semver.Version
}

func (v *GoVersion) Parsed() *semver.Version {
	return v.parsed
}

func (v *GoVersion) SetParsed(parsed *semver.Version) {
	v.parsed = parsed
}

func (v *GoVersion) Parse() (*GoVersion, error) {
	if v.parsed == nil {
		parsed, err := semver.NewVersion(v.Version)
		if err != nil {
			return nil, err
		}
		v.parsed = parsed
	}
	return v, nil
}

type KindOSArch struct {
	Kind string
	Os   string
	Arch string
}

var CurrentKind = currentKindOSArch()

func currentKindOSArch() KindOSArch {
	return KindOSArch{
		Kind: iND,
		Os:   oS,
		Arch: runtime.GOARCH,
	}
}

func (v *GoVersion) File(wanted ...*DownloadSettings) *File {
	var kind, os, arch string
	if len(wanted) > 0 {
		kind = wanted[0].Kind
		os = wanted[0].Os
		arch = wanted[0].Arch
	} else {
		kind = CurrentKind.Kind
		os = CurrentKind.Os
		arch = CurrentKind.Arch
	}
	for _, file := range v.Files {
		if file.Kind == kind && file.Os == os && file.Arch == arch {
			return &file
		}
	}
	return nil
}
