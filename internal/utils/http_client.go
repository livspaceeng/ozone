package utils

import (
	"context"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HttpClient interface {
	SendRequest(ctx context.Context, method string, url string, body io.Reader, headers map[string]string) (*http.Response, error)
}

type httpClient struct {
}

func NewHttpClient(cli *http.Client) HttpClient {
	return &httpClient{}
}

func (httpClnt httpClient) SendRequest(ctx context.Context, method string, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	httpRequest, _ := http.NewRequest(method, url, body)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	httpResponse, err := httpClient.Do(httpRequest.WithContext(ctx))

	if httpResponse == nil {
		return httpResponse, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return httpResponse, err
	}
	return httpResponse, nil

}