package labeler

import "net/http"

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultHttpClient struct {
	client *http.Client
}

func NewDefaultHttpClient() HttpClient {
	return &DefaultHttpClient{client: &http.Client{}}
}

func (d *DefaultHttpClient) Do(req *http.Request) (*http.Response, error) {
	return d.client.Do(req)
}
