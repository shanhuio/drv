package jarvis

import (
	"fmt"
	"log"
	"time"

	"shanhu.io/g/errcode"
	"shanhu.io/g/jsonx"
	"shanhu.io/g/settings"
	"shanhu.io/drv/drvapi"
	"shanhu.io/drv/homeapp"
	"shanhu.io/drv/homeapp/nextcloud"
)

func endpointInitConfig(d *drive) (*drvapi.EndpointInitConfig, error) {
	if f := d.config.EndpointInitConfigFile; f != "" {
		config := new(drvapi.EndpointInitConfig)
		if err := jsonx.ReadFile(f, config); err != nil {
			return nil, errcode.Annotate(err, "read local config")
		}
		return config, nil
	}

	if !d.hasServer() {
		return &drvapi.EndpointInitConfig{Apps: []string{}}, nil
	}

	c, err := d.dialServer()
	if err != nil {
		return nil, errcode.Annotate(err, "dial server")
	}
	config := new(drvapi.EndpointInitConfig)
	if err := c.Call("/pubapi/endpoint/config", nil, config); err != nil {
		return nil, errcode.Annotate(err, "fetch remote config")
	}
	return config, nil
}

func initDone(d *drive) error {
	if !d.hasServer() {
		return nil
	}

	ncPass, err := settings.String(d.settings, nextcloud.KeyAdminPass)
	if err != nil {
		if errcode.IsNotFound(err) {
			ncPass = ""
		} else {
			return errcode.Annotate(err, "read nextcloud password")
		}
	}
	jarvisPass, err := settings.String(d.settings, keyJarvisPass)
	if err != nil {
		return errcode.Annotate(err, "read core password")
	}

	info := &drvapi.InitInfo{
		Time: time.Now().UnixNano(),

		JarvisPassword:    jarvisPass,
		NextcloudPassword: ncPass,
	}

	client, err := d.dialServer()
	if err != nil {
		return errcode.Annotate(err, "dial for init done")
	}
	const p = "/pubapi/endpoint/initdone"
	if err := client.Call(p, info, nil); err != nil {
		return errcode.Annotate(err, "report init done")
	}
	return nil
}

func install(d *drive, r *drvapi.Release) error {
	initConfig, err := endpointInitConfig(d)
	if err != nil {
		return errcode.Annotate(err, "read endpoint config")
	}

	// TODO(h8liu): fetch owner and owner's ssh keys and merge them?

	// Populate endpoint configs.
	domain := initConfig.MainDomain
	if domain == "" {
		domain = fmt.Sprintf("%s.homedrv.com", d.name)
	}
	if err := d.settings.Set(homeapp.KeyMainDomain, domain); err != nil {
		return errcode.Annotate(err, "save main domain")
	}
	if doms := initConfig.NextcloudDomains; len(doms) > 0 {
		if err := d.settings.Set(nextcloud.KeyDomains, doms); err != nil {
			return errcode.Annotate(err, "save nextcloud domains")
		}
	}
	if f := initConfig.FabricsServer; f != "" {
		if err := d.settings.Set(keyFabricsServerDomain, f); err != nil {
			return errcode.Annotate(err, "save fabrics server domain")
		}
	}

	d.appRegistry.setRelease(r)

	apps := initConfig.Apps
	if apps == nil {
		apps = []string{nextcloud.Name}
	}
	if err := d.apps.install(apps); err != nil {
		return errcode.Annotate(err, "install nextcloud suite")
	}

	log.Println("install doorway")

	doorwayConfig := &doorwayConfig{
		domain:        domain,
		fabricsServer: initConfig.FabricsServer,
	}
	doorway := newDoorway(d, doorwayConfig)
	if err := doorway.install(r.Doorway); err != nil {
		return errcode.Annotate(err, "install doorway")
	}
	doorway.pingDomains()

	if err := initDone(d); err != nil {
		return errcode.Annotate(err, "send back init info")
	}

	log.Printf("HomeDrive successfully installed at https://%s", domain)

	endpointURL := "https://www.homedrive.io/endpoint/" + d.name
	log.Printf("See password(s) at %s", endpointURL)

	return nil
}

func downloadAndInstall(d *drive) error {
	if d.config.Channel == "" {
		return errcode.InvalidArgf("install channel not specified")
	}

	if err := checkSystem(d); err != nil {
		return errcode.Annotate(err, "check system")
	}

	dl, err := downloader(d)
	if err != nil {
		return errcode.Annotate(err, "init downloader")
	}
	dlConfig := d.downloadConfig()
	dlConfig.LatestOnly = true
	release, err := dl.DownloadRelease(dlConfig)
	if err != nil {
		return errcode.Annotate(err, "download release")
	}

	if err := install(d, release); err != nil {
		return errcode.Annotate(err, "install failed")
	}

	if err := d.settings.Set(keyBuild, release); err != nil {
		return errcode.Annotate(err, "commit build")
	}
	return nil
}
