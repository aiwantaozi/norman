package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rancher/norman/httperror"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const reqMaxSize = (2 * 1 << 20) + 1

var bodyMethods = map[string]bool{
	http.MethodPut:  true,
	http.MethodPost: true,
}

type Decode func(interface{}) error

func ReadBody(req *http.Request) (map[string]interface{}, error) {
	if !bodyMethods[req.Method] {
		return nil, nil
	}

	contentType := req.Header.Get("Content-type")
	return decodeBody(contentType, io.LimitReader(req.Body, maxFormSize))
}

func getDecoder(contentType string, reader io.Reader) Decode {
	if contentType == "application/yaml" {
		return yaml.NewYAMLToJSONDecoder(reader).Decode
	}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	return decoder.Decode
}

func ReadBodyWithoutLosingContent(req *http.Request) (map[string]interface{}, error) {
	if !bodyMethods[req.Method] {
		return nil, nil
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	contentType := req.Header.Get("Content-type")
	return decodeBody(contentType, bytes.NewReader(bodyBytes))
}

func decodeBody(contentType string, reader io.Reader) (map[string]interface{}, error) {
	decode := getDecoder(contentType, reader)

	data := map[string]interface{}{}
	if err := decode(&data); err != nil {
		return nil, httperror.NewAPIError(httperror.InvalidBodyContent,
			fmt.Sprintf("Failed to parse body: %v", err))
	}
	return data, nil
}
