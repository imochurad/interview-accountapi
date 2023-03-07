package interview_accountapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type HttpAccountsClient interface {
	// Fetch returns a pointer to an object of type AccountData based on provided identifier.
	// If there is any internal client error during request placement and response analysis,
	// such error will be wrapped in HTTPError object, pointer to which will be returned to the caller.
	// If the response returned is not identified as a successful operation (status code 200),
	// the pointer to instantiated HTTPError object will be returned,
	// the AccountData pointer will be set to nil in this case.
	// The return values are mutually exclusive, you either get a valid AccountData object
	// if operation succeeded or HTTPError if there was any error.
	Fetch(id string) (*AccountData, *HTTPError)

	// Create returns a pointer to a newly created object of type AccountData.
	// If there is any internal client error during request placement and response analysis,
	// such error will be wrapped in HTTPError object, pointer to which will be returned to the caller.
	// If the response returned is not identified as a successful operation (status code 201),
	// the pointer to instantiated HTTPError object will be returned,
	// the AccountData pointer will be set to nil in this case.
	// The return values are mutually exclusive, you either get a valid AccountData object
	// if operation succeeded or HTTPError if there was any error.
	Create(a *AccountData) (*AccountData, *HTTPError)

	// Delete returns a pointer to a HTTPError struct if there was any internal client error
	// during request placement and response analysis.
	// If the response returned is not identified as a successful operation (status code 204),
	// the pointer to instantiated HTTPError object will be returned.
	Delete(id string, version int64) *HTTPError
}

const servicePath = "v1/organisation/accounts"
const jsonContentType = "application/json"
const contentType = "Content-Type"

type ReadInputStream func(io.Reader) ([]byte, error)
type HttpGet func(string) (*http.Response, error)
type HttpPost func(url, contentType string, body io.Reader) (resp *http.Response, err error)
type NewRequest func(string, string, io.Reader) (*http.Request, error)
type DoRequest func(*http.Request) (*http.Response, error)
type Serialize func(any) ([]byte, error)

type httpAccountsClientImpl struct {
	host             string
	client           *http.Client
	readInput        ReadInputStream
	doHttpGet        HttpGet
	doHttpPost       HttpPost
	createNewRequest NewRequest
	doRequest        DoRequest
	serialize        Serialize
}

func (hac *httpAccountsClientImpl) Fetch(id string) (*AccountData, *HTTPError) {
	if !isValidUUID(id) {
		return nil,
			&HTTPError{
				Message: "id must be a valid uuid",
			}
	}

	path := fmt.Sprintf("%s/%s/%s", hac.host, servicePath, id)
	resp, err := hac.doHttpGet(path)
	if err != nil {
		return nil,
			&HTTPError{
				Cause:   err,
				Message: "Error placing a Get Http request",
			}
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	responseData, httpErr := hac.readPayload(resp)
	if httpErr != nil {
		return nil, httpErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil,
			unexpectedStatusCode(http.StatusOK, resp.StatusCode, "Get", responseData)
	}

	cType := resp.Header.Get(contentType)
	if !strings.HasPrefix(cType, jsonContentType) {
		return nil,
			&HTTPError{
				StatusCode:      resp.StatusCode,
				Message:         fmt.Sprintf("Unexpected  %s, expecting %s, got %s", contentType, jsonContentType, cType),
				ResponsePayload: responseData,
			}
	}

	responseEnvelope, httpErr := deserializeToResponseEnvelope(responseData)
	if httpErr != nil {
		return nil, httpErr
	}

	return accountDataOrError(responseEnvelope, responseData)
}

func (hac *httpAccountsClientImpl) Create(account *AccountData) (*AccountData, *HTTPError) {
	requestEnvelope := Envelope[AccountData]{
		Data: account,
	}
	requestData, err := hac.serialize(requestEnvelope)
	if err != nil {
		return nil,
			&HTTPError{
				Cause:   err,
				Message: "Unable to serialize payload",
			}
	}

	reader := bytes.NewReader(requestData)
	resp, err := hac.doHttpPost(hac.host+"/"+servicePath, jsonContentType, reader)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil,
			&HTTPError{
				Cause:   err,
				Message: "Error placing a Post Http request",
			}
	}

	responseData, httpErr := hac.readPayload(resp)
	if httpErr != nil {
		return nil, httpErr
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, unexpectedStatusCode(http.StatusCreated, resp.StatusCode, "Post", responseData)
	}

	responseEnvelope, httpErr := deserializeToResponseEnvelope(responseData)
	if httpErr != nil {
		return nil, httpErr
	}

	return accountDataOrError(responseEnvelope, responseData)
}

