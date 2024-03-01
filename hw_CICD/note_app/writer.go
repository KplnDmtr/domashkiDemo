package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ServerResponseWriter(w http.ResponseWriter, header int, payload interface{}) {
	var JSONbytes []byte
	var err error
	JSONbytes, err = json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(header)
	_, err = w.Write(JSONbytes)
	if err != nil {
		fmt.Printf("write error occured\n")
		return
	}
}
