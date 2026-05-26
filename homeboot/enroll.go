package homeboot

import (
	"io/ioutil"
	"net/url"

	"shanhu.io/std/errcode"
)

func cmdEnroll(args []string) error {
	flags := cmdFlags.New()
	server := flags.String("server", defaultServer, "server to register")
	name := flags.String("name", "", "endpoint name")
	code := flags.String("code", "", "passcode")
	pubKey := flags.String("pubkey", "", "public key file")
	flags.ParseArgs(args)

	if *name == "" {
		return errcode.InvalidArgf("name not specified")
	}
	if *code == "" {
		return errcode.InvalidArgf("passcode not specified")
	}

	serverURL, err := url.Parse(*server)
	if err != nil {
		return errcode.Annotatef(err, "invalid server url: %q", *server)
	}
	if *pubKey == "" {
		return errcode.InvalidArgf("public key not specified")
	}
	pubKeyBytes, err := ioutil.ReadFile(*pubKey)
	if err != nil {
		return errcode.Annotate(err, "read public key")
	}

	return registerEndpoint(serverURL, *name, *code, pubKeyBytes)
}
