package pkg

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/x0f5c3/zerolog/log"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
	"github.com/x0f5c3/go-manager/pkg/semver"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

const (
	dLURL = "https://go.dev/dl/?mode=json&include=all"
	iND   = "archive"
	oS    = runtime.GOOS
)

type DownloadSettings struct {
	OutDir string
	OSTriple
}

func (d *DownloadSettings) AddToFlags(cmd *cobra.Command, persistent bool) {
	var flags *pflag.FlagSet
	if persistent {
		flags = cmd.PersistentFlags()
	} else {
		flags = cmd.Flags()
	}
	flags.StringVarP(&d.OutDir, "out-dir", "o", d.OutDir, "output directory")
	flags.StringVarP(&d.OSTriple.Os, "os", "s", d.OSTriple.Os, "os")
	flags.StringVarP(&d.OSTriple.Arch, "arch", "a", d.OSTriple.Arch, "arch")
	flags.StringVarP(&d.OSTriple.Kind, "kind", "k", d.OSTriple.Kind, "kind")
}

func NewDownloadSettings(outDir string, osTriple ...OSTriple) *DownloadSettings {
	if len(osTriple) > 0 {
		return &DownloadSettings{OutDir: outDir, OSTriple: osTriple[0]}
	}
	return &DownloadSettings{OutDir: outDir, OSTriple: CurrentKind}
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
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(dLURL)
	req.Header.SetMethod("GET")
	req.Header.Set("User-Agent", "manic-go")
	err := fasthttp.Do(req, resp)
	if err != nil {
		return nil, err
	}
	var versions Versions
	if err := json.Unmarshal(resp.Body(), &versions); err != nil {
		return nil, err
	}
	versions = versions.OnlyStable()
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

func (v *Versions) Parse() (*Versions, error) {
	pb, err := pterm.DefaultProgressbar.WithTitle("Parsing versions").WithTotal(len(*v)).Start()
	if err != nil {
		return nil, err
	}
	defer func(pb *pterm.ProgressbarPrinter) {
		_, err := pb.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to stop progress bar")
		}
	}(pb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	resChan := make(chan *GoVersion, len(*v))
	var versions Versions
	for _, version := range *v {
		wg.Add(1)
		go func(version *GoVersion) {
			defer wg.Done()
			defer pb.Increment()
			parsed, err := version.Parse()
			if err != nil {
				cancel()
				return
			}
			resChan <- parsed
		}(version)
	}
	go func() {
		wg.Wait()
		cancel()
		close(resChan)
	}()
	for {
		select {
		case <-ctx.Done():
			return &versions, nil
		case version := <-resChan:
			versions = append(versions, version)
		}
	}
}

func (v *Versions) OnlyStable() Versions {
	var versions Versions
	for _, version := range *v {
		if version.Stable {
			versions = append(versions, version)
		}
	}
	return versions
}

func (v *Versions) Len() int {
	return len(*v)
}

func (v *Versions) Less(i, j int) bool {
	cmp := semver.Compare((*v)[j].Version, (*v)[i].Version)
	if cmp != 0 {
		return cmp < 0
	}
	return (*v)[j].Version < (*v)[i].Version
}

func (v *Versions) Swap(i, j int) {
	(*v)[i], (*v)[j] = (*v)[j], (*v)[i]
}

type GoVersion struct {
	Version string          `json:"version"`
	Stable  bool            `json:"stable"`
	Files   []File          `json:"files"`
	Parsed  *semver.Version `json:"-"`
}

func (v *GoVersion) Parse() (*GoVersion, error) {
	if v.Parsed != nil {
		return v, nil
	}
	parsed, err := semver.ParseFromGo(v.Version)
	if err != nil {
		return nil, err
	}
	v.Parsed = parsed
	return v, nil
}

type OSTriple struct {
	Kind string
	Os   string
	Arch string
}

func NewOSTriple(kind string, os string, arch string) OSTriple {
	return OSTriple{Kind: kind, Os: os, Arch: arch}
}

var CurrentKind = currentKindOSArch()

func currentKindOSArch() OSTriple {
	return NewOSTriple(iND, oS, runtime.GOARCH)
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
