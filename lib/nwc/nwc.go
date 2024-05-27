package nwc

import (
	"fmt"
	"context"
	_ "embed"
	"time"
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/rs/zerolog"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	NIP47_INFO_KIND                  = 13194
	NIP47_REQUEST_KIND               = 23194
	NIP47_RESPONSE_KIND              = 23195

	NIP47_PAY_INVOICE_METHOD         = "pay_invoice"
	NIP47_GET_BALANCE_METHOD         = "get_balance"
	NIP47_GET_INFO_METHOD            = "get_info"
	NIP47_MAKE_INVOICE_METHOD        = "make_invoice"
	NIP47_LOOKUP_INVOICE_METHOD      = "lookup_invoice"
	NIP47_LIST_TRANSACTIONS_METHOD   = "list_transactions"
	NIP47_PAY_KEYSEND_METHOD         = "pay_keysend"
	NIP47_MULTI_PAY_INVOICE_METHOD   = "multi_pay_invoice"
	NIP47_MULTI_PAY_KEYSEND_METHOD   = "multi_pay_keysend"
	NIP47_SIGN_MESSAGE_METHOD        = "sign_message"

	NIP47_ERROR_RATE_LIMITED         = "RATE_LIMITED"
	NIP47_ERROR_NOT_IMPLEMENTED      = "NOT_IMPLEMENTED"
	NIP47_ERROR_NOT_FOUND            = "NOT_FOUND"
	NIP47_ERROR_INSUFFICIENT_BALANCE = "INSUFFICIENT_BALANCE"
	NIP47_ERROR_QUOTA_EXCEEDED       = "QUOTA_EXCEEDED"
	NIP47_ERROR_RESTRICTED           = "RESTRICTED"
	NIP47_ERROR_UNAUTHORIZED         = "UNAUTHORIZED"
	NIP47_ERROR_INTERNAL             = "INTERNAL"
	NIP47_ERROR_OTHER                = "OTHER"

	NIP47_CAPABILITIES               = "pay_invoice get_balance make_invoice lookup_invoice get_info list_transactions"
	NIP47_NOTIFICATION_TYPES         = "payment_received" // payment_received, balance_updated, payment_sent, channel_opened, channel_closed
)

const (
	REQUEST_EVENT_STATUS_RECEIVED = "received"
	REQUEST_EVENT_STATUS_RUNNING = "running"
	REQUEST_EVENT_STATUS_DONE = "done"
	RESPONSE_EVENT_STATUS_CREATED = "created"
	RESPONSE_EVENT_STATUS_SENDING = "sending"
	RESPONSE_EVENT_STATUS_DONE = "done"
)

//go:embed db/init.sql
var db_init_sql string

type Nip47Request struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Nip47PayParams struct {
	Invoice string `json:"invoice"`
}

type Nip47LookupInvoiceParams struct {
	Invoice string `json:"invoice"`
	PaymentHash string `json:"payment_hash"`
}

type Nip47ListTransactionsParams struct {
	From uint `json:"from,omitempty"`
	Until uint `json:"until,omitempty"`
	Limit uint `json:"limit,omitempty"`
	Offset uint `json:"offset,omitempty"`
	Unpaid bool `json:"unpaid,omitempty"`
	Type string `json:"type,omitempty"`
}

type Nip47InvoiceParams struct {
	Amount uint64 `json:"amount"`
	Description string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	Expiry uint `json:"expiry"`
}

