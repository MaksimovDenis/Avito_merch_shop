package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const secretKey = "testsecret"

func TestJWTMaker(t *testing.T) {
	maker := NewJWTMaker(secretKey)

	tests := []struct {
		name         string
		id           int64
		username     string
		duration     time.Duration
		shouldExpire bool
		shouldFail   bool
		signingNone  bool
	}{
		{
			name:     "Valid Token",
			id:       123,
			username: "testuser",
			duration: time.Minute,
		},
		{
			name:         "Expired Token",
			id:           456,
			username:     "expiredUser",
			duration:     -time.Minute,
			shouldExpire: true,
			shouldFail:   true,
		},
		{
			name:        "Invalid Signing Method",
			id:          789,
			username:    "invalidSignUser",
			duration:    time.Minute,
			signingNone: true,
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenStr string

			var claims *UserClaims

			var err error

			if tt.signingNone {
				claims, _ = NewUserClaims(tt.id, tt.username, tt.duration)
				token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				tokenStr, _ = token.SignedString(jwt.UnsafeAllowNoneSignatureType)
			} else {
				tokenStr, claims, err = maker.CreateToken(tt.id, tt.username, tt.duration)
				assert.NoError(t, err, "CreateToken should not return an error")
				assert.NotEmpty(t, tokenStr, "Token should not be empty")
				assert.NotNil(t, claims, "Claims should not be nil")
			}

			parsedClaims, err := maker.VerifyToken(tokenStr)

			if tt.shouldFail {
				assert.Error(t, err, "VerifyToken should return an error")
				assert.Nil(t, parsedClaims, "Claims should be nil for invalid tokens")
			} else {
				assert.NoError(t, err, "VerifyToken should not return an error")
				assert.NotNil(t, parsedClaims, "Claims should not be nil")
				assert.Equal(t, tt.id, parsedClaims.ID, "UserID should match")
				assert.Equal(t, tt.username, parsedClaims.UserName, "UserName should match")

				if tt.shouldExpire {
					assert.True(t, parsedClaims.ExpiresAt.Before(time.Now()), "Token should be expired")
				} else {
					assert.WithinDuration(t, time.Now().Add(tt.duration),
						parsedClaims.ExpiresAt.Time, time.Second, "Expiration time should be correct")
				}
			}
		})
	}
}
