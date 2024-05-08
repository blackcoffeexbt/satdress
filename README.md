# Satdress

Lightning Address Server

## How to run

1. Download the binary from the releases page (or compile with `go build` or `go get`)
2. Create a config file using the example file `config-sample.yml`.
3. Start the app with `./satdress`

## Features

- [x] [Lightning Address](https://github.com/andrerfneves/lightning-address#readme)
- [ ] [NIP-57](https://github.com/nostr-protocol/nips/blob/master/57.md) (Nostr Lightning Zaps)

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

