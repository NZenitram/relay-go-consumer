package main

import (
	"log"

	"github.com/socketlabs/socketlabs-go/injectionapi"
)

// SocketLabsErrorHandler manages error handling for SocketLabs email sending
type SocketLabsErrorHandler struct {
	// You can add fields here if needed, such as a database connection
}

// NewSocketLabsErrorHandler creates a new SocketLabsErrorHandler
func NewSocketLabsErrorHandler() *SocketLabsErrorHandler {
	return &SocketLabsErrorHandler{}
}

// SocketLabsResponse represents the response from the SocketLabs API
type SocketLabsResponse struct {
	ErrorCode string `json:"ErrorCode"`
	Message   string `json:"Message"`
	Success   bool   `json:"Success"`
	// Add other fields as needed based on the actual API response
}

// HandleSendError processes errors from SocketLabs email sending attempts
func (h *SocketLabsErrorHandler) HandleSendError(toEmail string, err error, response *injectionapi.SendResponse) {
	log.Printf("Failed to send email to %s: %v", toEmail, err)

	if response != nil {
		log.Printf("SocketLabs Response: Result=%s, TransactionReceipt=%s, ResponseMessage=%s",
			response.Result.ToString(), response.TransactionReceipt, response.ResponseMessage)

		for _, addressResult := range response.AddressResults {
			log.Printf("Address Result: Email=%s, Accepted=%t, ErrorCode=%s",
				addressResult.EmailAddress, addressResult.Accepted, addressResult.ErrorCode)
		}

		if response.Result != injectionapi.SendResultSUCCESS {
			log.Printf("SocketLabs Error: %s", response.Result.ToResponseMessage())
		}

		h.storeSendFailure(toEmail, err, response)
	} else {
		h.storeSendFailure(toEmail, err, nil)
	}
}

// storeSendFailure is a placeholder for future implementation of storing send failures
func (h *SocketLabsErrorHandler) storeSendFailure(toEmail string, err error, response *injectionapi.SendResponse) {
	// TODO: Implement logic to save send failures to a database or log processing
	// This could include:
	// - Connecting to a database
	// - Formatting the error and response data
	// - Inserting the data into a 'failed_sends' table
	// - Potentially triggering alerts or notifications for critical failures

	log.Printf("TODO: Store send failure for %s in database", toEmail)

	if response != nil {
		// Example of data you might want to store
		failureData := map[string]interface{}{
			"toEmail":            toEmail,
			"error":              err.Error(),
			"sendResult":         response.Result.ToString(),
			"responseMessage":    response.ResponseMessage,
			"transactionReceipt": response.TransactionReceipt,
			"addressResults":     response.AddressResults,
		}

		log.Printf("Failure data to be stored: %+v", failureData)
	}
}

// You can add more methods here as needed, such as:
// - Methods to retrieve and analyze stored failures
// - Methods to retry failed sends
// - Methods to generate reports on send failures
