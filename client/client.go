package megaclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/t3rm1n4l/go-mega"
	"io/ioutil"
	"path"
	"time"
)

const (
	PATH_WIDTH = 50
	SIZE_WIDTH = 5
)

type MegaClient struct {
	cfg  *Config
	mega *mega.Mega
}

type Config struct {
	BaseUrl         string
	Retries         int
	DownloadWorkers int
	UploadWorkers   int
	TimeOut         int
	User            string
	Password        string
	Recursive       bool
	Force			bool
}

type Path struct {
	prefix string
	path   []string
	size   int64
	t      int
	ts     time.Time
}

func (p *Path) SetPrefix(s string) {
	p.prefix = s
}

func (p Path) GetPath() string {
	x := path.Join(p.path...)
	if p.t == mega.FOLDER {
		x = x + "/"
	}

	x = p.prefix + x

	return x
}

func (p Path) String() string {
	return fmt.Sprintf("%-*s %-*d %s", PATH_WIDTH, p.GetPath(), SIZE_WIDTH, p.size, p.ts.Format(time.RFC3339))
}

const (
	ROOT  = "mega"
	TRASH = "trash"
)

var (
	EINVALID_CONFIG = errors.New("Invalid json config")
	EINVALID_PATH   = errors.New("Invalid mega path")
)

func (cfg *Config) Parse(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		return EINVALID_CONFIG
	}

	return nil
}

func NewMegaClient(conf *Config) *MegaClient {
	c := &MegaClient{
		cfg:  conf,
		mega: mega.New(),
	}

	if conf.BaseUrl != "" {
		c.mega.SetAPIUrl(conf.BaseUrl)
	}

	if conf.Retries != 0 {
		c.mega.SetRetries(conf.Retries)
	}

	if conf.DownloadWorkers != 0 {
		c.mega.SetDownloadWorkers(conf.DownloadWorkers)
	}

	if conf.UploadWorkers != 0 {
		c.mega.SetUploadWorkers(conf.UploadWorkers)
	}

	if conf.TimeOut != 0 {
		c.mega.SetTimeOut(time.Duration(conf.TimeOut) * time.Second)
	}

	return c
}

func (mc *MegaClient) Login() error {
	err := mc.mega.Login(mc.cfg.User, mc.cfg.Password)

	if err != nil {
		return err
	}

	return mc.mega.GetFileSystem()

}

func (mc *MegaClient) List(resource string) (*[]Path, error) {
	var root *mega.Node
	var paths []Path
	var err error

	root, pathsplit, err := getLookupParams(resource, mc.mega.FS)
	if err != nil {
		return nil, err
	}

	var nodes []*mega.Node
	if len(*pathsplit) > 0 {
		nodes, err = mc.mega.FS.PathLookup(root, *pathsplit)
	}

	if err == nil {
		l := len(nodes)

		switch {
		case len(*pathsplit) == 0:
			nodes = root.GetChildren()
		case l > 0:
			nodes = nodes[l-1:]
			if len(nodes) == 1 {
				nodes = nodes[0].GetChildren()
			}
		}

		for _, n := range nodes {
			for _, p := range getPaths(n, mc.cfg.Recursive) {
				p.SetPrefix(resource)
				paths = append(paths, p)
			}
		}
		return &paths, nil
	}

	return nil, err
}

func (mc *MegaClient) Delete(resource string) error {
	root, pathsplit, err := getLookupParams(resource, mc.mega.FS)
	if err != nil {
		return err
	}

	var nodes []*mega.Node
	if len(*pathsplit) > 0 {
		nodes, err = mc.mega.FS.PathLookup(root, *pathsplit)
	} else {
		err = EINVALID_PATH
	}

	if err != nil {
		return err
	}

	l := len(nodes)
	node := nodes[l-1]

	return mc.mega.Delete(node, mc.cfg.Force)
}

func (s *MegaClient) Move(srcpath, dstpath string) {
	return
}

func (s *MegaClient) Get(srcpath, dstpath string) {
	return
}

func (s *MegaClient) Put(srcpath, dstpath string) {
	return
}

func (s *MegaClient) Sync(srcpath, dstpath string) {
	return

}
