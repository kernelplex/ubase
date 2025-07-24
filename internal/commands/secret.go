package commands

import (
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/kernelplex/ubase/lib/ubcli"
	"github.com/kernelplex/ubase/lib/ubsecurity"
)

func SecretCommand() ubcli.Command {
	flagset := flag.NewFlagSet("secret", flag.ExitOnError)
	var lengthFlag uint
	flagset.UintVar(&lengthFlag, "length", 32, "Length of the secret key")

	secretGenerate := func(args []string) error {
		flagset.Parse(args)
		bytes := ubsecurity.GenerateSecureRandom(uint32(lengthFlag))
		encoded := base64.StdEncoding.EncodeToString(bytes)
		fmt.Println(encoded)
		return nil
	}

	return ubcli.Command{
		Name:    "secret",
		Help:    "Generate a new secret key. This can be used to make unique secrets for PEPPER and SECRET_KEY settings.",
		Run:     secretGenerate,
		FlagSet: flagset,
	}
}
