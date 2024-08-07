package client

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"go_code/myselfgo/utils"
)

const (
	errNilConnPool = "http connection pool is nil"
)

// HTTPClient specifies the information of http client
type HTTPClient struct {
	Timeout             int          // timeout specifies a time limit for requests made by this http client
	MaxConnsPerHost     int          // limit the total number of connections per host
	MaxIdleConnsPerHost int          // control the maximum idle(keep-alive) connections to keep per host
	API                 string       //  api address of http server
	Client              *http.Client // http connection pool
}

// PostJSON post with json body and application/json header
// return (body, statusCode, contentType, error)
func (hc *HTTPClient) PostJSON(api string, data []byte) ([]byte, int, string, error) {
	if hc == nil || hc.Client == nil {
		return []byte{}, 0, "", errors.New(errNilConnPool)
	}

	if api == "" {
		api = hc.API
	}

	// headers
	headers := map[string]string{
		"Content-Type": "application/json; charset=UTF-8",
	}

	// create a new request
	request, err := hc.newRequest("POST", api, data, headers)
	if err != nil {
		return []byte{}, 0, "", err
	}

	// do request
	return hc.doRequest(request)
}

// PostForm posts with form data
// return (body, statusCode, contentType, error)
func (hc *HTTPClient) PostForm(api string, valueMap map[string]string) ([]byte, int, string, error) {
	if hc == nil || hc.Client == nil {
		return []byte{}, 0, "", errors.New(errNilConnPool)
	}

	if api == "" {
		api = hc.API
	}

	// headers
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	// url encode
	values := url.Values{}
	for k, v := range valueMap {
		values.Set(k, v)
	}

	// create a new request
	request, err := hc.newRequest("POST", api, []byte(values.Encode()), headers)
	if err != nil {
		return []byte{}, 0, "", err
	}

	return hc.doRequest(request)
}

// PostRaw posts with binary data
// return (body, statusCode, contentType, error)
func (hc *HTTPClient) PostRaw(api string, rawData []byte) ([]byte, int, string, error) {
	if hc == nil || hc.Client == nil {
		return []byte{}, 0, "", errors.New(errNilConnPool)
	}

	if api == "" {
		api = hc.API
	}

	// headers
	headers := map[string]string{
		"Content-Type": "Content-Type: application/octet-stream",
	}

	// create a new request
	request, err := hc.newRequest("POST", api, rawData, headers)
	if err != nil {
		return []byte{}, 0, "", err
	}

	return hc.doRequest(request)
}

// Get creates a http get request to server
// return (body, statusCode, contentType, error)
func (hc *HTTPClient) Get(api string) ([]byte, int, string, error) {
	if hc == nil || hc.Client == nil {
		return []byte{}, 0, "", errors.New(errNilConnPool)
	}

	if api == "" {
		api = hc.API
	}

	// headers
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	// create a new request
	request, err := hc.newRequest("GET", api, nil, headers)
	if err != nil {
		return []byte{}, 0, "", err
	}

	return hc.doRequest(request)
}

// GetWithParam creates a http get request with param encoded in url to server
// return (body, statusCode, contentType, error)
func (hc *HTTPClient) GetWithParam(api string, valueMap map[string]string, headers map[string]string) ([]byte, int, string, error) {
	if hc == nil || hc.Client == nil {
		return []byte{}, 0, "", errors.New(errNilConnPool)
	}

	if api == "" {
		api = hc.API
	}

	// headers
	if headers == nil {
		headers = map[string]string{
			"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		}
	}

	// url encode
	values := url.Values{}
	for k, v := range valueMap {
		values.Set(k, v)
	}

	newAPI := utils.JoinStrWithSep("?", api, values.Encode())
	// create a new request
	request, err := hc.newRequest("GET", newAPI, nil, headers)
	if err != nil {
		return []byte{}, 0, "", err
	}

	return hc.doRequest(request)
}

// newRequest creates a new http request
func (hc *HTTPClient) newRequest(method, api string, data []byte, headers map[string]string) (*http.Request, error) {
	// create a new request
	request, err := http.NewRequest(method, api, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// set header
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	return request, nil
}

//	doRequest do request and return (body, statusCode, contentType, error)
//
// return (response, code, Content-Type, error)
func (hc *HTTPClient) doRequest(request *http.Request) ([]byte, int, string, error) {
	code := http.StatusOK // use 200 by default
	var err error
	var resp *http.Response

	// do request
	resp, err = hc.Client.Do(request)
	if err != nil { // Client.Timeout exceeded while awaiting headers, then status code will be 504
		code = http.StatusGatewayTimeout
		return []byte{}, http.StatusGatewayTimeout, "", err // set timeout code in case of error
	}
	defer resp.Body.Close() // official examples also ignore error in https://golang.org/pkg/net/http/

	contentType := resp.Header.Get("Content-Type")
	// discard the body when the status code is not 200
	code = resp.StatusCode
	if code != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body) // you must read the body to re-use the http connection
		return []byte{}, code, contentType, err
	}

	// try to read the body when status code is 200
	// it maybe occur error that net/http: request canceled (Client.Timeout exceeded while reading body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		code = http.StatusGatewayTimeout
		return []byte{}, code, contentType, err
	}

	// succeed
	return body, http.StatusOK, contentType, nil // the err will be nil when succeed to read the response body
}
