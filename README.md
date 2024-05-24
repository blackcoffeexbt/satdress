# Satdress

Lightning Address Server

## How to run

Before you begin you'll want to setup a Lighting backend, for example [phoenixd](https://github.com/ACINQ/phoenixd/).

1. Compile `satdress` and `satdress-cli` by running `make`. This will run `go build` and `go build ./cli/satdress-cli.go`.
2. Copy and edit the sample config at `config-sample.yml`, you create keys and view other settings with `satdress-cli`.
3. Start the app with `./satdress --conf <path/to/config.yml>`.
4. Copy `satdress` to the system (e.g. `/var/local/bin/satdress`).
5. Configure, install and enable the daemon scripts (e.g. `scripts/lightning-address.service` to `/etc/systemd/system/lightning-address.service`).

## Features

- [x] [Lightning Address](https://github.com/andrerfneves/lightning-address#readme)
- [x] [NIP-57](https://github.com/nostr-protocol/nips/blob/master/57.md) (Nostr Lightning Zaps)
- [x] [NIP-47](https://github.com/nostr-protocol/nips/blob/master/47.md) (Nostr Wallet Connect)

## Backends

Full support:
- [x] Phoenix ([phoenixd](https://github.com/ACINQ/phoenixd/))

Limited support:
- [x] Commando ([Core Lightning](https://github.com/ElementsProject/lightning))
- [x] Sparko
- [x] LND
- [x] LNBits
- [x] LNPay
- [x] Eclair

## Docker Build

Build the image:
```bash
docker build -t satdress .
```

Notes:
- The datadir is a docker volume and is mounted at `/var/lib/satdress`.
- The configuration file is a [docker secret](https://docs.docker.com/engine/swarm/secrets/) and is expected to be mounted at `/run/secrets/satdress.yml`.

Create a secret for the service:
```bash
docker secret create satdress.yml ./config.yml
```

Create and start the service using the secret config:
```bash
docker service create --name satdress --secret satdress.yml satdress:latest
```

## Screenshots

<img align="left" src="assets/satdress-send.png" width="320"/>
<img align="left" src="assets/satdress-payment.png" width="320"/>

