package homeboot

import (
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strings"

	"shanhu.io/drv/drvapi"
	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/g/creds"
	"shanhu.io/g/flagutil"
	"shanhu.io/g/httputil"
	"shanhu.io/g/jsonx"
	"shanhu.io/g/rsautil"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
	"shanhu.io/std/tarutil"
)

// BootConfig is a JSON marshallable file that is saved on
// the file system, often as /opt/homedrv/boot.jsonx
// It specifies the flags used by homeinstall.
type BootConfig struct {
	Drive        *drvcfg.Config
	Code         string
	Download     bool `json:",omitempty"`
	LegacyNaming bool
}

func stableChannel() string {
	if runtime.GOARCH == "amd64" {
		return "stable"
	}
	// Architecture other than amd64 should use their corresponding
	// release channels.
	return "stable-" + runtime.GOARCH
}

func (c *BootConfig) declareFlags(flags *flagutil.FlagSet) {
	drv := c.Drive
	flags.StringVar(
		&drv.Server, "server", defaultServer, "server to register",
	)
	flags.StringVar(&drv.Name, "name", "", "endpoint name")
	flags.StringVar(&c.Code, "code", "", "registration one time passcode")
	flags.StringVar(
		&drv.Channel, "channel", stableChannel(),
		"release channel to subscribe",
	)
	flags.StringVar(
		&drv.DockerSock, "docker", "", "docker unix domain socket",
	)
	flags.BoolVar(&c.Download, "download", true, "download docker image")
	flags.BoolVar(
		&c.LegacyNaming, "legacy_naming", false,
		"uses legacy naming, when used, -network is ignored",
	)
	flags.IntVar(
		&drv.HTTPPort, "http_port", 0,
		"http port to bind, "+
			"0 means 80 when managing OS or not auto_avoid_port_binding, "+
			"-1 means no binding",
	)
	flags.IntVar(
		&drv.HTTPSPort, "https_port", 0,
		"https port to bind, "+
			"0 means 443 when managing OS or not auto_avoid_port_binding, "+
			"-1 means no binding",
	)
	flags.BoolVar(
		&drv.AutoAvoidPortBinding, "auto_avoid_port_binding", true,
		"avoid binding ports when the port is 0 and not managing the OS",
	)
}

func (c *BootConfig) fixLegacyNaming() {
	if c.LegacyNaming {
		c.Drive.Naming = nil
	}
}

func newBootConfig() *BootConfig {
	return &BootConfig{
		Drive: &drvcfg.Config{Naming: &drvcfg.Naming{}},
	}
}

type boot struct {
	*BootConfig
}

func newBoot(config *BootConfig) *boot {
	return &boot{BootConfig: config}
}

func (b *boot) downloadCore(
	dock *docker.Client, config *drvcfg.Config, pem []byte,
) (string, error) {
	drv := b.Drive
	user := "~" + drv.Name

	serverURL, err := url.Parse(drv.Server)
	if err != nil {
		return "", errcode.Annotate(err, "parse server URL")
	}
	ep := &creds.RobotEndpoint{Server: serverURL, User: user, Key: pem}
	c, err := ep.Dial()
	if err != nil {
		return "", err
	}

	d := NewOfficialDownloader(c, dock)
	rel, err := d.DownloadRelease(&DownloadConfig{
		Channel:  drv.Channel,
		CoreOnly: true,
		Naming:   config.Naming,
	})
	if err != nil {
		return "", errcode.Annotate(err, "load core image")
	}

	return rel.Jarvis, nil
}

func (b *boot) saveDriveConfig(
	files *tarutil.Stream, c *drvcfg.Config,
) error {
	bs, err := jsonx.Marshal(c)
	if err != nil {
		return errcode.Annotate(err, "marshal core config")
	}
	if err != nil {
		return errcode.Annotate(err, "generate core config")
	}
	files.AddBytes("config.jsonx", tarutil.ModeMeta(0644), bs)
	return nil
}

func registerEndpoint(server *url.URL, name, code string, pub []byte) error {
	client := &httputil.Client{Server: server}
	const p = "/pubapi/endpoint/register"
	req := &drvapi.RegisterRequest{
		Name:       name,
		PassCode:   code,
		ControlKey: strings.TrimSpace(string(pub)),
	}
	return client.Call(p, req, nil)
}

func (b *boot) run() error {
	drv := b.Drive
	client := docker.NewUnixClient(drv.DockerSock)

	serverURL, err := url.Parse(drv.Server)
	if err != nil {
		return errcode.Annotate(err, "invalid server url")
	}

	pri, pub, err := rsautil.GenerateKey(nil, 0)
	if err != nil {
		return errcode.Annotate(err, "generate identity")
	}

	files := tarutil.NewStream()
	files.AddBytes("jarvis.pem", tarutil.ModeMeta(0600), pri)
	files.AddBytes("jarvis.pub", tarutil.ModeMeta(0644), pub)

	if err := b.saveDriveConfig(files, drv); err != nil {
		return err
	}

	if err := registerEndpoint(
		serverURL, drv.Name, b.Code, pub,
	); err != nil {
		return errcode.Annotate(err, "register endpoint")
	}
	log.Println("endpoint registered")

	var image string
	if b.Download {
		dl, err := b.downloadCore(client, drv, pri)
		if err != nil {
			return errcode.Annotate(err, "download core docker")
		}
		image = dl
		log.Println("HomeDrive core downloaded")
	}

	hasSysDock := true
	if err := CheckSystemDock(); err != nil {
		if !errcode.IsNotFound(err) {
			return errcode.Annotate(err, "check system docker socket")
		}
		hasSysDock = false
	}

	config := &CoreConfig{
		Drive:       drv,
		Image:       image,
		Files:       files,
		BindSysDock: hasSysDock,
	}
	id, err := b.startCore(client, config)
	if err != nil {
		return err
	}

	log.Printf("HomeDrive core started: %s", id)

	core := drvcfg.Core(drv.Naming)
	logsCmd := fmt.Sprintf("docker logs --follow %s", core)
	log.Printf("To track the installation progress, run: \n  %s", logsCmd)

	return nil
}

func (b *boot) startCore(
	client *docker.Client, config *CoreConfig,
) (string, error) {
	network := drvcfg.Network(config.Drive.Naming)

	found, err := docker.HasNetwork(client, network)
	if err != nil {
		return "", errcode.Annotatef(err, "check network: %q", network)
	}
	if !found {
		log.Printf("creating network %q ...", network)
		if err := docker.CreateNetwork(client, network); err != nil {
			return "", errcode.Annotatef(
				err, "create network: %q", network,
			)
		}
	}
	return StartCore(client, config)
}
