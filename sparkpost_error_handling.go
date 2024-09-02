package main

import (
	"encoding/json"
	"log"
)

type SparkPostErrorHandler struct {
	// You can add fields here if needed, such as a database connection
}

type SparkPostError struct {
	Message     string `json:"message"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type SparkPostErrorResponse struct {
	Errors []SparkPostError `json:"errors"`
}

func NewSparkPostErrorHandler() *SparkPostErrorHandler {
	return &SparkPostErrorHandler{}
}

func (h *SparkPostErrorHandler) HandleSendError(id string, res interface{}, err error) {
	// log.Printf("Failed to send email with SparkPost: %v Transmission ID: %s", err, id)

	errorResponse := h.parseErrorResponse(err.Error())
	if errorResponse != nil && len(errorResponse.Errors) > 0 {
		// for _, e := range errorResponse.Errors {
		// 	log.Printf("SparkPost Error: Code=%s, Message=%s, Description=%s", e.Code, e.Message, e.Description)
		// }
		h.storeErrorMessage(id, *errorResponse)
	} else {
		log.Printf("Unable to parse error response or no errors found")
		h.storeErrorMessage(id, SparkPostErrorResponse{
			Errors: []SparkPostError{{
				Message:     err.Error(),
				Code:        "UNKNOWN",
				Description: "Unable to parse error response",
			}},
		})
	}
}

func (h *SparkPostErrorHandler) parseErrorResponse(errStr string) *SparkPostErrorResponse {
	var errors []SparkPostError
	err := json.Unmarshal([]byte(errStr), &errors)
	if err != nil {
		log.Printf("Error parsing SparkPost response: %v", err)
		return nil
	}

	return &SparkPostErrorResponse{Errors: errors}
}

func (h *SparkPostErrorHandler) storeErrorMessage(id string, errorResponse SparkPostErrorResponse) {
	// TODO: Implement logic to store error message
	// This could include:
	// - Connecting to a database
	// - Formatting the error data
	// - Inserting the data into an 'email_send_errors' table
	// - Potentially triggering alerts or notifications for critical errors

	log.Printf("TODO: Store error message for Transmission ID: %s", id)
	log.Printf("Error details to be stored: %+v", errorResponse)
}
