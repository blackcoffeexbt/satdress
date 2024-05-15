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
)

type Backend interface {
	HandlePayInvoice(context.Context, Nip47Request) (*Nip47Response, *Nip47Error)
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
			Preimage: "preimageplaceholder",
		},
	}

	return &response, nil
}

func (b *PhoenixBackend) HandleUnknownMethod(ctx context.Context, nip47req Nip47Request) (*Nip47Response, *Nip47Error) {

	return &Nip47Response{}, &Nip47Error{}
}
