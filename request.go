package govcr

import (
	"net/http"
	"net/url"
	"regexp"
)

// A RequestFilter can be used to remove / amend undesirable header / body elements from the request.
//
// For instance, if your application sends requests with a timestamp held in a part of
// the header / body, you likely want to remove it or force a static timestamp via
// RequestFilterFunc to ensure that the request body matches those saved on the cassette's track.
//
// A Filter should return the request with any modified values.
type RequestFilter func(req Request) Request

// RequestFilters is a slice of RequestFilter
type RequestFilters []RequestFilter

// A Request provides the request parameters.
type Request struct {
	Header http.Header
	Body   []byte
	Method string
	URL    url.URL
}

// RequestAddHeaderValue will add or overwrite a header to the request
// before the request is matched against the cassette.
func RequestAddHeaderValue(key, value string) RequestFilter {
	return func(req Request) Request {
		req.Header.Add(key, value)
		return req
	}
}

// RequestDeleteHeaderKeys will delete one or more header keys on the request
// before the request is matched against the cassette.
func RequestDeleteHeaderKeys(keys ...string) RequestFilter {
	return func(req Request) Request {
		for _, key := range keys {
			req.Header.Del(key)
		}
		return req
	}
}

// RequestExcludeHeaderFunc is a hook function that is used to filter the Header.
//
// Typically this can be used to remove / amend undesirable custom headers from the request.
//
// For instance, if your application sends requests with a timestamp held in a custom header,
// you likely want to exclude it from the comparison to ensure that the request headers are
// considered a match with those saved on the cassette's track.
//
// Parameters:
//  - parameter 1 - Name of header key in the Request
//
// Return value:
// true - exclude header key from comparison
// false - retain header key for comparison
//
// Deprecated - This function will be removed on or after April 25th 2019
func RequestExcludeHeaderFunc(fn func(key string) bool) RequestFilter {
	return func(req Request) Request {
		for key := range req.Header {
			if fn(key) {
				req.Header.Del(key)
			}
		}
		return req
	}
}

// OnMethod will return a new filter that will only apply 'r'
// if the method of the request matches.
// Original filter is unmodified.
func (r RequestFilter) OnMethod(method string) RequestFilter {
	return func(req Request) Request {
		if req.Method != method {
			return req
		}
		return r(req)
	}
}

// OnPath will return a request filter that will only apply 'r'
// if the url string of the request matches the supplied regex.
// Original filter is unmodified.
func (r RequestFilter) OnPath(pathRegEx string) RequestFilter {
	if pathRegEx == "" {
		pathRegEx = ".*"
	}
	re := regexp.MustCompile(pathRegEx)
	return func(req Request) Request {
		if !re.MatchString(req.URL.String()) {
			return req
		}
		return r(req)
	}
}

// Add one or more filters at the end of the filter chain.
func (r *RequestFilters) Add(filters ...RequestFilter) {
	v := *r
	v = append(v, filters...)
	*r = v
}

// Prepend one or more filters before the current ones.
func (r *RequestFilters) Prepend(filters ...RequestFilter) {
	src := *r
	dst := make(RequestFilters, 0, len(filters)+len(src))
	dst = append(dst, filters...)
	*r = append(dst, src...)
}

// combined returns the filters as a single filter.
func (r RequestFilters) combined() RequestFilter {
	return func(req Request) Request {
		for _, filter := range r {
			req = filter(req)
		}
		return req
	}
}