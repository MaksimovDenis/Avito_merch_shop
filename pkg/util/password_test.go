package util

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "securepassword", false},
		{"Empty password", "", false},
		{"Long password", "averyverylongpasswordthatexceedstypicalpasswordlength", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(tt.password))
				if err != nil {
					t.Errorf("Hashed password does not match original: %v", err)
				}
			}
		})
	}
}
