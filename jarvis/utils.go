package jarvis

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"shanhu.io/drv/strutil"
	"shanhu.io/g/httputil"
	"shanhu.io/std/jsonx"
	"shanhu.io/std/tarutil"
)

func addJSONXToTarStream(
	s *tarutil.Stream, f string, m *tarutil.Meta, obj any,
) error {
	bs, err := jsonx.Marshal(obj)
	if err != nil {
		return err
	}
	s.AddBytes(f, m, bs)
	return nil
}

func pingDomains(domains []string) {
	set := strutil.MakeSet(domains)
	list := strutil.SortedList(set)

	client := http.DefaultClient

	done := make(map[string]bool)
	for i := 0; i < 3; i++ {
		if len(done) == len(list) {
			break
		}
		for _, d := range list {
			if done[d] {
				continue
			}
			u := &url.URL{Scheme: "https", Host: d}
			code, err := httputil.GetCode(client, u.String())
			if err != nil {
				// TODO(h8liu): investigate why we have EOF error from get.
				if !strings.HasSuffix(err.Error(), ": EOF") {
					log.Printf("warning: ping %s got error: %s", u, err)
				}
			} else if code != http.StatusOK {
				log.Printf("warning: ping %q got status %d", u, code)
			} else {
				done[d] = true
			}
		}
		time.Sleep(time.Second)
	}
}
