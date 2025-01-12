package auth_test

import (
	"testing"

	"github.com/0x2e/fusion/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	for _, tt := range []struct {
		explanation string
		input       string
		wantErr     error
	}{
		{
			explanation: "valid password succeeds",
			input:       "mypassword",
			wantErr:     nil,
		},
		{
			explanation: "empty password returns ErrPasswordTooShort",
			input:       "",
			wantErr:     auth.ErrPasswordTooShort,
		},
	} {
		t.Run(tt.explanation, func(t *testing.T) {
			got, err := auth.HashPassword(tt.input)
			require.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.NotEmpty(t, got.Bytes())
			}
		})
	}
}

func TestHashedPasswordEquals(t *testing.T) {
	for _, tt := range []struct {
		explanation     string
		hashedPassword1 auth.HashedPassword
		hashedPassword2 auth.HashedPassword
		want            bool
	}{
		{
			explanation:     "same passwords match",
			hashedPassword1: mustHashPassword("password1"),
			hashedPassword2: mustHashPassword("password1"),
			want:            true,
		},
		{
			explanation:     "different passwords don't match",
			hashedPassword1: mustHashPassword("password1"),
			hashedPassword2: mustHashPassword("password2"),
			want:            false,
		},
	} {
		t.Run(tt.explanation, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.hashedPassword1.Equals(tt.hashedPassword2))
		})
	}
}

func mustHashPassword(password string) auth.HashedPassword {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}
