package doorway

import (
	"crypto/tls"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"shanhu.io/g/jsonx"
	"shanhu.io/g/osutil"
	"shanhu.io/std/errcode"
)

func readHostMap(p string) (map[string]string, error) {
	m := make(map[string]string)
	if err := jsonx.ReadFile(p, &m); err != nil {
		return nil, err
	}
	return m, nil
}

type manualCertEntry struct {
	Key   string // key file
	Certs string // certificate bundle
}

func readManualCerts(h *osutil.Home) (map[string]*tls.Certificate, error) {
	entries := make(map[string]*manualCertEntry)
	if err := jsonx.ReadFile(h.Etc("certs.jsonx"), &entries); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errcode.Annotate(err, "read certs.jsonx")
	}

	var domains []string
	for d := range entries {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	certs := make(map[string]*tls.Certificate)
	for _, d := range domains {
		entry := entries[d]
		cert, err := tls.LoadX509KeyPair(entry.Certs, entry.Key)
		if err != nil {
			return nil, errcode.Annotatef(err, "read cert for %q", d)
		}
		certs[d] = &cert
	}
	return certs, nil
}

func removeCertsBefore(dir string, t time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errcode.Annotatef(err, "read dir %q", dir)
	}

	var toRemove []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "acme_account+key" {
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return errcode.Annotatef(err, "get info of %q", entry)
		}
		if info.ModTime().Before(t) {
			toRemove = append(toRemove, entry.Name())
		}
	}
	sort.Strings(toRemove)

	for _, name := range toRemove {
		log.Printf("remove old cert %q", name)
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return errcode.Annotatef(err, "remove old cert %q", name)
		}
	}
	return nil
}

func serverConfigFromHome(h *osutil.Home) (*ServerConfig, error) {
	hostMap, err := readHostMap(h.Etc("host-map.jsonx"))
	if err != nil {
		return nil, errcode.Annotate(err, "read host map")
	}

	certCacheDir := h.Var("autocert")
	dirExists, err := osutil.IsDir(certCacheDir)
	if err != nil {
		return nil, errcode.Annotate(err, "check cert cache dir")
	}
	if !dirExists {
		if err := os.Mkdir(certCacheDir, 0700); err != nil {
			return nil, errcode.Annotate(err, "make cert cache dir")
		}
	}
	certCleanseCut := time.Date(2022, 1, 28, 0, 0, 0, 0, time.UTC)
	if err := removeCertsBefore(certCacheDir, certCleanseCut); err != nil {
		log.Print("error on removing old certs: ", err)
	}

	manualCerts, err := readManualCerts(h)
	if err != nil {
		return nil, errcode.Annotate(err, "load manual certs")
	}

	return &ServerConfig{
		HostMap:       hostMap,
		AutoCertCache: autocert.DirCache(certCacheDir),
		ManualCerts:   manualCerts,
	}, nil
}

func httpServerConfigFromHome(h *osutil.Home) (*HTTPServerConfig, error) {
	config := new(HTTPServerConfig)
	p := h.Etc("http.jsonx")
	if err := jsonx.ReadFile(p, config); err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	return config, nil
}

func fabricsConfigFromHome(h *osutil.Home) (*FabricsConfig, error) {
	c := new(FabricsConfig)
	p := h.Etc("fabrics.jsonx")
	if err := jsonx.ReadFile(p, c); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}

// ConfigFromHome reads Config from the given directories.
func ConfigFromHome(homeDir string) (*Config, error) {
	h, err := osutil.NewHome(homeDir)
	if err != nil {
		return nil, errcode.Annotate(err, "make home")
	}

	c := new(Config)

	serverConfig, err := serverConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "build server config")
	}
	c.Server = serverConfig

	httpConfig, err := httpServerConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read http server config")
	}
	c.HTTPServer = httpConfig

	fabConfig, err := fabricsConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics config")
	}

	if fabConfig.User != "" {
		c.Fabrics = fabConfig

		pemPath := h.Var("fabrics.pem")
		id, err := newFileIdentity(pemPath)
		if err != nil {
			return nil, errcode.Annotate(err, "read fabrics identity pem")
		}
		c.FabricsIdentity = id
	}

	return c, nil
}
