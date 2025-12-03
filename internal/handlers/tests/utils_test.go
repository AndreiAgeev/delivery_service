package handler_tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"delivery-system/internal/handlers"
)

func TestExtractUUIDFromPath(t *testing.T) {
	for _, tc := range extractUUIDFromPathTestCases {
		result, err := handlers.ExtractUUIDFromPath(tc.path, tc.prefix)
		if !tc.hasError {
			assert.IsType(t, uuid.UUID{}, result, "expected UUID")
			continue
		}
		assert.Error(t, err, "expected error")
	}
}
