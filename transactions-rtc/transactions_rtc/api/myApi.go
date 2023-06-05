package api

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
	tr "transactions_rtc/transactions"
)

const (
	RMQurl = "amqp://guest:guest@rabbitmq:5672"
)

type MyApi struct{}

func NewMyApi() *MyApi {
	return &MyApi{}
}

const (
	ApiTransaction = "/transaction/"
)

// ServeHTTP godoc
//
//	@Summary		Returns transaction info, depending on request
//	@Description		if POST request, returns transaction, created by request
//				if GET, returns transaction that has the transaction_id mentioned
//	@Accept			if POST x-www-form-urlencoded
//	@Produce		json
//	@Param			transaction_id	path int	true	"used when GET"
//	@Param			amount			body int	true	"used when POST"
//	@Param			from_id			body int	true	"used when POST"
//	@Param			to_id			body int	true	"used when POST"
//	@Success		200	{object}	tr.TransactionDb
//	@Failure		405	{object}	error
//	@Failure		500	{object}	error
//	@Router			/transaction/{transaction_id} [get]
//	@Router			/transaction/ [post]
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data *tr.TransactionDb
	var err error
	switch r.URL.Path {
	case ApiTransaction:
		data, err = srv.handlerTransaction(r)
	default:
		err = &ApiError{http.StatusMethodNotAllowed, errors.New("unknown method")}
	}
	if err != nil {
		var ae *ApiError
		ae = handleError(err)
		ae.WriteResponse(w)
		return
	}
	isOk := ApiAnswer{http.StatusOK, data}
	isOk.WriteResponse(w)
}

// handlerTransaction godoc
//
//	@Summary		Returns transaction info, depending on request
//	@Description		if POST request, returns transaction, created by request
//				if GET, returns transaction that has the transaction_id mentioned
//	@Accept			if POST x-www-form-urlencoded
//	@Produce		{object} tr.TransactionDb //знаю, что неправильно
//	@Param			transaction_id	path int	true	"used when GET"
//	@Param			amount			body int	true	"used when POST"
//	@Param			from_id			body int	true	"used when POST"
//	@Param			to_id			body int	true	"used when POST"
//	@Success		200	{object}	tr.TransactionDb
//	@Failure		405	{object}	error
//	@Failure		500	{object}	error
//	@Router			/transaction/{transaction_id} [get]
//	@Router			/transaction/ [post]
func (srv *MyApi) handlerTransaction(r *http.Request) (*tr.TransactionDb, error) {
	var res tr.TransactionDb
	var err error

	vals, err := SetValues(r)
	if err != nil {
		return nil, err
	}
	switch r.Method {
	case http.MethodPost:
		jsonResp, err := setTrMsg(vals)
		if err != nil {
			return nil, err
		}
		var rr []byte
		rr, err = sendToDB(jsonResp)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(rr, &res)
		if err != nil {
			return nil, err
		}
	case http.MethodGet:
		jsonResp, err := setTrMsgGet(vals)
		if err != nil {
			return nil, err
		}
		var rr []byte
		rr, err = sendToDB(jsonResp)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(rr, &res)
		if err != nil {
			return nil, err
		}
	default:
		return nil, &ApiError{http.StatusMethodNotAllowed, errors.New("api invalid method")}
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func setTrMsg(vals *url.Values) ([]byte, error) {
	t, err := tr.InitPost(vals)
	if err != nil {
		return nil, &ApiError{http.StatusBadRequest, err}
	}
	jst, err := json.Marshal(t)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	tMsg := tr.TransactionMsg{Action: tr.CreateTransactionAction, Data: jst}

	jsonResp, err := json.Marshal(tMsg)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	return jsonResp, nil
}

func setTrMsgGet(vals *url.Values) ([]byte, error) {
	t, err := tr.InitGet(vals)
	if err != nil {
		return nil, &ApiError{http.StatusBadRequest, err}
	}
	jst, err := json.Marshal(t)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	tMsg := tr.TransactionMsg{Action: tr.GetTransactionAction, Data: jst}

	jsonResp, err := json.Marshal(tMsg)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	return jsonResp, nil
}

func sendToDB(data []byte) (res []byte, err error) {
	conn, err := amqp.Dial(RMQurl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrId := randomString(32)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          data,
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = d.Body
			break
		}
	}

	return
}

func SetValues(r *http.Request) (*url.Values, error) {
	var vals url.Values
	switch r.Method {
	case http.MethodGet:
		vals = r.URL.Query()
		return &vals, nil
	case http.MethodPost:
		var body []byte
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, &ApiError{http.StatusInternalServerError, err}
		}
		vals, err = url.ParseQuery(string(body))
		if err != nil {
			return nil, &ApiError{http.StatusInternalServerError, err}
		}
		return &vals, nil
	default:
		return nil, &ApiError{http.StatusMethodNotAllowed, errors.New("unknown method")}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
