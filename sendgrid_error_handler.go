package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sendgrid/rest"
)

type SendgridError struct {
	Message string      `json:"message"`
	Field   string      `json:"field"`
	Help    interface{} `json:"help"`
}

type SendgridErrorResponse struct {
	Errors []SendgridError `json:"errors"`
}

func HandleSendgridError(res *rest.Response, err error, to string) {
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return
	}

	if res.StatusCode != 200 {
		var errorResponse SendgridErrorResponse
		err := json.Unmarshal([]byte(res.Body), &errorResponse)
		if err != nil {
			log.Printf("Failed to decode error response: %v", err)
			return
		}

		for _, sendgridErr := range errorResponse.Errors {
			errorMessage := fmt.Sprintf("Sendgrid error for %s: %s (Field: %s)", to, sendgridErr.Message, sendgridErr.Field)
			log.Println(errorMessage)

			// TODO: Decide where to store the error information
			// Options to consider:
			// 1. Write to a dedicated error log file
			// 2. Store in a database table for error tracking
			// 3. Send to an error monitoring service (e.g., Sentry, Rollbar)
			// 4. Publish to a message queue for further processing

			// Example placeholder for storing the error:
			storeError(errorMessage)
		}
	}
}

func storeError(errorMessage string) {
	// TODO: Implement error storage mechanism
	log.Println("Error stored:", errorMessage)
}
