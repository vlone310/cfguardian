package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPasswordHasher(t *testing.T) {
	hasher := NewPasswordHasher(12)
	require.NotNil(t, hasher)
}

func TestPasswordHasher_HashPassword(t *testing.T) {
	hasher := NewPasswordHasher(10) // Use cost 10 for faster tests
	
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!#$%^&*()"},
		{"long password", "this-is-a-very-long-password-with-many-characters-1234567890"},
		{"unicode password", "pāsswörd123"},
		{"empty password", ""}, // bcrypt can hash empty strings
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.Hash(tt.password)
			
			if tt.password == "" || len(tt.password) < 8 {
				require.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			
			// Hash should start with $2a$ (bcrypt prefix)
			assert.True(t, len(hash) > 0 && hash[0] == '$')
			
			// Hash should be different from password
			assert.NotEqual(t, tt.password, hash)
		})
	}
}

func TestPasswordHasher_Hash_Uniqueness(t *testing.T) {
	hasher := NewPasswordHasher(10)
	password := "testPassword123"
	
	// Generate multiple hashes of the same password
	hash1, err1 := hasher.Hash(password)
	hash2, err2 := hasher.Hash(password)
	hash3, err3 := hasher.Hash(password)
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	
	// All hashes should be different (due to random salt)
	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash2, hash3)
	assert.NotEqual(t, hash1, hash3)
	
	// But all should verify against the same password
	assert.NoError(t, hasher.Verify(password, hash1))
	assert.NoError(t, hasher.Verify(password, hash2))
	assert.NoError(t, hasher.Verify(password, hash3))
}

func TestPasswordHasher_Verify(t *testing.T) {
	hasher := NewPasswordHasher(10)
	password := "correctPassword123"
	
	// Hash the password
	hash, err := hasher.Hash(password)
	require.NoError(t, err)
	
	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "wrong password",
			password: "wrongPassword456",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "case sensitive - lowercase",
			password: "correctpassword123",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "case sensitive - uppercase",
			password: "CORRECTPASSWORD123",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "password with extra character",
			password: password + "x",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "password with missing character",
			password: password[:len(password)-1],
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.Verify(tt.password, tt.hash)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewPasswordHasher(10)
	password := "validPassword123"
	
	tests := []struct {
		name string
		hash string
	}{
		{"empty hash", ""},
		{"invalid format", "not-a-bcrypt-hash"},
		{"truncated hash", "$2a$10$"},
		{"random string", "random123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.Verify(password, tt.hash)
			assert.Error(t, err, "Invalid hash should return error")
		})
	}
}

func TestPasswordHasher_DifferentCosts(t *testing.T) {
	costs := []int{4, 10, 12}
	password := "testPassword123"
	
	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hasher := NewPasswordHasher(cost)
			
			hash, err := hasher.Hash(password)
			require.NoError(t, err)
			
			// Verify password works regardless of cost
			assert.NoError(t, hasher.Verify(password, hash))
		})
	}
}

func TestPasswordHasher_RealWorldScenario(t *testing.T) {
	// Simulate real-world usage
	hasher := NewPasswordHasher(12)
	
	// User registers with password
	userPassword := "SecureP@ssw0rd!"
	storedHash, err := hasher.Hash(userPassword)
	require.NoError(t, err)
	
	// User logs in with correct password
	assert.NoError(t, hasher.Verify(userPassword, storedHash),
		"User should be able to login with correct password")
	
	// Attacker tries wrong passwords
	wrongPasswords := []string{
		"wrongpassword",
		"SecureP@ssw0rd", // Missing !
		"secureP@ssw0rd!", // Wrong case
		"",
		"SecureP@ssw0rd!x", // Extra character
	}
	
	for _, wrongPwd := range wrongPasswords {
		assert.Error(t, hasher.Verify(wrongPwd, storedHash),
			"Wrong password '%s' should be rejected", wrongPwd)
	}
}

