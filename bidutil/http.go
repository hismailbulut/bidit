package bidutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	ERROR_CODE_SUCCESS                 = 0
	ERROR_CODE_REQUEST_FAILED          = -1
	ERROR_CODE_BODY_READ_FAILED        = -2
	ERROR_CODE_REQUEST_CREATION_FAILED = -3
)

type Params map[any]any

func makeURL(base string, path []string, params Params) string {
	urlStr := base
	for i := range path {
		urlStr += "/" + path[i]
	}
	if len(params) > 0 {
		urlStr += "?"
		i := 0
		for key, value := range params {
			urlStr += fmt.Sprint(key) + "=" + fmt.Sprint(value)
			if i != len(params)-1 {
				urlStr += "&"
			}
			i++
		}
	}
	return urlStr
}

func doRequest(req *http.Request) ([]byte, error, int) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err, ERROR_CODE_REQUEST_FAILED
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err, ERROR_CODE_BODY_READ_FAILED
	}
	if resp.StatusCode != http.StatusOK {
		print("Not OK Resp:", string(body))
		return nil, errors.New(fmt.Sprint(resp.Status)), resp.StatusCode
	}
	return body, nil, 0
}

func GetRequest(header http.Header, base string, path []string, params Params) ([]byte, error, int) {
	url := makeURL(base, path, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err, ERROR_CODE_REQUEST_CREATION_FAILED
	}
	req.Header = header
	body, err, code := doRequest(req)
	if err != nil {
		return nil, err, code
	}
	return body, nil, code
}

func PostRequest(header http.Header, base string, path []string, params Params, payload []byte /* json string */) ([]byte, error, int) {
	url := makeURL(base, path, params)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err, ERROR_CODE_REQUEST_CREATION_FAILED
	}
	req.Header = header
	body, err, code := doRequest(req)
	if err != nil {
		return nil, err, code
	}
	return body, nil, code
}
