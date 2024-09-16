package main

import (
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCalculateWeightsForTimeRange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testCases := []struct {
		name     string
		mockData []ProviderStats
	}{
		{
			name: "Balanced data",
			mockData: []ProviderStats{
				{"Provider A", 1000, 800, 50, 600, 10, 20},
				{"Provider B", 1200, 900, 60, 650, 15, 25},
				{"Provider C", 1100, 850, 55, 620, 12, 22},
				{"Provider D", 1300, 950, 40, 670, 8, 15},
			},
		},
		{
			name: "Data with anomaly",
			mockData: []ProviderStats{
				{"Provider A", 100, 80, 5, 60, 1, 2},
				{"Provider B", 120, 90, 39, 65, 2, 3}, // High bounce rate
				{"Provider C", 110, 85, 6, 62, 1, 2},
				{"Provider D", 130, 95, 4, 67, 1, 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"provider_name", "total_events", "delivered_events", "bounce_events", "open_events", "deferred_events", "spam_report_events"})
			for _, stat := range tc.mockData {
				rows.AddRow(stat.Name, stat.TotalEvents, stat.DeliveredEvents, stat.BounceEvents, stat.OpenEvents, stat.DeferredEvents, stat.SpamReportEvents)
			}

			mock.ExpectQuery("SELECT esp.provider_name, COUNT.*").WillReturnRows(rows)

			userID := 1
			credentials := Credentials{}
			startTime := time.Now().AddDate(0, -1, 0)
			endTime := time.Now()

			weights, err := calculateWeightsForTimeRange(db, userID, credentials, startTime, endTime)
			if err != nil {
				t.Fatalf("Error calculating weights: %v", err)
			}

			// Check if weights sum to approximately 1000
			totalWeight := 0
			for _, weight := range weights {
				totalWeight += weight
			}
			if math.Abs(float64(totalWeight-1000)) > 1 {
				t.Errorf("Total weight should be close to 1000, got %d", totalWeight)
			}

			// Check if weights are reasonable (non-zero for valid providers)
			for provider, weight := range weights {
				if isValidProvider(provider, credentials) && weight == 0 {
					t.Errorf("Weight for valid provider %s should not be zero", provider)
				}
			}

			// For the anomaly case, check if the problematic provider has a lower weight
			if tc.name == "Data with anomaly" {
				if weights["Provider B"] >= weights["Provider A"] || weights["Provider B"] >= weights["Provider C"] || weights["Provider B"] >= weights["Provider D"] {
					t.Errorf("Provider B should have a lower weight due to high bounce rate. Weights: %v", weights)
				}
			}

			t.Logf("Calculated weights: %v", weights)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}
