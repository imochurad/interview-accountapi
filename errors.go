package interview_accountapi

type HTTPError struct {
	Cause           error
	Message         string
	StatusCode      int
	ResponsePayload *[]byte
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + " : " + e.Cause.Error()
}
