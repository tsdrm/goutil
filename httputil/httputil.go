package httputil

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/txze/goutil"
	"io"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"os"
	"reflect"
	"strings"
)

const (
	ReturnBytes = iota
	ReturnString
	ReturnJSON
	ReturnXML
)

func formatValue(returnType int, raw []byte, v interface{}) (err error) {
	switch returnType {
	case ReturnBytes:
		value := reflect.ValueOf(v).Elem()
		value.Set(reflect.ValueOf(raw))
	case ReturnString:
		value := reflect.ValueOf(v).Elem()
		value.Set(reflect.ValueOf(string(raw)))
	case ReturnJSON:
		err = json.Unmarshal(raw, v)
	case ReturnXML:
		err = xml.Unmarshal(raw, v)
	default:
		return fmt.Errorf("unknown return type")
	}
	return
}

var HTTPClient = Client{
	Client: http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	},
}

type Client struct {
	http.Client
}

//Get send http request with GET method.
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) Get(url string, returnType int, result interface{}) (err error) {
	return c.GetWithHeader(url, nil, returnType, result)
}

//GetWithHeader send http request with GET method and given header
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) GetWithHeader(url string, header map[string]string, returnType int, result interface{}) (err error) {
	code, _, err := c.Request("GET", url, header, nil, returnType, result)
	if err != nil {
		return
	}

	if code != http.StatusOK {
		return fmt.Errorf("http response code; %v", code)
	}
	return
}

//PostForm issues a POST to the specified URL,
//with data's keys and values URL-encoded as the request body.
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) PostForm(url string, form goutil.Map, returnType int, result interface{}) error {
	values := url2.Values{}
	for k := range form {
		values.Set(k, form.GetString(k))
	}
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	return c.PostWithHeader(url, header, strings.NewReader(values.Encode()), returnType, result)
}

//PostJSON issues a POST to the specified URL,
//with the JSON-encoded data as the request body.
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) PostJSON(url string, body interface{}, returnType int, result interface{}) error {
	header := map[string]string{
		"Content-Type": "application/json",
	}

	bys, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.PostWithHeader(url, header, bytes.NewReader(bys), returnType, result)
}

//PostXML issues a POST to the specified URL,
//with the XML-encoded data as the request body.
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) PostXML(url string, body interface{}, returnType int, result interface{}) error {
	header := map[string]string{
		"Content-Type": "application/xml",
	}

	bys, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	return c.PostWithHeader(url, header, bytes.NewReader(bys), returnType, result)
}

//PostWithHeader issues a POST to the specified URL,
//with a given header and io.Reader as the request body.
//stores the result in the value pointed to by result according to the resultType.
func (c *Client) PostWithHeader(url string, header map[string]string, body io.Reader, returnType int, result interface{}) error {
	code, _, err := c.Request("POST", url, header, body, returnType, result)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("http response code; %v", code)
	}
	return nil
}

//Download issues a GET to the specified URL,
//and save the response data to the savePath.
func (c *Client) Download(url, savePath string) (int64, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("http response code: %v", res.StatusCode)
	}
	f, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	l, err := io.Copy(buf, res.Body)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	buf.Flush()

	return l, err
}

func (c *Client) Request(method, url string, hds map[string]string, data io.Reader, returnType int, result interface{}) (code int, header map[string]string, err error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return
	}

	for k, v := range hds {
		req.Header.Set(k, v)
	}

	res, err := c.Do(req)
	if err != nil {
		return
	}

	header = make(map[string]string)
	for k, _ := range res.Header {
		header[k] = res.Header.Get(k)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	err = formatValue(returnType, body, result)
	if err != nil {
		return
	}
	code = res.StatusCode
	return
}

func Request(method, url string, hds map[string]string, data io.Reader, returnType int, result interface{}) (code int, header map[string]string, err error) {
	return HTTPClient.Request(method, url, hds, data, returnType, result)
}

func Get(url string, returnType int, v interface{}) (err error) {
	return HTTPClient.Get(url, returnType, &v)
}

func GetWithHeader(url string, header map[string]string, returnType int, result interface{}) (err error) {
	return HTTPClient.GetWithHeader(url, header, returnType, &result)
}

func PostForm(url string, form goutil.Map, returnType int, result interface{}) error {
	return HTTPClient.PostForm(url, form, returnType, &result)
}

func PostJSON(url string, body interface{}, returnType int, result interface{}) error {
	return HTTPClient.PostJSON(url, body, returnType, &result)
}

func PostXML(url string, body interface{}, returnType int, result interface{}) error {
	return HTTPClient.PostXML(url, body, returnType, &result)
}

func PostWithHeader(url string, header map[string]string, body io.Reader, returnType int, result interface{}) error {
	return HTTPClient.PostWithHeader(url, header, body, returnType, result)
}

func Download(url, savePath string) (int64, error) {
	return HTTPClient.Download(url, savePath)
}
