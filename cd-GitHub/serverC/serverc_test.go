package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	testCases := []struct {
		desc   string
		val    int
		valAdd int
		want   int
	}{
		{
			"add 100 to 8",
			8,
			100,
			108,
		}, {
			"add 100 to 20",
			20,
			100,
			120,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {

			body := &Value{
				ServiceName: "serverB",
				Value:       testCase.val,
			}

			payloadBuf := new(bytes.Buffer)
			err := json.NewEncoder(payloadBuf).Encode(body)
			if err != nil {
				fmt.Println("JSON Encode error in Test")
			}
			request, _ := http.NewRequest(http.MethodPost, "/post", payloadBuf)
			response := httptest.NewRecorder()
			gm := NewGlobalVarManager()
			gm.postCall(response, request)

			requestGet, _ := http.NewRequest(http.MethodPost, "/get", nil)
			responseGet := httptest.NewRecorder()
			results := []Value{}

			gm.getCall(responseGet, requestGet)

			err = json.NewDecoder(responseGet.Body).Decode(&results)

			if err != nil {
				fmt.Printf("JSON Decode error in Test, %v", err)
			}
			if len(results) == len(testCases) {
				fmt.Println(results)
				for index := range results {
					if results[index].Value != testCases[index].want {
						t.Errorf("Test Failed - got %v, want %v", results[index].Value, testCases[index].want)
					}

				}
			}

		})
	}

}
