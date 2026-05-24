package jarvis

import (
	"fmt"
	"sort"
	"strings"

	"shanhu.io/g/errcode"
	"shanhu.io/g/httputil"
	"shanhu.io/g/nameutil"
	"shanhu.io/g/settings"
	"shanhu.io/homedrv/drv/homeapp"
)

func loadCustomSubs(s settings.Settings) (map[string]string, error) {
	customSubs := make(map[string]string)
	if err := s.Get(keyCustomSubs, &customSubs); err != nil {
		if errcode.IsNotFound(err) {
			// Just ignore and start with empty subs list.
			return customSubs, nil
		}
		return nil, err
	}
	return customSubs, nil
}

func cmdCustomSubs(args []string) error {
	flags := cmdFlags.New()
	sock := flags.String(
		"sock", "var/jarvis.sock", "jarvis unix domain socket",
	)
	add := flags.Bool("add", false, "adds a sub domain")
	remove := flags.Bool("remove", false, "removes a sub domain")
	cflags := newClientFlags(flags)
	args = flags.ParseArgs(args)
	list := !*add && !*remove

	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	subMap, err := loadCustomSubs(d.settings)
	if err != nil {
		return errcode.Annotate(err, "read custom subdomains")
	}

	if list {
		if len(args) != 0 {
			return errcode.InvalidArgf("list takes no command")
		}
		subs := []string{}
		for sub := range subMap {
			subs = append(subs, sub)
		}
		sort.Strings(subs)
		for _, sub := range subs {
			fmt.Printf("%s -> %s\n", sub, subMap[sub])
		}
		return nil
	}

	fullDomain := func(sub string) (string, error) {
		// See if user specified a full domain name or just the subdomain.
		idx := strings.Index(sub, ".")
		if idx > 0 {
			// User specified a full domain name. Nothing else to do.
			return sub, nil
		}
		if idx == 0 {
			return "", errcode.InvalidArgf("subdomain can not start with dot")
		}

		// User only specified the custom subdomain label.
		if err := nameutil.CheckLabel(sub); err != nil {
			return "", errcode.Annotate(err, "check subdomain")
		}

		mainDomain, err := settings.String(d.settings, homeapp.KeyMainDomain)
		if err != nil {
			return "", errcode.Annotate(err, "expand subdomain")
		}
		return sub + "." + mainDomain, nil
	}

	if *add {
		if len(args) != 2 {
			return errcode.InvalidArgf("add command takes 2 arguments")
		}
		domain, err := fullDomain(args[0])
		if err != nil {
			return errcode.Annotate(err, "get full domain name")
		}
		dest := args[1]

		if _, ok := subMap[domain]; ok {
			return errcode.InvalidArgf("subdomain %q already exist", domain)
		}
		subMap[domain] = dest
	} else if *remove {
		if len(args) != 1 {
			return errcode.InvalidArgf("remove command takes 1 argument")
		}
		domain, err := fullDomain(args[0])
		if err != nil {
			return errcode.Annotate(err, "get full domain name")
		}
		if _, ok := subMap[domain]; !ok {
			return errcode.InvalidArgf("subdomain %q not in sub list", domain)
		}
		delete(subMap, domain)
	}

	if err := d.settings.Set(keyCustomSubs, subMap); err != nil {
		return errcode.Annotate(err, "save custom subdomain map")
	}

	// Ping jarvis to recreate doorway so that hostmap will be updated.
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/admin/recreate-doorway", nil, nil)
}
