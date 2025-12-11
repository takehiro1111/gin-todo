package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	hasher := NewBcryptHasher()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "[正常系] 通常のパスワード",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "[正常系] 空のパスワード",
			password: "",
			wantErr:  false,
		},
		{
			name:     "[正常系] 長いパスワード",
			password: "verylongpassword1234567890abcdefghijklmnop",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			assert.True(t, len(hash) >= 4 && hash[:4] == "$2a$", "bcryptのハッシュは$2a$で始まる")
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	hasher := NewBcryptHasher()
	password := "samepassword"

	hash1, err1 := hasher.HashPassword(password)
	hash2, err2 := hasher.HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "bcryptは毎回異なるソルトを使うため、同じパスワードでも異なるハッシュになる")
}

func TestCheckPassword(t *testing.T) {
	hasher := NewBcryptHasher()

	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "[正常系] 正しいパスワード",
			hashedPassword: string(hashedPassword),
			password:       password,
			wantErr:        false,
		},
		{
			name:           "[異常系] 間違ったパスワード",
			hashedPassword: string(hashedPassword),
			password:       "wrongpassword",
			wantErr:        true,
		},
		{
			name:           "[異常系] 空のパスワード",
			hashedPassword: string(hashedPassword),
			password:       "",
			wantErr:        true,
		},
		{
			name:           "[異常系] 無効なハッシュ",
			hashedPassword: "invalidhash",
			password:       password,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.CheckPassword(tt.hashedPassword, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestHashAndCheckPassword_Integration(t *testing.T) {
	hasher := NewBcryptHasher()
	password := "integrationtest123"

	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)

	err = hasher.CheckPassword(hash, password)
	assert.NoError(t, err, "正しいパスワードで検証")

	err = hasher.CheckPassword(hash, "wrongpassword")
	assert.Error(t, err, "間違ったパスワードで検証")
}
