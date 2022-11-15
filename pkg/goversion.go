package pkg

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sort"
	"sync"

	"github.com/coreos/go-semver/semver"
	"github.com/x0f5c3/manic-go/pkg/downloader"
)

const (
	DLURL = "https://go.dev/dl/?mode=json&include=all"
	KIND  = "archive"
	OS    = runtime.GOOS
)

var GoVersions = func() Versions {
	resp, err := http.Get(DLURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var versions Versions
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		panic(err)
	}
	vers := versions.Parse()
	sort.Slice(*vers, func(i, j int) bool {
		return (*vers)[i].Parsed().LessThan(*(*vers)[j].Parsed())
	})
	return *vers
}()

var Latest = GoVersions[0]

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

func (f *File) Download() error {
	dl, err := downloader.New(f.URL(), f.Sha256, nil, &f.Size)
	if err != nil {
		return err
	}
	dled, err := dl.Download(10, 10, true)
	if err != nil {
		return err
	}
	return dled.Save(f.Filename)
}

type Versions []GoVersion

func (v *Versions) Parse() *Versions {
	resChan := make(chan *GoVersion)
	wg := &sync.WaitGroup{}
	for _, ver := range *v {
		wg.Add(1)
		go func(ver *GoVersion) {
			defer wg.Done()
			ver.Parsed()
			resChan <- ver
		}(&ver)
	}
	go func() {
		wg.Wait()
		close(resChan)
	}()
	var res Versions
	for ver := range resChan {
		res = append(res, *ver)
	}
	return &res
}

type GoVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
	parsed  *semver.Version
}

func (v *GoVersion) Parsed() *semver.Version {
	if v.parsed == nil {
		v.parsed = semver.New(v.Version)
	}
	return v.parsed
}

func (v *GoVersion) File() *File {
	for _, file := range v.Files {
		if file.Kind == KIND && file.Os == OS {
			return &file
		}
	}
	return nil
}