func (hac *httpAccountsClientImpl) Delete(id string, version int64) (e *HTTPError) {
	if !isValidUUID(id) {
		return &HTTPError{
			Message: "id must be a valid uuid",
		}
	}

	fullPath := fmt.Sprintf("%s/%s/%s?version=%d", hac.host, servicePath, id, version)

	req, err := hac.createNewRequest(http.MethodDelete, fullPath, nil)

	if err != nil {
		return &HTTPError{
			Cause:   err,
			Message: "Error preparing Delete Http request",
		}
	}

	resp, err := hac.doRequest(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return &HTTPError{
			Cause:   err,
			Message: "Error placing Delete Http request",
		}
	}

	if resp.StatusCode != http.StatusNoContent {
		responseData, httpErr := hac.readPayload(resp)
		if httpErr != nil {
			return httpErr
		}
		return unexpectedStatusCode(http.StatusNoContent, resp.StatusCode, "Delete", responseData)
	}
	return nil
}

func deserializeToResponseEnvelope(responseData *[]byte) (*Envelope[AccountData], *HTTPError) {
	var responseEnvelope *Envelope[AccountData]
	err := json.Unmarshal(*responseData, &responseEnvelope)

	if err != nil {
		return nil, &HTTPError{
			Cause:           err,
			Message:         "Error deserializing json",
			ResponsePayload: responseData,
		}
	}
	return responseEnvelope, nil
}

func accountDataOrError(responseEnvelope *Envelope[AccountData], responseData *[]byte) (*AccountData, *HTTPError) {
	// making sure we are not returning null for the http error and then for the value, making it either-or
	if responseEnvelope.Data == nil {
		return nil, &HTTPError{
			Message:         fmt.Sprintf("Got an empty object after deserialization, json payload was an empty object?"),
			ResponsePayload: responseData,
		}
	}
	return responseEnvelope.Data, nil
}

func (hac *httpAccountsClientImpl) readPayload(resp *http.Response) (*[]byte, *HTTPError) {
	responseData, err := hac.readInput(resp.Body)

	if err != nil {
		return nil, &HTTPError{
			Cause:   err,
			Message: "Error processing response body",
		}
	}
	return &responseData, nil
}

func (hac *httpAccountsClientImpl) init() {
	if hac.readInput == nil {
		hac.readInput = io.ReadAll
	}
	if hac.doHttpGet == nil {
		hac.doHttpGet = hac.client.Get
	}
	if hac.doHttpPost == nil {
		hac.doHttpPost = hac.client.Post
	}
	if hac.createNewRequest == nil {
		hac.createNewRequest = http.NewRequest
	}
	if hac.doRequest == nil {
		hac.doRequest = hac.client.Do
	}
	if hac.serialize == nil {
		hac.serialize = json.Marshal
	}
}

func unexpectedStatusCode(expected int, actual int, operation string, respPayload *[]byte) *HTTPError {
	return &HTTPError{
		StatusCode: actual,
		Message: fmt.Sprintf("Unexpected response code returned for %s operation, expected %d, got %d",
			operation,
			expected,
			actual),
		ResponsePayload: respPayload,
	}
}

type AccountsHttpClientFactory struct{}

func (AccountsHttpClientFactory) MakeClient(baseUrl string) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	client := http.Client{}
	httpClient := httpAccountsClientImpl{
		host:   baseUrl,
		client: &client}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithInputReader(baseUrl string, readInput ReadInputStream) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, readInput: readInput}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithHttpGetter(baseUrl string, doHttpGet HttpGet) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, doHttpGet: doHttpGet}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithHttpPoster(baseUrl string, doHttpPost HttpPost) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, doHttpPost: doHttpPost}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithNewRequestCreator(baseUrl string, createNewRequest NewRequest) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, createNewRequest: createNewRequest}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithRequestInvoker(baseUrl string, doRequest DoRequest) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, doRequest: doRequest}
	httpClient.init()
	return &httpClient, nil
}

func (AccountsHttpClientFactory) MakeTestClientWithSerializer(baseUrl string, serialize Serialize) (HttpAccountsClient, error) {
	if err := validateUrl(baseUrl); err != nil {
		return nil, err
	}
	httpClient := httpAccountsClientImpl{host: baseUrl, client: &http.Client{}, serialize: serialize}
	httpClient.init()
	return &httpClient, nil
}

func validateUrl(baseUrl string) error {
	_, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return errors.New("invalid URL provided")
	}
	return nil
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
