package nwc

import (
	"context"
	"net/url"
	"net/http"
	"encoding/base64"
	"encoding/json"
	"io"
	"fmt"
	"strings"

	decodepay "github.com/nbd-wtf/ln-decodepay"
)

type PhoenixBalanceResult struct {
	Balance uint64 `json:"balanceSat"`
	FeeCredit uint64 `json:"feeCreditSat"`
}

type PhoenixLookupInvoiceResult struct {
	Invoice string `json:"invoice"`
	Description string `json:"description"`
	Fees uint64 `json:"fees"`
	CompleteAt uint `json:"completeAt"`
	CreatedAt uint `json:"createdAt"`
	Preimage string `json:"preimage"`
	PaymentHash string `json:"paymentHash"`
}

type PhoenixInvoiceResult struct {
	Amount uint64 `json:"amountSat"`
	PaymentHash string `json:"paymentHash"`
	Invoice string `json:"serialized"`
}

type PhoenixGetInfoResult struct {
	NodeId string `json:"nodeId"`
	// TODO channels
	Chain string `json:"chain"`
	Version string `json:"version"`
}

type Backend interface {
	HandlePayInvoice(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
	HandleGetBalance(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
	HandleMakeInvoice(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
	HandleLookupInvoice(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
	HandleGetInfo(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
	HandleUnknownMethod(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
}

type PhoenixBackend struct {
	Host string
	Key string
}

func (b *PhoenixBackend) payInvoice(invoice string) error {
	payload := url.Values{}
	payload.Set("invoice", invoice)

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		"http://"+b.Host+"/payinvoice",
		strings.NewReader(payload.Encode()),
	)

	if err != nil {
		return err
	}

	keyb64 := base64.StdEncoding.EncodeToString([]byte("phoenix-cli:"+b.Key))

	req.Header.Add("Authorization", "Basic "+keyb64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		text := string(body)
		if len(text) > 300 {
			text = text[:300]
		}
		return fmt.Errorf("call to phoenix failed (%d): %s", res.StatusCode, text)
	}

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

func (b *PhoenixBackend) getBalance() (*uint64, error) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"http://"+b.Host+"/getbalance",
		nil,
	)

	if err != nil {
		return nil, err
	}

	keyb64 := base64.StdEncoding.EncodeToString([]byte("phoenix-cli:"+b.Key))

	req.Header.Add("Authorization", "Basic "+keyb64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		text := string(body)
		if len(text) > 300 {
			text = text[:300]
		}
		return nil, fmt.Errorf("call to phoenix failed (%d): %s", res.StatusCode, text)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result = PhoenixBalanceResult{}

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		return nil, err
	}

	return &result.Balance, nil
}

func (b *PhoenixBackend) lookupInvoice(paymentHash string) (*PhoenixLookupInvoiceResult, error) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"http://"+b.Host+"/payinvoice/"+paymentHash,
		nil,
	)

	if err != nil {
		return nil, err
	}

	keyb64 := base64.StdEncoding.EncodeToString([]byte("phoenix-cli:"+b.Key))

	req.Header.Add("Authorization", "Basic "+keyb64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		text := string(body)
		if len(text) > 300 {
			text = text[:300]
		}
		return nil, fmt.Errorf("call to phoenix failed (%d): %s", res.StatusCode, text)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// TODO return nil, nil for not found

	var result = PhoenixLookupInvoiceResult{}

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *PhoenixBackend) makeInvoice(params Nip47InvoiceParams) (*PhoenixInvoiceResult, error) {
	payload := url.Values{}

	if params.DescriptionHash != "" {
		payload.Set("descriptionHash", params.DescriptionHash)
	} else {
		payload.Set("description", params.Description)
	}

	// TODO use params.Expiry

	payload.Add("amountSat", fmt.Sprintf("%d", params.Amount/1000))

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		"http://"+b.Host+"/createinvoice",
		strings.NewReader(payload.Encode()),
	)

	if err != nil {
		return nil, err
	}

	keyb64 := base64.StdEncoding.EncodeToString([]byte("phoenix-cli:"+b.Key))

	req.Header.Add("Authorization", "Basic "+keyb64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		text := string(body)
		if len(text) > 300 {
			text = text[:300]
		}
		return nil, fmt.Errorf("call to phoenix failed (%d): %s", res.StatusCode, text)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result = PhoenixInvoiceResult{}

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *PhoenixBackend) getInfo() (*PhoenixGetInfoResult, error) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"http://"+b.Host+"/getinfo",
		nil,
	)

	if err != nil {
		return nil, err
	}

	keyb64 := base64.StdEncoding.EncodeToString([]byte("phoenix-cli:"+b.Key))

	req.Header.Add("Authorization", "Basic "+keyb64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		text := string(body)
		if len(text) > 300 {
			text = text[:300]
		}
		return nil, fmt.Errorf("call to phoenix failed (%d): %s", res.StatusCode, text)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result = PhoenixGetInfoResult{}

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *PhoenixBackend) HandleGetBalance(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {

	balance, err := b.getBalance()

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not get balance",
		}
	}

	response := Nip47Response{
		ResultType: "get_balance",
		Result: Nip47GetBalanceResult{
			Balance: *balance,
		},
	}

	return &response, nil
}

func (b *PhoenixBackend) HandleLookupInvoice(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {


	var params Nip47LookupInvoiceParams

	err := json.Unmarshal(nip47req.Params, &params)

	var paymentHash string

	if params.PaymentHash == "" {
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(params.Invoice))
		if err != nil {
			return nil, &Nip47Error{
				Code: NIP47_ERROR_INTERNAL,
				Message: "could not decode invoice",
			}
		}
		paymentHash = paymentRequest.PaymentHash
	} else {
		paymentHash = params.PaymentHash
	}

	// TODO include a lookup for outgoing payments

	result, err := b.lookupInvoice(paymentHash)

	if result == nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_NOT_FOUND,
			Message: "could not find invoice",
		}
	}

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not load invoice",
		}
	}

	bolt11, err := decodepay.Decodepay(result.Invoice)

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not decode invoice",
		}
	}

	response := Nip47Response{
		ResultType: "lookup_invoice",
		Result: Nip47InvoiceResult{
			Type: "incoming",
			Invoice: result.Invoice,
			Description: result.Description,
			DescriptionHash: bolt11.DescriptionHash,
			Preimage: result.Preimage,
			PaymentHash: paymentHash,
			Amount: uint64(bolt11.MSatoshi),
			FeesPaid: result.Fees,
			CreatedAt: result.CreatedAt,
			ExpiresAt: uint(bolt11.Expiry),
			SettledAt: result.CompleteAt,
			// TODO metadata
		},
	}

	return &response, nil
}

