package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func NewTransport(responseHeaderTimeout time.Duration) *http.Transport {
	return &http.Transport{
		ResponseHeaderTimeout: responseHeaderTimeout,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false,
	}
}

func New(upstreamURL string, transport *http.Transport) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
		},
		Transport: transport,
	}, nil
}

func NewStrippingAuth(upstreamURL string, transport *http.Transport) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			r.Out.Header.Del("Authorization")
		},
		Transport: transport,
	}, nil
}

func NewStrippingAuthWithRewrite(
	upstreamURL string,
	timeout time.Duration,
	rewrite func(path string) string,
) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(upstreamURL)

	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			r.Out.Header.Del("Authorization")

			if rewrite != nil {

				r.Out.URL.Path = rewrite(r.In.URL.Path)
			}
		},
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout,
		},
	}, nil
}
