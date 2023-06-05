package transactions

import (
	"encoding/json"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
)

type TransactionToGet struct {
	Id int64 `json:"transaction_id" db:"id"`
}

func InitGet(vals *url.Values) (*TransactionToGet, error) {
	var newT TransactionToGet
	for _, name := range []string{"transaction_id"} {
		v := vals.Get(name)
		if v == "" {
			return nil, errors.Errorf("need param %v", name)
		}
		param, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		// я знаю, что можно было рефлексией
		switch name {
		case "transaction_id":
			newT.Id = int64(param)
		}
	}
	return &newT, nil
}

type TransactionToPost struct {
	// id      int64     `db:"id"`
	Amount int64 `json:"amount" db:"amount"`
	FromId int64 `json:"from_id" db:"account_from_id"`
	ToId   int64 `json:"to_id" db:"account_to_id"`
	//DateUTC time.Time `json:"date_utc" db:"date"`
}

func InitPost(vals *url.Values) (*TransactionToPost, error) {
	var newT TransactionToPost
	for _, name := range []string{"amount", "from_id", "to_id"} {
		v := vals.Get(name)
		if v == "" {
			return nil, errors.Errorf("need param %v", name)
		}
		param, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		// я знаю, что можно было рефлексией
		switch name {
		case "amount":
			newT.Amount = int64(param)
		case "from_id":
			newT.FromId = int64(param)
		case "to_id":
			newT.ToId = int64(param)
		}
	}
	return &newT, nil
}

type TransactionMsg struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

const (
	CreateTransactionAction = "create_transaction"
	GetTransactionAction    = "get_transaction"
)

type TransactionDb struct {
	Id     int64 `db:"id"`
	FromId int64 `json:"from_id" db:"account_from_id"`
	ToId   int64 `json:"to_id" db:"account_to_id"`
	Amount int64 `json:"amount" db:"amount"`
	//DateUTC time.Time `json:"date_utc" db:"date"`
}

func RowToTransaction(row *pgx.Row) (*TransactionDb, error) {
	var t TransactionDb
	err := row.Scan(&t.Id, &t.Amount, &t.FromId, &t.ToId)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
