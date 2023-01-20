package utils

import (
	"io"
	"net/http"
)

type HttpClient interface {
	SendRequest(method string, url string, body io.Reader, headers map[string]string) (*http.Response, error)
}

type httpClient struct {
}

func NewHttpClient(cli *http.Client) HttpClient {
	return &httpClient{}
}

func (httpClnt httpClient) SendRequest(method string, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	httpRequest, _ := http.NewRequest(method, url, body)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	httpClient := http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)

	if httpResponse == nil {
		return httpResponse, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return httpResponse, err
	}
	return httpResponse, nil

}