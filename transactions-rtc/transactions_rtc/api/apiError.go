package api

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type ApiError struct {
	HTTPStatus int
	Err        error
}

type ApiAnswer struct {
	HTTPStatus int
	msg        any
}

func (ae ApiError) Error() string {
	return ae.Err.Error()
}

func (answ *ApiAnswer) toJson() ([]byte, error) {
	resp := make(map[string]any)
	resp["response"] = answ.msg
	resp["error"] = ""
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	return jsonResp, nil
}

func (answ ApiAnswer) WriteResponse(w http.ResponseWriter) {
	msg, err := answ.toJson()
	if err != nil {
		txt := fmt.Sprintf("Error happened in http.ResponseWriter write. Err: %s", err)
		errInner := &ApiError{http.StatusInternalServerError, errors.New(txt)}
		errInner.WriteResponse(w)
		log.Fatalf(txt)
	} else {
		w.WriteHeader(answ.HTTPStatus)
		w.Header().Set("Content-Type", "application/json")
	}
	w.Write(msg)
}

func (answ ApiError) toJson() ([]byte, error) {
	resp := make(map[string]string)
	resp["error"] = answ.Error()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.Errorf("Error happened in JSON marshal. Err: %s", err)
	}
	return jsonResp, nil
}

func (answ ApiError) WriteResponse(w http.ResponseWriter) {
	msg, err := answ.toJson()
	if err != nil {
		txt := fmt.Sprintf("Error happened in http.ResponseWriter write. Err: %s", err)
		errInner := &ApiError{http.StatusInternalServerError, errors.New(txt)}
		errInner.WriteResponse(w)
		log.Fatalf(txt)
	} else {
		w.WriteHeader(answ.HTTPStatus)
		w.Header().Set("Content-Type", "application/json")
	}
	w.Write(msg)
}

func handleError(err error) *ApiError {
	switch err.(type) {
	case ApiError:
		ae, ok := err.(ApiError)
		if !ok {
			innerErr := ApiError{http.StatusInternalServerError,
				errors.Errorf("error converting %v to ApiError", err)}
			return &innerErr
		}
		return &ae
	case *ApiError:
		return err.(*ApiError)
	default:
		return &ApiError{http.StatusInternalServerError, err}
	}
}
