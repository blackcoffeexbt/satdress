package main

import (
	"fmt"
	"os"
	"net/url"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/rs/zerolog"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/urfave/cli/v2"
	"github.com/knadh/koanf/v2"
	"github.com/mdp/qrterminal/v3"
)

type User struct {
	Name string `koanf:"name"`
	Kind string `koanf:"kind"`
	Host string `koanf:"host"`
	Key string `koanf:"key"`
	Pak string `koanf:"pak"`
	Waki string `koanf:"waki"`
	NodeId string `koanf:"nodeid"`
	Rune string `koanf:"rune"`
	NWCPubKey string `koanf:"nwcpubkey"`
	NWCSecret string `koanf:"nwcsecret"`
	NWCRelay string `koanf:"nwcrelay"`
}

type Settings struct {
	Users []User `koanf:"users"`
	NostrPrivateKey    string `koanf:"nostrprivatekey"`
}

var (
	s Settings
	k = koanf.New(".")
	log    = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Username lookup map.
	userMap = make(map[string]User)
)

func loadSettings(ctx *cli.Context) {
	filename := ctx.String("conf")

	if err := k.Load(file.Provider(filename), yaml.Parser()); err != nil {
		log.Fatal().Err(err).Msg("error loading file: %v")
	}


	k.Unmarshal("", &s)

	for _, user := range s.Users {
		userMap[user.Name] = user
	}
}

func keygen(ctx *cli.Context) error {
	privatekey := nostr.GeneratePrivateKey()
	publickey, err := nostr.GetPublicKey(privatekey)

	if err != nil {
		return err
	}

	nsec, err := nip19.EncodePrivateKey(privatekey)

	if err != nil {
		return err
	}

	npub, err := nip19.EncodePublicKey(publickey)

	if err != nil {
		return err
	}

	fmt.Printf("private key: %s\n", privatekey)
	fmt.Printf("public key: %s\n", publickey)
	fmt.Printf("nsec: %s\n", nsec)
	fmt.Printf("npub: %s\n", npub)

	return nil
}

func connectQRCode(ctx *cli.Context) error {
	loadSettings(ctx)

	username := ctx.String("user")

	user, ok := userMap[username]

	if ok {
		pubkey, err := nostr.GetPublicKey(s.NostrPrivateKey)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to get pubkey")
		}

		if user.NWCRelay == "" {
			log.Fatal().Err(err).Msg("missing relay")
		}

		if user.NWCSecret == "" {
			log.Fatal().Err(err).Msg("missing secret")
		}

		params := url.Values{}
		params.Add("relay", user.NWCRelay)
		params.Add("secret", user.NWCSecret)

		connect := "nostr+walletconnect://"+pubkey+"?"+params.Encode()

		qrterminal.Generate(connect, qrterminal.M, os.Stdout)
	} else {
		log.Fatal().Msg("no user")
	}

	return nil
}

func connectString(ctx *cli.Context) error {
	loadSettings(ctx)

	username := ctx.String("user")

	user, ok := userMap[username]

	if ok {
		pubkey, err := nostr.GetPublicKey(s.NostrPrivateKey)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to get pubkey")
		}

		if user.NWCRelay == "" {
			log.Fatal().Err(err).Msg("missing relay")
		}

		if user.NWCSecret == "" {
			log.Fatal().Err(err).Msg("missing secret")
		}

		params := url.Values{}
		params.Add("relay", user.NWCRelay)
		params.Add("secret", user.NWCSecret)

		fmt.Println("nostr+walletconnect://"+pubkey+"?"+params.Encode())
	} else {
		log.Fatal().Msg("no user")
	}

	return nil
}

func main() {
	app := &cli.App{
		Name: "satdress-cli",
		Usage: "A utility for satdress lightning address server.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "conf",
				Value: "config.yml",
				Usage: "the path to the config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "nwc",
				Usage:   "nostr wallet connect commands",
				Subcommands: []*cli.Command{
					{
						// TODO break keygen into
						// nostr-keygen
						// nostr-nwc-secret
						// use options to configure
						Name:  "keygen",
						Usage: "create a new nostr private key (32-byte)",
						Action: keygen,
					},
					{
						Name:  "connect-qrcode",
						Usage: "view nostr wallet connect qrcode",
						Action: connectQRCode,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "user",
								Value: "alice",
								Usage: "the username",
							},
						},
					},
					{
						Name:  "connect-string",
						Usage: "view nostr wallet connect string",
						Action: connectString,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "user",
								Value: "alice",
								Usage: "the username",
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("error running")
	}
}