type Nip47Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Nip47Response struct {
	Error      *Nip47Error `json:"error,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	ResultType string      `json:"result_type"`
}

type Nip47PayInvoiceResult struct {
	Preimage string `json:"preimage"`
}

type Nip47GetBalanceResult struct {
	Balance uint64 `json:"balance"`
}

type Nip47ListTransactionsResult struct {
	Transactions []Nip47InvoiceResult `json:"transactions"`
}

type Nip47InvoiceResult struct {
	Type string `json:"type"`
	Invoice string `json:"invoice"`
	Description string `json:"description,omitempty"`
	DescriptionHash string `json:"description_hash,omitempty"`
	Preimage string `json:"preimage,omitempty"`
	PaymentHash string `json:"payment_hash"`
	Amount uint64 `json:"amount"`
	FeesPaid uint64 `json:"fees_paid,omitempty"`
	CreatedAt uint `json:"created_at"`
	ExpiresAt uint `json:"expires_at,omitempty"`
	SettledAt uint `json:"settled_at,omitempty"`
}

type Nip47GetInfoResult struct {
	Alias string `json:"alias,omitempty"`
	Color string `json:"color,omitempty"`
	PubKey string `json:"pubkey"`
	Network string `json:"network"`
	BlockHeight uint `json:"block_height,omitempty"`
	BlockHash string `json:"block_hash,omitempty"`
	Methods []string `json:"methods"`
}

type RequestEvent struct {
	ID        uint
	NostrId   string `validate:"required"`
	PubKey    string
	User      string
	Raw       string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

type ResponseEvent struct {
	ID              uint
	NostrId         string `validate:"required"`
	RequestNostrId  string `validate:"required"`
	PubKey          string
	User            string
	Raw             string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NWCUser struct {
	Name string
	NWCSecret string
	NWCPubKey string
	Relay string
	Kind string
	Key string
	Host string
}

type NWCParams struct {
	PrivateKey string
	PublicKey string
	Users []NWCUser
	Logger *zerolog.Logger
	DBPath string
}

func (r *RequestEvent) GetNip47Request(p *NWCParams, user *NWCUser) (*Nip47Request, error) {
	ss, err := nip04.ComputeSharedSecret(user.NWCPubKey, p.PrivateKey)

	if err != nil {
		return nil, err
	}

	var evt = nostr.Event{}

	err = json.Unmarshal([]byte(r.Raw), &evt)

	if err != nil {
		return nil, err
	}

	payload, err := nip04.Decrypt(evt.Content, ss)

	if err != nil {
		return nil, err
	}

	request := &Nip47Request{}

	err = json.Unmarshal([]byte(payload), request)

	if err != nil {
		return nil, err
	}

	return request, nil
}

func InitDB(db *gorm.DB, p *NWCParams) {
	if err := db.Exec(db_init_sql).Error; err != nil {
		p.Logger.Fatal().Err(err).Msg("could not init db")
	}
}

func CommitResponseEvent(db *gorm.DB, p *NWCParams, user *NWCUser, response *nostr.Event, requestNostrId string) (*ResponseEvent, error) {
	re := &ResponseEvent{
		NostrId: response.ID,
		Raw: response.String(),
		RequestNostrId: requestNostrId,
		PubKey: response.PubKey,
		User: user.Name,
		Status: RESPONSE_EVENT_STATUS_CREATED,
	}

	tx := db.Begin()
	tx.Table("response_events").Create(re)
	tx.Table("request_events").Where("nostr_id = ?", requestNostrId).Update("status", REQUEST_EVENT_STATUS_DONE)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		p.Logger.Warn().Err(err).Msg("unable to save response")
		return nil, err
	}

	return re, nil
}

func CreateNostrResponse(p *NWCParams, refPubKey string, refID string, content interface{}, tags nostr.Tags, ss []byte) (result *nostr.Event, err error) {
	payloadBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	if err != nil {
		return nil, err
	}

	allTags := nostr.Tags{[]string{"p", refPubKey}, []string{"e", refID}}
	allTags = append(allTags, tags...)

	resp := &nostr.Event{
		PubKey:    p.PublicKey,
		CreatedAt: nostr.Now(),
		Kind:      NIP47_RESPONSE_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = resp.Sign(p.PrivateKey)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func notImplemented() (*Nip47Response, *Nip47Error) {
	return nil, &Nip47Error{
		Code: NIP47_ERROR_NOT_IMPLEMENTED,
		Message: "Not implemented.",
	}
}

func ExecuteRequest(ctx context.Context, db *gorm.DB, p *NWCParams, user *NWCUser, request *RequestEvent) (*ResponseEvent, error) {
	var backend Backend

	switch user.Kind {
	case "phoenix":
		backend = &PhoenixBackend{
			Host: user.Host,
			Key: user.Key,
		}
	default:
		return nil, fmt.Errorf("unsupported backend: %s", user.Kind)
	}

	ss, err := nip04.ComputeSharedSecret(user.NWCPubKey, p.PrivateKey)

	if err != nil {
		return nil, err
	}

	nip47Request, err := request.GetNip47Request(p, user)

	if err != nil {
		return nil, err
	}

	if request.Status == REQUEST_EVENT_STATUS_RUNNING {
		// TODO handle this exception
	}

	if request.Status != REQUEST_EVENT_STATUS_RECEIVED {
		p.Logger.Warn().Str("status", request.Status).Msg("must have 'received' status, ignoring")
		return nil, nil
	}

	err = db.Table("request_events").Where("id = ?", request.ID).Update("status", REQUEST_EVENT_STATUS_RUNNING).Error

	if err != nil {
		return nil, err
	}

	var nip47Resp *Nip47Response
	var nip47Err *Nip47Error

	switch nip47Request.Method {
	case NIP47_PAY_INVOICE_METHOD:
		nip47Resp, nip47Err = backend.HandlePayInvoice(ctx, *nip47Request)
	case NIP47_MULTI_PAY_INVOICE_METHOD:
		nip47Resp, nip47Err = notImplemented()
	case NIP47_MULTI_PAY_KEYSEND_METHOD:
		nip47Resp, nip47Err = notImplemented()
	case NIP47_PAY_KEYSEND_METHOD:
		nip47Resp, nip47Err = notImplemented()
	case NIP47_GET_BALANCE_METHOD:
		nip47Resp, nip47Err = backend.HandleGetBalance(ctx, *nip47Request)
	case NIP47_MAKE_INVOICE_METHOD:
		nip47Resp, nip47Err = backend.HandleMakeInvoice(ctx, *nip47Request)
	case NIP47_LOOKUP_INVOICE_METHOD:
		nip47Resp, nip47Err = backend.HandleLookupInvoice(ctx, *nip47Request)
	case NIP47_LIST_TRANSACTIONS_METHOD:
		nip47Resp, nip47Err = backend.HandleListTransactions(ctx, *nip47Request)
	case NIP47_GET_INFO_METHOD:
		nip47Resp, nip47Err = backend.HandleGetInfo(ctx, *nip47Request)
	case NIP47_SIGN_MESSAGE_METHOD:
		nip47Resp, nip47Err = notImplemented()
	default:
		nip47Resp, nip47Err = backend.HandleUnknownMethod(ctx, *nip47Request)
	}

	var nostrResp *nostr.Event

	if nip47Err != nil {
		p.Logger.Warn().Str("method", nip47Request.Method).Str("code", nip47Err.Code).Str("message", nip47Err.Message).Msg("created nip47 error")
		nostrResp, err = CreateNostrResponse(p, request.PubKey, request.NostrId, Nip47Response{
			Error: nip47Err,
		}, nil, ss)
	} else {
		p.Logger.Info().Str("result_type", nip47Resp.ResultType).Msg("created nip47 response")
		nostrResp, err = CreateNostrResponse(p, request.PubKey, request.NostrId, nip47Resp, nil, ss)
	}

	if err != nil {
		p.Logger.Warn().Err(err).Msg("unable to create nostr response")
		return nil, err
	}

	return CommitResponseEvent(db, p, user, nostrResp, request.NostrId)
}

func HandleEvent(db *gorm.DB, p *NWCParams, user *NWCUser, event *nostr.Event) (*RequestEvent, *Nip47Error) {
	p.Logger.Info().Str("user", user.Name).Str("event_id", event.ID).Msg("handling event")

	requestEvent := RequestEvent{}

	findEventResult := db.Table("request_events").Where("nostr_id = ?", event.ID).Find(&requestEvent)

	if findEventResult.RowsAffected != 0 {
		p.Logger.Warn().Str("nostr_id", event.ID).Msg("event already processed")
		return nil, nil
	}

	if user.NWCPubKey != event.PubKey {
		p.Logger.Warn().Str("pubkey", user.NWCPubKey).Msg("ignoring event, does not match pubkey")
		return nil, &Nip47Error{
			Code: NIP47_ERROR_UNAUTHORIZED,
			Message: "The public key is not authorized",
		}
	}

	revent := RequestEvent {
		NostrId: event.ID,
		PubKey: event.PubKey,
		User: user.Name,
		Raw: event.String(),
		Status: REQUEST_EVENT_STATUS_RECEIVED,
	}

	if err := db.Table("request_events").Create(&revent).Error; err != nil {
		p.Logger.Warn().Err(err).Str("node_id", event.ID).Msg("could not save event")

		return nil, &Nip47Error{
			Code: NIP47_ERROR_INTERNAL,
			Message: "Internal error",
		}
	}

	p.Logger.Info().Str("user", user.Name).Str("event_id", event.ID).Msg("ended event")

	return &revent, nil
}

func GetNip47Info(ctx context.Context, p *NWCParams, relay *nostr.Relay) (*nostr.Event, error) {
	sctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	filters := []nostr.Filter{{
		Kinds:   []int{NIP47_INFO_KIND},
		Authors: []string{p.PublicKey},
		Limit:   1,
	}}

	sub, err := relay.Subscribe(sctx, filters)
	defer sub.Unsub()

	if err != nil {
		return nil, err
	}

	var event nostr.Event

	for ev := range sub.Events {
		event = *ev
	}

	return &event, nil
}

func PublishResponseEvent(ctx context.Context, p *NWCParams, db *gorm.DB, relay *nostr.Relay, resp *ResponseEvent) error {
	if resp.Status != RESPONSE_EVENT_STATUS_CREATED {
		p.Logger.Warn().Str("status", resp.Status).Msg("must have 'created' status, ignoring")
		return nil
	}

	var event = nostr.Event{}

	err := json.Unmarshal([]byte(resp.Raw), &event)
	if err != nil {
		return err
	}

	err = db.Table("response_events").Where("id = ?", resp.ID).Update("status", RESPONSE_EVENT_STATUS_SENDING).Error
	if err != nil {
		return err
	}

	if !relay.IsConnected() {
		interval := 3 * time.Second

		for {
			p.Logger.Warn().Str("relay_url", relay.URL).Msg("relay is disconnected, attempting to reconnect")

			r := nostr.NewRelay(context.Background(), relay.URL)

			p.Logger.Info().Str("relay_url", relay.URL).Msg("connecting...")

			err := r.Connect(ctx)

			if err != nil {
				p.Logger.Warn().Err(err).Int("interval", int(interval)).Msg("unable to connect")
			} else {
				p.Logger.Info().Str("relay_url", relay.URL).Msg("connected")
				*relay = *r
				break
			}

			time.Sleep(interval)
			interval = interval * 17 / 10
		}
	}

	err = relay.Publish(ctx, event)

	if err != nil {
		return err
	}

	err = db.Table("response_events").Where("id = ?", resp.ID).Update("status", RESPONSE_EVENT_STATUS_DONE).Error
	if err != nil {
		return err
	}

	return nil
}

func PublishNip47Info(ctx context.Context, p *NWCParams, relay *nostr.Relay) {
	ev := &nostr.Event{}
	ev.Kind = NIP47_INFO_KIND
	ev.Content = NIP47_CAPABILITIES
	ev.CreatedAt = nostr.Now()
	ev.PubKey = p.PublicKey
	ev.Tags = nostr.Tags{[]string{"notifications", NIP47_NOTIFICATION_TYPES}}
	err := ev.Sign(p.PrivateKey)

	if err != nil {
		p.Logger.Fatal().Err(err).Msg("could not sign")
		return
	}

	err = relay.Publish(ctx, *ev)

	if err != nil {
		p.Logger.Fatal().Err(err).Msg("nostr publish error")
		return
	}

	p.Logger.Info().Str("event_id", ev.ID).Msg("published info event")
}

func StartListener(ctx context.Context, db *gorm.DB, p *NWCParams, user NWCUser, pool *nostr.SimplePool, requests chan<- RequestEvent, responses chan<- ResponseEvent) {
	p.Logger.Info().Str("user", user.Name).Msg("start event worker")

	filters := []nostr.Filter{{
		Kinds:   []int{NIP47_REQUEST_KIND},
		Authors: []string{user.NWCPubKey},
		Limit:   1000,
	}}

	events := pool.SubMany(ctx, []string{user.Relay}, filters)

	var incoming nostr.IncomingEvent

	for {

		p.Logger.Info().Str("user", user.Name).Msg("waiting for events")

		incoming = <- events

		evt := incoming.Event

		if evt == nil {
			break
		}

		revent, nip47err := HandleEvent(db, p, &user, evt)

		if revent != nil && nip47err == nil {

			requests <- *revent

		} else if (nip47err != nil) {
			ss, err := nip04.ComputeSharedSecret(evt.PubKey, p.PrivateKey)

			if ss != nil {
				response, err := CreateNostrResponse(p, evt.PubKey, evt.ID, Nip47Response{
					Error: nip47err,
				}, nil, ss)

				if response != nil {
					rsp, _ := CommitResponseEvent(db, p, &user, response, evt.ID)
					if rsp != nil {
						responses <- *rsp
					}
				} else if err != nil {
					p.Logger.Warn().Err(err).Msg("unable to create nostr response")

				}
			} else if err != nil {
				p.Logger.Warn().Err(err).Msg("unable to compute shared secret")
			}
		}

		p.Logger.Info().Str("user", user.Name).Str("event_id", evt.ID).Msg("finished event")
	}
}

func StartExecuter(ctx context.Context, db *gorm.DB, p *NWCParams, user NWCUser, requests <-chan RequestEvent, responses chan<- ResponseEvent) {
	var request RequestEvent

	for {
		request = <-requests

		response, err := ExecuteRequest(ctx, db, p, &user, &request)

		if err != nil {
			p.Logger.Warn().Err(err).Msg("unable to execute")
		} else {
			responses <- *response
		}
	}
}

func StartPublisher(ctx context.Context, db *gorm.DB, p *NWCParams, relay *nostr.Relay, responses <-chan ResponseEvent) {
	var response ResponseEvent

	for {
		response = <-responses

		err := PublishResponseEvent(ctx, p, db, relay, &response)

		if err != nil {
			p.Logger.Warn().Err(err).Msg("unable to publish")
		}
	}
}

func Start(ctx context.Context, p *NWCParams) {
	p.Logger.Info().Str("dbpath", p.DBPath).Msg("using database file")

	db, err := gorm.Open(sqlite.Open(p.DBPath), &gorm.Config{})

	//defer db.Close()

	if err != nil {
		p.Logger.Fatal().Err(err).Msg("error loading database")
	}

	InitDB(db, p)

	pool := nostr.NewSimplePool(ctx)

	for _, user := range p.Users {

		if user.Relay == "" {
			continue
		}

		responses := make(chan ResponseEvent)
		requests := make(chan RequestEvent)

		relay, err := nostr.RelayConnect(ctx, user.Relay)

		if err != nil {
			p.Logger.Fatal().Err(err).Str("relay", user.Relay).Msg("could not connect")
		}

		info, err := GetNip47Info(ctx, p, relay)
		if err != nil {
			p.Logger.Warn().Err(err).Msg("could not get nwc info from relay")
		}

		if info != nil {
			if info.Content != NIP47_CAPABILITIES {
				PublishNip47Info(ctx, p, relay)
			} else {
				p.Logger.Info().Str("info", info.ID).Msg("received info from relay")
			}
		} else {
			PublishNip47Info(ctx, p, relay)
		}

		p.Logger.Info().Str("pubkey", user.NWCPubKey).Msg("filtering for requests from pubkey")

		// TODO query the database for unfinished requests and responses
		// and then send them to their channels to be completed.

		go StartListener(ctx, db, p, user, pool, requests, responses)

		go StartExecuter(ctx, db, p, user, requests, responses)

		go StartPublisher(ctx, db, p, relay, responses)
	}

	<-ctx.Done()
}
