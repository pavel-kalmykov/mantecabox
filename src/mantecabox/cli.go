package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"mantecabox/cli"

	"github.com/alexflint/go-arg"
	"github.com/go-http-utils/headers"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/resty.v1"
)

func init() {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resty.SetHostURL("https://localhost:10443")
	resty.SetHeader(headers.ContentType, "application/json")
	resty.SetHeader(headers.Accept, "application/json")

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	resty.SetOutputDirectory(home + "/Mantecabox")
}

func main() {
	var args struct {
		Operation       string   `arg:"positional, required" help:"(signup|login|transfer|help)"`
		TransferActions []string `arg:"positional" help:"(list|((upload|download|remove) <files>...)"`
	}
	parser := arg.MustParse(&args)

	switch args.Operation {
	case "signup":
		err := cli.Signup(cli.ReadCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during signup: %v\n", err)
		}
		break
	case "login":
		err := cli.Login(cli.ReadCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during login: %v\n", err)
		}
		break
	case "transfer":
		err := cli.Transfer(args.TransferActions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during transfer: %v\n", err)
		}
		break
	case "help":
		parser.WriteHelp(os.Stdin)
		break
	default:
		parser.Fail(fmt.Sprintf(`Operation "%v" not recognized`, args.Operation))
		break
	}
}
