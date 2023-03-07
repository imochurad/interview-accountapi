package interview_accountapi

import (
	"strings"
	"testing"
)

func assertAccountData(t *testing.T, actual *AccountData, expected *AccountData) {
	if expected != nil && actual == nil {
		t.Errorf("Expecting account data to be not nil")
		return
	}

	if expected == nil && actual != nil {
		t.Errorf("Expecting acount date to be nil")
		return
	}

	if expected == nil && actual == nil {
		return
	}

	if expected.ID != actual.ID {
		t.Errorf("AccountData ID doesn't match, expected=%s, got=%s", expected.ID, actual.ID)
	}

	if expected.OrganisationID != actual.OrganisationID {
		t.Errorf("AccountData OrganisationID doesn't match, expected=%s, got=%s", expected.OrganisationID, actual.OrganisationID)
	}

	if expected.Type != actual.Type {
		t.Errorf("AccountData Type doesn't match, expected=%s, got=%s", expected.Type, actual.Type)
	}

	assertPrimitivePointers(t, actual.Version, expected.Version, "Version")

	assertAttributes(t, actual.Attributes, expected.Attributes)
}

func assertAttributes(t *testing.T, actual *AccountAttributes, expected *AccountAttributes) {
	if expected != nil && actual == nil {
		t.Errorf("Expecting account attributes to be not nil")
		return
	}

	if expected == nil && actual != nil {
		t.Errorf("Expecting account attributes to be nil")
		return
	}

	if expected == nil && actual == nil {
		return
	}

	assertPrimitivePointers(t, actual.AccountClassification, expected.AccountClassification, "Attributes.AccountClassification")
	assertPrimitivePointers(t, actual.AccountMatchingOptOut, expected.AccountMatchingOptOut, "Attributes.AccountMatchingOptOut")
	assertPrimitivePointers(t, actual.Country, expected.Country, "Attributes.Country")
	assertPrimitivePointers(t, actual.JointAccount, expected.JointAccount, "Attributes.JointAccount")
	assertPrimitivePointers(t, actual.Status, expected.Status, "Attributes.Status")
	assertPrimitivePointers(t, actual.Switched, expected.Switched, "Attributes.Switched")

	if expected.AccountNumber != actual.AccountNumber {
		t.Errorf("AccountData Attributes.AccountNumber doesn't match, expected=%s, got=%s",
			expected.AccountNumber,
			actual.AccountNumber)
	}

	if !assertPrimitiveSlices(actual.AlternativeNames, expected.AlternativeNames) {
		actualAlternativeNamesStr := "nil"
		expectedAlternativeNamesStr := "nil"
		if actual.AlternativeNames != nil {
			actualAlternativeNamesStr = strings.Join(actual.AlternativeNames, ",")
		}
		if expected.AlternativeNames != nil {
			expectedAlternativeNamesStr = strings.Join(expected.AlternativeNames, ",")
		}

		t.Errorf("AlternativeNames doesn't match with the expected value, expected=%s, got=%s",
			actualAlternativeNamesStr, expectedAlternativeNamesStr)
	}

	if expected.BankID != actual.BankID {
		t.Errorf("AccountData Attributes.BankID doesn't match, expected=%s, got=%s", expected.BankID, actual.BankID)
	}

	if expected.BankIDCode != actual.BankIDCode {
		t.Errorf("AccountData Attributes.BankID doesn't match, expected=%s, got=%s", expected.BankIDCode, actual.BankIDCode)
	}

	if expected.BaseCurrency != actual.BaseCurrency {
		t.Errorf("AccountData Attributes.BaseCurrency doesn't match, expected=%s, got=%s", expected.BaseCurrency, actual.BaseCurrency)
	}

	if expected.Bic != actual.Bic {
		t.Errorf("AccountData Attributes.Bic doesn't match, expected=%s, got=%s", expected.Bic, actual.Bic)
	}

	if expected.Iban != actual.Iban {
		t.Errorf("AccountData Attributes.Iban doesn't match, expected=%s, got=%s", expected.Iban, actual.Iban)
	}

	if expected.CustomerId != actual.CustomerId {
		t.Errorf("AccountData Attributes.CustomerId doesn't match, expected=%s, got=%s", expected.CustomerId, actual.CustomerId)
	}

	if expected.SecondaryIdentification != actual.SecondaryIdentification {
		t.Errorf("AccountData Attributes.SecondaryIdentification doesn't match, expected=%s, got=%s",
			expected.SecondaryIdentification,
			actual.SecondaryIdentification)
	}

	if !assertPrimitiveSlices(actual.Name, expected.Name) {
		actualNamesStr := "nil"
		expectedNamesStr := "nil"
		if actual.Name != nil {
			actualNamesStr = strings.Join(actual.Name, ",")
		}
		if expected.Name != nil {
			expectedNamesStr = strings.Join(expected.Name, ",")
		}

		t.Errorf("Name doesn't match with the expected value, expected=%s, got=%s",
			actualNamesStr, expectedNamesStr)
	}
}

func assertHttpError(t *testing.T, actual *HTTPError, expected *HTTPError) {
	if expected != nil && actual == nil {
		t.Errorf("Expecting http error to be not nil")
		return
	}

	if expected == nil && actual != nil {
		t.Errorf("Expecting http error to be nil")
		return
	}

	if expected == nil && actual == nil {
		return
	}

	if expected.Cause == nil && actual.Cause != nil {
		t.Errorf("HttpError cause should be nil")
	}

	if expected.Cause != nil && actual.Cause == nil {
		t.Errorf("HttpError cause should not be nil")
	}

	if actual.Message != expected.Message {
		t.Errorf("HttpError message doesn't match, expected=%s, got=%s", expected.Message, actual.Message)
	}

	if actual.StatusCode != expected.StatusCode {
		t.Errorf("HttpError status code doesn't match, expected=%d, got=%d", expected.StatusCode, actual.StatusCode)
	}

	if actual.Error() != expected.Error() {
		t.Errorf("HttpError detailed message doesn't match, expected=%s, got=%s", expected.Error(), actual.Error())
	}

	if actual.ResponsePayload == nil && expected.ResponsePayload != nil {
		t.Errorf("Actual response payload should not be nil")
	}

	if actual.ResponsePayload != nil && expected.ResponsePayload == nil {
		t.Errorf("Actual response payload should be nil")
	}

	if actual.ResponsePayload != nil && expected.ResponsePayload != nil &&
		!assertPrimitiveSlices(*actual.ResponsePayload, *expected.ResponsePayload) {
		actualRespPayloadStr := "nil"
		expectedRespPayloadStr := "nil"
		if actual.ResponsePayload != nil {
			actualRespPayloadStr = string(*actual.ResponsePayload)
		}
		if expected.ResponsePayload != nil {
			expectedRespPayloadStr = string(*expected.ResponsePayload)
		}

		t.Errorf("Payload byte slice doesn't match with the expected value, expected=%s, got=%s",
			expectedRespPayloadStr, actualRespPayloadStr)
	}

}
