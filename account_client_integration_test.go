package interview_accountapi

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"testing"
)

func Test_Integration_Fetch_IdIsNotUuid(t *testing.T) {
	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	account, httpErr := client.Fetch("blah")
	assertHttpError(t, httpErr, &HTTPError{
		Message: "id must be a valid uuid",
	})
	assertAccountData(t, account, nil)
}

func Test_Integration_Fetch_404(t *testing.T) {
	id, _ := uuid.NewUUID()
	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	account, httpErr := client.Fetch(id.String())
	expectedPayload := []byte(`{"error_message":"record ` + id.String() + ` does not exist"}`)
	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      404,
		Message:         "Unexpected response code returned for Get operation, expected 200, got 404",
		ResponsePayload: &expectedPayload,
	})
	assertAccountData(t, account, nil)
}

func Test_Integration_CreateBadRequest(t *testing.T) {
	class := "Unexpected Classification"
	accountMatchingOptOut := false
	country := "CA"
	jointAccount := true
	status := "pending"
	switched := true
	requestAccount := &AccountData{
		Attributes: &AccountAttributes{
			AccountClassification:   &class,
			AccountMatchingOptOut:   &accountMatchingOptOut,
			AccountNumber:           "A1234567",
			AlternativeNames:        []string{"x", "y", "z"},
			BankID:                  "GBDSC",
			BankIDCode:              "BIDC",
			BaseCurrency:            "CAD",
			Bic:                     "AAAAAABB",
			Country:                 &country,
			Iban:                    "II00",
			JointAccount:            &jointAccount,
			Name:                    []string{"a", "b", "c"},
			SecondaryIdentification: "Driver's License",
			Status:                  &status,
			Switched:                &switched,
		},
		ID:             uuid.NewString(),
		OrganisationID: uuid.NewString(),
		Type:           "accounts",
	}

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	createRespAccount, httpErr := client.Create(requestAccount)

	responsePayload := []byte(`{"error_message":"validation failure list:\nvalidation failure list:\nvalidation failure list:\naccount_classification in body should be one of [Personal Business]"}`)

	assertHttpError(t, httpErr, &HTTPError{
		StatusCode:      400,
		ResponsePayload: &responsePayload,
		Message:         "Unexpected response code returned for Post operation, expected 201, got 400",
	})
	assertAccountData(t, createRespAccount, nil)
}

func Test_Integration_CreateDuplicateConstraint(t *testing.T) {
	requestAccount := getValidAccountData()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	createRespAccount, httpErr := client.Create(requestAccount)

	assertHttpError(t, httpErr, nil)
	assertAccountData(t, createRespAccount, requestAccount)

	createRespAccount, httpErr = client.Create(requestAccount)

	responsePayload := []byte(`{"error_message":"Account cannot be created as it violates a duplicate constraint"}`)

	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		Message:         "Unexpected response code returned for Post operation, expected 201, got 409",
		StatusCode:      409,
		ResponsePayload: &responsePayload,
	})
	assertAccountData(t, createRespAccount, nil)
}

func Test_Integration_CreateFetchDelete(t *testing.T) {
	requestAccount := getValidAccountData()

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	createRespAccount, httpErr := client.Create(requestAccount)

	assertHttpError(t, httpErr, nil)
	assertAccountData(t, createRespAccount, requestAccount)

	fetchRespAccount, httpErr := client.Fetch(createRespAccount.ID)
	assertHttpError(t, httpErr, nil)
	assertAccountData(t, fetchRespAccount, requestAccount)

	httpErr = client.Delete(fetchRespAccount.ID, *fetchRespAccount.Version)
	assertHttpError(t, httpErr, nil)

	httpErr = client.Delete(fetchRespAccount.ID, *fetchRespAccount.Version)

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 404",
		StatusCode:      404,
		ResponsePayload: &emptyByteSlice,
	})

	expectedPayload := []byte(`{"error_message":"record ` + requestAccount.ID + ` does not exist"}`)

	fetchRespAccount, httpErr = client.Fetch(createRespAccount.ID)
	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		ResponsePayload: &expectedPayload,
		StatusCode:      404,
		Message:         "Unexpected response code returned for Get operation, expected 200, got 404",
	})
	assertAccountData(t, fetchRespAccount, nil)
}

func Test_Integration_Delete(t *testing.T) {
	requestAccount := getValidAccountData()

	id := requestAccount.ID
	version := requestAccount.Version

	clientFactory := AccountsHttpClientFactory{}
	client, _ := clientFactory.MakeClient(getBaseUrl())
	httpErr := client.Delete(id, *version) // there is nothing to delete, 404 is the correct answer

	emptyByteSlice := make([]byte, 0)

	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		StatusCode:      404,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 404",
		ResponsePayload: &emptyByteSlice,
	})

	client.Create(requestAccount)

	wrongVersion := int64(1)
	responsePayload := []byte(`{"error_message":"invalid version"}`)

	httpErr = client.Delete(id, wrongVersion)

	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		StatusCode:      409,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 409",
		ResponsePayload: &responsePayload,
	})

	httpErr = client.Delete(id, *version)
	assertHttpError(t, httpErr, nil)

	httpErr = client.Delete(id, *version) // there is nothing to delete, 404 is the correct answer

	assertHttpError(t, httpErr, &HTTPError{
		Cause:           nil,
		StatusCode:      404,
		Message:         "Unexpected response code returned for Delete operation, expected 204, got 404",
		ResponsePayload: &emptyByteSlice,
	})
}

func getValidAccountData() *AccountData {
	accountMatchingOptOut := false
	country := "CA"
	jointAccount := true
	status := "pending"
	switched := true
	version := int64(0)
	id := uuid.NewString()
	requestAccount := &AccountData{
		Attributes: &AccountAttributes{
			AccountMatchingOptOut:   &accountMatchingOptOut,
			AccountNumber:           "A1234567",
			AlternativeNames:        []string{"x", "y", "z"},
			BankID:                  "GBDSC",
			BankIDCode:              "BIDC",
			BaseCurrency:            "CAD",
			Bic:                     "AAAAAABB",
			Country:                 &country,
			Iban:                    "II00",
			JointAccount:            &jointAccount,
			Name:                    []string{"a", "b", "c"},
			SecondaryIdentification: "Driver's License",
			Status:                  &status,
			Switched:                &switched,
		},
		ID:             id,
		OrganisationID: uuid.NewString(),
		Type:           "accounts",
		Version:        &version,
	}
	return requestAccount
}

func getBaseUrl() string {
	baseServiceUrl := os.Getenv("ACCOUNTS_SERVICE_BASE_URL")
	if baseServiceUrl == "" {
		baseServiceUrl = "http://localhost:8080"
	}
	fmt.Println("Base Service URL = " + baseServiceUrl)
	return baseServiceUrl
}
