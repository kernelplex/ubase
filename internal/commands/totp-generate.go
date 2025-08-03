package commands

import (
	"flag"
	"fmt"

	"github.com/kernelplex/ubase/lib/ubapp"
	"github.com/kernelplex/ubase/lib/ubcli"
)

func TotpGenerateCommand() ubcli.Command {
	var (
		email string
	)

	flagset := flag.NewFlagSet("secret", flag.ExitOnError)
	flagset.StringVar(&email, "email", "", "Email address")

	totpGenerate := func(args []string) error {
		// agent := GetAgent()

		// Prompt for any missing required fields
		email = maybeReadInput("Email: ", email)

		app := ubapp.NewUbaseAppEnvConfig()
		defer app.Shutdown()
		totpService := app.GetTOTPService()
		totpUrl, err := totpService.GenerateTotp(email)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", totpUrl)

		return nil
	}

	return ubcli.Command{
		Name:    "totp-generate",
		Help:    "Generate a new secret key. This can be used to make unique secrets for PEPPER and SECRET_KEY settings.",
		Run:     totpGenerate,
		FlagSet: flagset,
	}
}
