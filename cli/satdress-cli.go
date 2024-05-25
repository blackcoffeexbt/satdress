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

func viewNostrKeys(ctx *cli.Context) error {
	nsec := ctx.String("nsec")
	npub := ctx.String("npub")
	privatehex := ctx.String("private-hex")
	publichex := ctx.String("public-hex")

	var err error
	var hex any
	var prefix string

	if nsec != "" {
		prefix, hex, err = nip19.Decode(nsec)
	} else if npub != "" {
		prefix, hex, err = nip19.Decode(npub)
	}

	if err != nil {
		return err
	}

	if prefix == "npub" {
		fmt.Printf("public hex: %s\n", hex)

		return nil
	} else if prefix == "nsec" {
		fmt.Printf("private hex: %s\n", hex)

		public, err := nostr.GetPublicKey(hex.(string))

		if err != nil {
			return err
		}

		fmt.Printf("public hex: %s\n", public)

		pub, err := nip19.EncodePublicKey(public)

		if err != nil {
			return err
		}

		fmt.Printf("npub: %s\n", pub)

		return nil
	}

	if privatehex != "" {
		priv, err := nip19.EncodePrivateKey(privatehex)

		if err != nil {
			return err
		}

		fmt.Printf("nsec: %s\n", priv)

		public, err := nostr.GetPublicKey(privatehex)

		if err != nil {
			return err
		}

		fmt.Printf("public hex: %s\n", public)

		pub, err := nip19.EncodePublicKey(public)

		if err != nil {
			return err
		}

		fmt.Printf("npub: %s\n", pub)

		return nil
	}

	if publichex != "" {
		pub, err := nip19.EncodePublicKey(publichex)

		if err != nil {
			return err
		}

		fmt.Printf("npub: %s\n", pub)

		return nil
	}

	return fmt.Errorf("Must supply input.")
}

func createNostrKeys(ctx *cli.Context) error {
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

	fmt.Printf("nsec: %s\n", nsec)
	fmt.Printf("npub: %s\n", npub)
	fmt.Printf("private hex: %s\n", privatekey)
	fmt.Printf("public hex: %s\n", publickey)

	return nil
}

func createSecret(ctx *cli.Context) error {
	privatekey := nostr.GeneratePrivateKey()

	fmt.Printf("secret: %s\n", privatekey)

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
				Name:  "keygen",
				Usage: "create a new nostr private and public key (32-byte)",
				Action: createNostrKeys,
			},
			{
				Name: "keyencoding",
				Usage: "view a key with different encodings (e.g. nsec, npub, hex)",
				Action: viewNostrKeys,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "nsec",
						Usage: "the nsec key encoding value",
					},
					&cli.StringFlag{
						Name:  "npub",
						Usage: "the public key hex encoding value",
					},
					&cli.StringFlag{
						Name:  "private-hex",
						Usage: "the private key hex encoding value",
					},
					&cli.StringFlag{
						Name:  "public-hex",
						Usage: "the public key hex encoding value",
					},
				},
			},
			{
				Name:    "nwc",
				Usage:   "nostr wallet connect commands",
				Subcommands: []*cli.Command{
					{
						Name:  "create-secret",
						Usage: "create a new wallet connection secret (32-byte)",
						Action: createSecret,
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
