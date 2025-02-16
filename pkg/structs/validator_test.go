package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID string `validate:"uuid4_or_empty"`
}

func TestValidateUUID4rEmpty(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		isValid bool
	}{
		{"Valid UUID4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"Empty ID", "", true},
		{"Invalid UUID", "invalid-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := TestStruct{ID: tt.id}
			err := ValidateStruct(ts)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
