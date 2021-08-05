package rest

import (
	"fmt"
	"strconv"

	"github.com/bitfinexcom/bitfinex-api-go/pkg/convert"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/notification"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/wallet"
)

// WalletService manages data flow for the Wallet API endpoint
type WalletService struct {
	requestFactory
	Synchronous
}

// Retrieves all of the wallets for the account
// see https://docs.bitfinex.com/reference#rest-auth-wallets for more info
func (s *WalletService) Wallet() (*wallet.Snapshot, error) {
	req, err := s.requestFactory.NewAuthenticatedRequest(common.PermissionRead, "wallets")
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	os, err := wallet.SnapshotFromRaw(raw)
	if err != nil {
		return nil, err
	}

	return os, nil
}

// Submits a request to transfer funds from one Bitfinex wallet to another
// see https://docs.bitfinex.com/reference#transfer-between-wallets for more info
func (ws *WalletService) Transfer(from, to, currency, currencyTo string, amount float64) (*notification.Notification, error) {
	body := map[string]interface{}{
		"from":        from,
		"to":          to,
		"currency":    currency,
		"currency_to": currencyTo,
		"amount":      strconv.FormatFloat(amount, 'f', -1, 64),
	}
	req, err := ws.requestFactory.NewAuthenticatedRequestWithData(common.PermissionWrite, "transfer", body)
	if err != nil {
		return nil, err
	}
	raw, err := ws.Request(req)
	if err != nil {
		return nil, err
	}
	return notification.FromRaw(raw)
}

func (ws *WalletService) depositAddress(wallet string, method string, renew int) (*notification.Notification, error) {
	body := map[string]interface{}{
		"wallet":   wallet,
		"method":   method,
		"op_renew": renew,
	}
	req, err := ws.requestFactory.NewAuthenticatedRequestWithData(common.PermissionWrite, "deposit/address", body)
	if err != nil {
		return nil, err
	}
	raw, err := ws.Request(req)
	if err != nil {
		return nil, err
	}
	return notification.FromRaw(raw)
}

// Retrieves the deposit address for the given Bitfinex wallet
// see https://docs.bitfinex.com/reference#deposit-address for more info
func (ws *WalletService) DepositAddress(wallet, method string) (*notification.Notification, error) {
	return ws.depositAddress(wallet, method, 0)
}

// Submits a request to create a new deposit address for the give Bitfinex wallet. Old addresses are still valid.
// See https://docs.bitfinex.com/reference#deposit-address for more info
func (ws *WalletService) CreateDepositAddress(wallet, method string) (*notification.Notification, error) {
	return ws.depositAddress(wallet, method, 1)
}

// Submits a request to withdraw funds from the given Bitfinex wallet to the given address
// See https://docs.bitfinex.com/reference#withdraw for more info
func (ws *WalletService) Withdraw(wallet, method string, amount float64, address string, paymentId *string) (*notification.Notification, error) {
	body := map[string]interface{}{
		"wallet":  wallet,
		"method":  method,
		"amount":  strconv.FormatFloat(amount, 'f', -1, 64),
		"address": address,
	}
	if paymentId != nil {
		body["payment_id"] = *paymentId
	}
	req, err := ws.requestFactory.NewAuthenticatedRequestWithData(common.PermissionWrite, "withdraw", body)
	if err != nil {
		return nil, err
	}
	raw, err := ws.Request(req)
	if err != nil {
		return nil, err
	}
	return notification.FromRaw(raw)
}

type Movement2 struct {
	ID                      int64
	Currency                string
	CurrencyName            string
	MtsStarted              int64
	MtsUpdated              int64
	Status                  string
	Amount                  float64
	Fees                    float64
	DestinationAddress      string
	TransactionID           string
	WithdrawTransactionNote string
}

func movement2FromRaw(raw []interface{}) (n []Movement2, err error) {
	result := []Movement2{}
	for _, item := range raw {
		v, ok := item.([]interface{})
		if !ok {
			return
		}
		if len(v) != 22 {
			return nil, fmt.Errorf("data slice too short for notification: %#v", raw)
		}
		n := Movement2{
			ID:                      convert.I64ValOrZero(v[0]),
			Currency:                convert.SValOrEmpty(v[1]),
			CurrencyName:            convert.SValOrEmpty(v[2]),
			MtsStarted:              convert.I64ValOrZero(v[5]),
			MtsUpdated:              convert.I64ValOrZero(v[6]),
			Status:                  convert.SValOrEmpty(v[9]),
			Amount:                  convert.F64ValOrZero(v[12]),
			Fees:                    convert.F64ValOrZero(v[13]),
			DestinationAddress:      convert.SValOrEmpty(v[16]),
			TransactionID:           convert.SValOrEmpty(v[20]),
			WithdrawTransactionNote: convert.SValOrEmpty(v[21]),
		}
		result = append(result, n)
	}
	return result, nil
}

func (ws *WalletService) Movements(start *int64, end *int64, max *int32) (n []Movement2, err error) {
	var maxLimit int32 = 1000

	payload := map[string]interface{}{}
	if start != nil {
		payload["start"] = *start
	}
	if end != nil {
		payload["end"] = *end
	}
	if max != nil {
		if *max > maxLimit {
			return nil, fmt.Errorf("Max request limit:%d, got: %d", maxLimit, max)
		}
		payload["limit"] = *max
	}
	req, err := ws.requestFactory.NewAuthenticatedRequestWithData(common.PermissionRead, "movements/hist", payload)
	if err != nil {
		return nil, err
	}
	raw, err := ws.Request(req)
	if err != nil {
		return nil, err
	}
	return movement2FromRaw(raw)
}
