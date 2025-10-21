//go:build !withHttp && !withDns

package shared

import "net/http"

func GetDataRequest(baseUrl string, maxRetries int, cookies ...*http.Cookie) (*http.Response, error) {
	return nil, nil
}

func SendDataRequest(baseUrl string, params string, maxRetries int, cookies ...*http.Cookie) error {
	return nil
}
