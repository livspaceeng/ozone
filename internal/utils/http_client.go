package utils

import (
	"context"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpRequest.Header))
	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
	httpResponse, err := httpClient.Do(httpRequest.WithContext(ctx))

	if httpResponse == nil {
		return httpResponse, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return httpResponse, err
	}
	return httpResponse, nil

}