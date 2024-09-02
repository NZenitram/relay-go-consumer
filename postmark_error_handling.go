// postmark_error_handling.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PostmarkErrorResponse struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

// HandlePostmarkResponse processes the HTTP response from Postmark
func HandlePostmarkResponse(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return handlePostmarkError(resp.StatusCode, body)
	}

	// Log success response
	fmt.Printf("Email sent successfully. Response: %s\n", string(body))
	return nil
}

// handlePostmarkError processes non-200 status codes
func handlePostmarkError(statusCode int, body []byte) error {
	var postmarkError PostmarkErrorResponse
	if err := json.Unmarshal(body, &postmarkError); err != nil {
		return fmt.Errorf("failed to parse error response: %v", err)
	}

	errorMessage := fmt.Sprintf("Postmark API error: Code=%d, Message=%s", postmarkError.ErrorCode, postmarkError.Message)

	// TODO: Store error for later processing
	storePostmarkErrorForLater(statusCode, postmarkError)

	return fmt.Errorf("%s", errorMessage)
}

// storePostmarkErrorForLater is a placeholder function for storing Postmark errors
func storePostmarkErrorForLater(statusCode int, errorResponse PostmarkErrorResponse) {
	// TODO: Implement error storage logic
	fmt.Printf("TODO: Store Postmark error for later processing: StatusCode=%d, ErrorCode=%d, Message=%s\n",
		statusCode, errorResponse.ErrorCode, errorResponse.Message)
}
