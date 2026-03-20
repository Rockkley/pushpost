package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func New(upstreamURL string, timeout time.Duration) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
		},
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout,
		},
	}, nil
}

func NewStrippingAuth(upstreamURL string, timeout time.Duration) (*httputil.ReverseProxy, error) {
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
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout,
		},
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
