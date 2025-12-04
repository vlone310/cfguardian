package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEmail string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid email",
			input:     "user@example.com",
			wantEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name:      "valid email with uppercase",
			input:     "User@Example.COM",
			wantEmail: "user@example.com", // should be normalized to lowercase
			wantErr:   false,
		},
		{
			name:      "valid email with subdomain",
			input:     "user@mail.example.com",
			wantEmail: "user@mail.example.com",
			wantErr:   false,
		},
		{
			name:      "valid email with plus sign",
			input:     "user+tag@example.com",
			wantEmail: "user+tag@example.com",
			wantErr:   false,
		},
		{
			name:    "empty email",
			input:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "missing @ symbol",
			input:   "userexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing local part",
			input:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing domain",
			input:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid characters",
			input:   "user name@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "too long",
			input:   string(make([]byte, 260)) + "@example.com",
			wantErr: true,
			errMsg:  "email too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, email.IsEmpty(), "Email should be empty on error")
			} else {
				require.NoError(t, err)
				assert.False(t, email.IsEmpty(), "Email should not be empty")
				assert.Equal(t, tt.wantEmail, email.String())
			}
		})
	}
}

func TestEmail_String(t *testing.T) {
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)
	
	assert.Equal(t, "test@example.com", email.String())
}

func TestEmail_Equals(t *testing.T) {
	email1, _ := NewEmail("test@example.com")
	email2, _ := NewEmail("test@example.com")
	email3, _ := NewEmail("other@example.com")
	
	// Same email values should be equal
	assert.True(t, email1.Equals(email2))
	
	// Different email values should not be equal
	assert.False(t, email1.Equals(email3))
	
	// Empty email should not equal valid email
	emptyEmail := Email{}
	assert.False(t, email1.Equals(emptyEmail))
}

func TestEmail_Normalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"USER@DOMAIN.COM", "user@domain.com"},
		{"MiXeD@CaSe.com", "mixed@case.com"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, email.String())
		})
	}
}

