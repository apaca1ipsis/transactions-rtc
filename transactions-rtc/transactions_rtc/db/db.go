package main

import (
	"encoding/json"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"log"
	tr "transactions_rtc/transactions"
)

var Config = pgx.ConnConfig{
	Host: "pg", Port: 5432, Database: "transactions", User: "postgres", Password: "postgres"}

const RMQurl = "amqp://guest:guest@rabbitmq:5672"

type database struct {
	conn *pgx.Conn
}

func InitDatabase(c *pgx.Conn) *database {
	return &database{c}
}

func (d database) newTransaction(t *tr.TransactionToPost) (*tr.TransactionDb, error) {
	var inserted tr.TransactionDb

	if err := d.conn.QueryRow("INSERT INTO transactions"+
		" (amount, account_from_id, account_to_id) "+
		"VALUES ($1, $2, $3) "+
		"RETURNING id, amount, account_from_id, account_to_id",
		&t.Amount, &t.FromId, &t.ToId).Scan(&inserted.Id, &inserted.Amount, &inserted.FromId, &inserted.ToId); err != nil {
		return nil, err
	}
	return &inserted, nil
}

func (d database) getTransaction(t *tr.TransactionToGet) (*tr.TransactionDb, error) {
	row := d.conn.QueryRow("select * from transactions where id=$1", &t.Id)
	return tr.RowToTransaction(row)
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (db *database) handleMethod(data []byte) ([]byte, error) {
	var value tr.TransactionMsg
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}

	method := value.Action

	var tDb *tr.TransactionDb
	log.Printf(" [.] %v", method)
	switch method {
	case tr.CreateTransactionAction:
		var t tr.TransactionToPost
		err := json.Unmarshal(value.Data, &t)
		if err != nil {
			return nil, err
		}
		tDb, err = db.newTransaction(&t)
		if err != nil {
			return nil, err
		}
	case tr.GetTransactionAction:
		var t tr.TransactionToGet
		err := json.Unmarshal(value.Data, &t)
		if err != nil {
			return nil, err
		}
		tDb, err = db.getTransaction(&t)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown method")

	}
	var res []byte
	res, err = json.Marshal(tDb)
	if err != nil {
		return nil, err
	}
	return res, nil
}
