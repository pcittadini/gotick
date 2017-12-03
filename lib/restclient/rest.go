package restclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Endpoint struct {
	Host   string
	Path   string
	Method string
	Body   string
	Token  string
	File   string
	Form   string
}

type Response struct {
	TimeTaken  string
	Endpoint   Endpoint
	Result     string
	StatusCode string
}

// Latency returns the endpoint latency number
func (endpoint Endpoint) Do() (R Response, err error) {

	var statusCode int
	var res string

	url := endpoint.Host + endpoint.Path
	start := time.Now()

	if endpoint.File == "" {
		res, statusCode = httpClient(url,
			[]byte(endpoint.Body),
			endpoint.Method,
			endpoint.Token)
	} else {

		file, err := os.Open(endpoint.File)
		if err != nil {
			return R, err
		}
		defer file.Close()

		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			return R, err
		}

		fi, err := file.Stat()
		if err != nil {
			return R, err
		}
		file.Close()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile(endpoint.Form, fi.Name())
		if err != nil {
			return R, err
		}

		part.Write(fileContents)

		err = writer.Close()
		if err != nil {
			return R, err
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{
			Transport: tr,
		}

		req, _ := http.NewRequest(endpoint.Method, endpoint.Host+endpoint.Path, body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", endpoint.Token)

		resp, err := client.Do(req)
		if err != nil {
			return R, err
		}
		defer resp.Body.Close()

		r, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return R, err
		}

		res = string(r)
	}

	elapsed := time.Since(start)

	// error on the following status codes
	switch statusCode {
	case 400:
		return R, errors.New("status : bad request")
	case 500:
		return R, errors.New("status : internal server error")
	case 403:
		return R, errors.New("status : forbidden")
	case 401:
		return R, errors.New("status : unauthorized")
	}

	R.Endpoint = endpoint
	R.TimeTaken = elapsed.String()
	R.Result = res
	R.StatusCode = strconv.Itoa(statusCode)

	return R, nil
}

func httpClient(url string, jsonData []byte, httpMethod string, authToken string) (apiResponse string, statusCode int) {

	var jsonStr = jsonData
	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(jsonStr))
	// Disable keep-alive
	req.Close = true

	// If we got a auth token set it.
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
		// e.Customer.Username, _ = auth.Token(authToken)
	}
	req.Header.Set("Content-Type", "application/json")

	// Build the client and do the request with a 1 minute timeout
	timeout := time.Duration(60 * time.Second)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("[ util/ClientRest ] " + url + " unreachable")
		return "", 500
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), resp.StatusCode
}
