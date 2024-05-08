# Satdress

Lightning Address Server

## How to run

1. Download the binary from the releases page (or compile with `go build` or `go get`)
2. Create a config file using the example file `config-sample.yml`
3. Start the app with `./satdress`
4. Configure the daemon scripts (e.g. `scripts/lightning-address.service`)

## Features

- [x] [Lightning Address](https://github.com/andrerfneves/lightning-address#readme)
- [ ] [NIP-57](https://github.com/nostr-protocol/nips/blob/master/57.md) (Nostr Lightning Zaps)
- [ ] [NIP-47](https://github.com/nostr-protocol/nips/blob/master/47.md) (Nostr Wallet Connect)

## Backends

- [x] Phoenix ([phoenixd](https://github.com/ACINQ/phoenixd/))
- [x] Commando ([Core Lightning](https://github.com/ElementsProject/lightning))
- [x] Sparko
- [x] LND
- [x] LNBits
- [x] LNPay
- [x] Eclair

## Screenshots

<img align="left" src="assets/satdress-send.png" width="320"/>
<img align="left" src="assets/satdress-payment.png" width="320"/>