func (b *PhoenixBackend) HandleMakeInvoice(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {

	var params Nip47InvoiceParams

	err := json.Unmarshal(nip47req.Params, &params)

	result, err := b.makeInvoice(params)

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not create invoice",
		}
	}

	bolt11, err := decodepay.Decodepay(result.Invoice)

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not decode invoice",
		}
	}

	response := &Nip47Response{
		ResultType: "make_invoice",
		Result: Nip47InvoiceResult{
			Type: "incoming",
			Invoice: result.Invoice,
			Description: bolt11.Description,
			DescriptionHash: bolt11.DescriptionHash,
			PaymentHash: bolt11.PaymentHash,
			Amount: uint64(bolt11.MSatoshi),
			CreatedAt: uint(bolt11.CreatedAt),
			ExpiresAt: uint(bolt11.Expiry),
		},
	}

	return response, nil
}

func (b *PhoenixBackend) HandleGetInfo(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {
	result, err := b.getInfo()

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not get information",
		}
	}

	response := &Nip47Response{
		ResultType: "get_info",
		Result: Nip47GetInfoResult{
			//Alias: nil,
			//Color: nil,
			//PubKey: nil,
			Network: result.Chain, // mainnet, testnet, signet, or regtest
			//BlockHeight: nil,
			//BlockHash: nil,
			Methods: strings.Split(NIP47_CAPABILITIES, " "),
		},
	}

	return response, nil
}

func (b *PhoenixBackend) HandlePayInvoice(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {

	var params Nip47PayParams

	err := json.Unmarshal(nip47req.Params, &params)

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not decode",
		}
	}

	err = b.payInvoice(params.Invoice)

	if err != nil {
		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "could not pay",
		}
	}

	response := Nip47Response{
		ResultType: "pay_invoice",
		Result: Nip47PayInvoiceResult{
			// TODO include preimage from phoenixd
			Preimage: "preimageplaceholder",
		},
	}

	return &response, nil
}

func (b *PhoenixBackend) HandleUnknownMethod(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {

	return &Nip47Response{}, &Nip47Error{}
}
