package interview_accountapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountsClientFactory_MakeHttpClient_NotValidUrl(t *testing.T) {
	clientFactory := AccountsHttpClientFactory{}
	client, err := clientFactory.MakeClient("boom")
	if err == nil {
		t.Errorf("Expecting error to be non-nil")
	} else { // not nil
		expectedErrorMsg := "invalid URL provided"
		if err.Error() != expectedErrorMsg {
			t.Errorf("Expecting error message to be: %s, actual=%s", expectedErrorMsg, err.Error())
		}
	}
	if client != nil {
		t.Errorf("Expecting client to be nil")
	}
}

func TestAccountsClientFactory_MakeHttpClient(t *testing.T) {
	clientFactory := AccountsHttpClientFactory{}
	client, err := clientFactory.MakeClient("http://localhost:8080")
	if err != nil || client == nil {
		t.Errorf("Oops, unable to make a new http client, that's unexpected")
	}
}

func TestFetch_IdIsNotUuid(t *testing.T) {
	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient("https://abc.com")
	account, httpErr := client.Fetch("blah")

	assertHttpError(t, httpErr, &HTTPError{
		Message: "id must be a valid uuid",
	})
	assertAccountData(t, account, nil)
}

func TestFetch_ErrorPlacingGet(t *testing.T) {
	id, _ := uuid.NewUUID()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	err := errors.New("failed placing get")

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeTestClientWithHttpGetter(server.URL,
		func(path string) (*http.Response, error) {
			return nil, err
		})
	account, httpErr := client.Fetch(id.String())

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error placing a Get Http request",
		Cause:   err,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_StatusCodeNotOk(t *testing.T) {
	id, _ := uuid.NewUUID()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPathSuffix := fmt.Sprintf("/%s/%s", servicePath, id)
		if !strings.HasSuffix(r.URL.String(), expectedPathSuffix) {
			t.Errorf("invoked path doesn't match with the expected suffix")
		}
		if r.Method != "GET" {
			t.Errorf("unexpected http method, got=%s, expected=GET", r.Method)
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	account, httpErr := client.Fetch(id.String())

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      400,
		Message:         "Unexpected response code returned for Get operation, expected 200, got 400",
		ResponsePayload: &emptyByteSlice,
		Cause:           nil,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_ErrorProcessingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	err := errors.New("unable to finish reading input stream")

	client, _ := clientFactory.MakeTestClientWithInputReader(server.URL,
		func(reader io.Reader) ([]byte, error) {
			return nil, err
		})

	id, _ := uuid.NewUUID()
	account, httpErr := client.Fetch(id.String())

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error processing response body",
		Cause:   err,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_ContentTypeNotJson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	account, httpErr := client.Fetch(id.String())

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      200,
		Message:         "Unexpected  Content-Type, expecting application/json, got text/html",
		ResponsePayload: &emptyByteSlice,
		Cause:           nil,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_PayloadNotJsonDocument(t *testing.T) {
	payload := []byte("blah")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	account, httpErr := client.Fetch(id.String())

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      0,
		Message:         "Error deserializing json",
		Cause:           errors.New("invalid character 'b' looking for beginning of value"),
		ResponsePayload: &payload,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_PayloadEmptyJsonObject(t *testing.T) {
	payload := []byte("{}")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	account, httpErr := client.Fetch(id.String())

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      0,
		Message:         "Got an empty object after deserialization, json payload was an empty object?",
		Cause:           nil,
		ResponsePayload: &payload,
	})
	assertAccountData(t, account, nil)
}

func TestFetch_HappyPath(t *testing.T) {
	payload := []byte(`{
	"data":{
		"id": "0d209d7f-d07a-4542-947f-5885fddddae2",
		"organisation_id": "ba61483c-d5c5-4f50-ae81-6b8c039bea43",
		"type": "accounts",
		"version": 32,
		"attributes": {
			"account_classification": "Class Zero",
			"account_matching_opt_out": true,
			"alternative_names": ["a","b","c","d"],
			"bank_id": "400300",
			"bank_id_code": "GBDSC",
			"bic": "NWBKGB22",
			"country": "Canada",
			"base_currency": "CAD",
			"iban": "GB11NWBK40030041426819",
			"account_number": "41426819",
			"customer_id": "123",
			"joint_account": false,
			"status": "Pending",
			"switched": true,
			"secondary_identification": "Driver's License 123456",
			"name": ["x", "y", "z"]
			}
  		}
	}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	account, httpErr := client.Fetch(id.String())

	version := int64(32)
	assertHttpError(t, httpErr, nil)

	class := "Class Zero"
	accountMatchingOptOut := true
	country := "Canada"
	jointAccount := false
	status := "Pending"
	switched := true

	assertAccountData(t, account, &AccountData{
		ID:             "0d209d7f-d07a-4542-947f-5885fddddae2",
		OrganisationID: "ba61483c-d5c5-4f50-ae81-6b8c039bea43",
		Type:           "accounts",
		Version:        &version,
		Attributes: &AccountAttributes{
			AccountClassification:   &class,
			AccountMatchingOptOut:   &accountMatchingOptOut,
			AccountNumber:           "41426819",
			AlternativeNames:        []string{"a", "b", "c", "d"},
			BankID:                  "400300",
			BankIDCode:              "GBDSC",
			BaseCurrency:            "CAD",
			Bic:                     "NWBKGB22",
			Country:                 &country,
			CustomerId:              "123",
			Iban:                    "GB11NWBK40030041426819",
			JointAccount:            &jointAccount,
			Name:                    []string{"x", "y", "z"},
			SecondaryIdentification: "Driver's License 123456",
			Status:                  &status,
			Switched:                &switched,
		},
	})
}

func TestDelete_IdIsNotUuid(t *testing.T) {
	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient("https://abc.com")
	httpErr := client.Delete("blah", 2)

	assertHttpError(t, httpErr, &HTTPError{
		Message: "id must be a valid uuid",
	})
}

func TestDelete_StatusCodeNotOk(t *testing.T) {
	id, _ := uuid.NewUUID()
	version := 2
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPathSuffix := fmt.Sprintf("/%s/%s?version=%d", servicePath, id, version)
		if !strings.HasSuffix(r.URL.String(), expectedPathSuffix) {
			t.Errorf("invoked path doesn't match with the expected suffix")
		}
		if r.Method != "DELETE" {
			t.Errorf("unexpected http method, got=%s, expected=DELETE", r.Method)
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	httpErr := client.Delete(id.String(), 2)

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      400,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 400",
		ResponsePayload: &emptyByteSlice,
		Cause:           nil,
	})
}

func TestDelete_CannotCreateNewRequest(t *testing.T) {
	id, _ := uuid.NewUUID()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	err := errors.New("unable to create new request")

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeTestClientWithNewRequestCreator(server.URL,
		func(method string, path string, reader io.Reader) (*http.Request, error) {
			return nil, err
		})
	httpErr := client.Delete(id.String(), 2)

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error preparing Delete Http request",
		Cause:   err,
	})
}

func TestDelete_CannotPlaceRequest(t *testing.T) {
	id, _ := uuid.NewUUID()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	err := errors.New("unable to place request")

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeTestClientWithRequestInvoker(server.URL,
		func(request *http.Request) (*http.Response, error) {
			return nil, err
		})
	httpErr := client.Delete(id.String(), 2)

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error placing Delete Http request",
		Cause:   err,
	})
}

func TestDelete_PayloadNotJsonDocument(t *testing.T) {
	payload := []byte("blah") // we are not deserializing payload on delete, so let's check if it is added on non-204 code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // if status isn't 204, we attach payload to the error
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	httpErr := client.Delete(id.String(), 2)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      400,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 400",
		ResponsePayload: &payload,
		Cause:           nil,
	})
}

func TestDelete_ErrorProcessingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	err := errors.New("unable to finish reading input stream")

	client, _ := clientFactory.MakeTestClientWithInputReader(server.URL,
		func(reader io.Reader) ([]byte, error) {
			return nil, err
		})

	id, _ := uuid.NewUUID()
	httpErr := client.Delete(id.String(), 2)

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error processing response body",
		Cause:   err,
	})
}

func TestDelete_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)
	id, _ := uuid.NewUUID()
	httpErr := client.Delete(id.String(), 3)

	assertHttpError(t, httpErr, nil)
}

func TestCreate_StatusCodeNotCreated(t *testing.T) {
	version := int64(7)
	accountData := &AccountData{
		ID:             "0987654321",
		OrganisationID: "org123",
		Type:           "account",
		Version:        &version,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPathSuffix := fmt.Sprintf("/%s", servicePath)
		if !strings.HasSuffix(r.URL.String(), expectedPathSuffix) {
			t.Errorf("invoked path doesn't match with the expected suffix")
		}
		if r.Method != "POST" {
			t.Errorf("unexpected http method, got=%s, expected=DELETE", r.Method)
		}
		if r.Header.Get(contentType) != jsonContentType {
			t.Errorf("unexpected content type, got=%s, expected=%s", r.Header.Get(contentType), jsonContentType)
		}
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unable to read request's payload")
		}
		var responseEnvelope *Envelope[AccountData]
		err = json.Unmarshal(requestBody, &responseEnvelope)
		if err != nil {
			t.Errorf("Unable to deserialize json payload")
		}
		actualRequestAccountData := responseEnvelope.Data

		assertAccountData(t, actualRequestAccountData, accountData)

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)

	account, httpErr := client.Create(accountData)

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      400,
		Message:         "Unexpected response code returned for Post operation, expected 201, got 400",
		Cause:           nil,
		ResponsePayload: &emptyByteSlice,
	})
	assertAccountData(t, account, nil)
}

func TestCreate_ErrorDeserializingResponseOnEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	client, _ := clientFactory.MakeClient(server.URL)

	account, httpErr := client.Create(&AccountData{})

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		Message:         "Error deserializing json",
		ResponsePayload: &emptyByteSlice,
		Cause:           errors.New("unexpected end of JSON input"),
	})

	assertAccountData(t, account, nil)
}

func TestCreate_ErrorDeserializingResponseNotJsonDocument(t *testing.T) {
	payload := []byte("blah")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	client, _ := clientFactory.MakeClient(server.URL)

	account, httpErr := client.Create(&AccountData{})

	assertHttpError(t, httpErr, &HTTPError{
		Message:         "Error deserializing json",
		ResponsePayload: &payload,
		Cause:           errors.New("invalid character 'b' looking for beginning of value"),
	})

	assertAccountData(t, account, nil)
}

func TestCreate_ResponsePayloadEmptyJsonDocument(t *testing.T) {
	payload := []byte("{}")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write(payload)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	client, _ := clientFactory.MakeClient(server.URL)

	account, httpErr := client.Create(&AccountData{})

	assertHttpError(t, httpErr, &HTTPError{
		Message:         "Got an empty object after deserialization, json payload was an empty object?",
		ResponsePayload: &payload,
		Cause:           nil,
	})

	assertAccountData(t, account, nil)
}

func TestCreate_ErrorSerializingRequestPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	err := errors.New("cannot serialize")

	client, _ := clientFactory.MakeTestClientWithSerializer(server.URL,
		func(a any) ([]byte, error) {
			return nil, err
		})

	account, httpErr := client.Create(&AccountData{})

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Unable to serialize payload",
		Cause:   err,
	})

	assertAccountData(t, account, nil)
}

func TestCreate_ErrorPlacingPostRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	err := errors.New("cannot post")

	client, _ := clientFactory.MakeTestClientWithHttpPoster(server.URL,
		func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, err
		})

	account, httpErr := client.Create(&AccountData{})

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error placing a Post Http request",
		Cause:   err,
	})

	assertAccountData(t, account, nil)
}

func TestCreate_ErrorProcessingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}

	err := errors.New("unable to finish reading input stream")

	client, _ := clientFactory.MakeTestClientWithInputReader(server.URL,
		func(reader io.Reader) ([]byte, error) {
			return nil, err
		})

	account, httpErr := client.Create(&AccountData{})

	assertHttpError(t, httpErr, &HTTPError{
		Message: "Error processing response body",
		Cause:   err,
	})
	assertAccountData(t, account, nil)
}

func TestCreate_HappyPath(t *testing.T) {
	class := "class_A"
	accountMatchingOptOut := false
	country := "Canada"
	jointAccount := true
	status := "Pending"
	switched := true
	version := int64(7)
	requestAccount := &AccountData{
		Attributes: &AccountAttributes{
			AccountClassification:   &class,
			AccountMatchingOptOut:   &accountMatchingOptOut,
			AccountNumber:           "A1234567",
			AlternativeNames:        []string{"x", "y", "z"},
			BankID:                  "bid111",
			BankIDCode:              "bidc222",
			BaseCurrency:            "CAD",
			Bic:                     "bic333",
			Country:                 &country,
			CustomerId:              "cid444",
			Iban:                    "iban_007",
			JointAccount:            &jointAccount,
			Name:                    []string{"a", "b", "c"},
			SecondaryIdentification: "Driver's License",
			Status:                  &status,
			Switched:                &switched,
		},
		ID:             "id666",
		OrganisationID: "orgId777",
		Type:           "Bank Account",
		Version:        &version,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		requestBody, _ := io.ReadAll(r.Body) // returning same payload as response
		w.Write(requestBody)

	}))
	defer server.Close()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(server.URL)

	responseAccount, httpErr := client.Create(requestAccount)

	assertHttpError(t, httpErr, nil)
	assertAccountData(t, responseAccount, requestAccount)
}
