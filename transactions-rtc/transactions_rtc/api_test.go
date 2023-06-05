package main

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	myapi "transactions_rtc/api"
)

var (
	client = &http.Client{}
)

type Case struct {
	Method  string
	Path    string
	Query   string
	CheckDb bool
	Status  int
	Result  interface{}
}

// CaseResponse
type CR map[string]interface{}

func TestMyApi(t *testing.T) {
	ts := httptest.NewServer(myapi.NewMyApi())

	cases := []Case{
		Case{ // успешный запрос
			Path:    myapi.ApiTransaction,
			Method:  http.MethodPost,
			Query:   "amount=100&from_id=2&to_id=33",
			Status:  http.StatusOK,
			CheckDb: true,
			Result: CR{
				"error": "",
			},
		},
		Case{ // плохой запрос
			Path:   myapi.ApiTransaction,
			Method: http.MethodPost,
			Query:  "amount=100&to_id=33",
			Status: http.StatusBadRequest,
			Result: CR{
				"error": "need param from_id",
			},
		},
		Case{ // хороший запрос
			Path:   myapi.ApiTransaction,
			Method: http.MethodGet,
			Query:  "transaction_id=1",
			Status: http.StatusOK,
			Result: CR{
				"error": "",
				"response": CR{
					"Id":      1,
					"amount":  -100,
					"from_id": 1,
					"to_id":   2,
				},
			},
		},
		Case{ // плохой запрос
			Path:   myapi.ApiTransaction,
			Method: http.MethodGet,
			Query:  "account_from_id=1",
			Status: http.StatusBadRequest,
			Result: CR{
				"error": "need param transaction_id",
			},
		},
	}

	runTests(t, ts, cases)
}

func runTests(t *testing.T, ts *httptest.Server, cases []Case) {
	for idx, item := range cases {
		var (
			err      error
			result   CR
			expected CR
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		if item.Method == http.MethodPost {
			reqBody := strings.NewReader(item.Query)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != item.Status {
			t.Errorf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Errorf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		if item.CheckDb {
			expected = resToCr(CR{"error": "", "response": getLastInsertedCase()})
			result = resToCr(result)
		} else {
			data, _ := json.Marshal(item.Result)
			json.Unmarshal(data, &expected)
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("[%d] results not match\nGot: %#v\nExpected: %#v", idx, result, expected)
			continue
		} else {
			t.Logf("[%d] results good", idx)
		}
	}
}

var Config = pgx.ConnConfig{
	Host: "pg", Port: 5432, Database: "transactions", User: "postgres", Password: "postgres"}

func getLastInsertedCase() CR {
	// я знаю, что тесты к базе привязывать - это аморально
	dbc, err := pgx.Connect(Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbc.Close()
	row := dbc.QueryRow("select * from transactions order by id desc limit 1")
	var id, amount, fromid, toid int64
	err = row.Scan(&id, &amount, &fromid, &toid)
	if err != nil {
		log.Fatal(err)
	}
	return CR{
		"Id":      id,
		"amount":  amount,
		"from_id": fromid,
		"to_id":   toid,
	}
}

func resToCr(result CR) CR {
	r := result["response"]
	n := result["error"]
	var rr CR
	data, _ := json.Marshal(r)
	json.Unmarshal(data, &rr)
	return CR{"error": n, "response": rr}
}
