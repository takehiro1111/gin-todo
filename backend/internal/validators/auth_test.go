package validators

import (
	"errors"
	"testing"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthenticateValidator(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "[正常系] Email と Password 両方指定",
			email:    "test@example.com",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "[正常系] Email のみ指定",
			email:    "test@example.com",
			password: "",
			wantErr:  nil,
		},
		{
			name:     "[正常系] Password のみ指定",
			email:    "",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "[異常系] 両方空",
			email:    "",
			password: "",
			wantErr:  appErr.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.email != "" && tt.password != "" {
				err = NewAuthenticateValidator(
					WithEmail(tt.email),
					WithPassword(tt.password),
				)
			} else if tt.email != "" {
				err = NewAuthenticateValidator(
					WithEmail(tt.email),
				)
			} else if tt.password != "" {
				err = NewAuthenticateValidator(
					WithPassword(tt.password),
				)
			} else {
				err = NewAuthenticateValidator()
			}

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestNewAuthenticateValidator_EmailFormat(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "[正常系] 標準的なメールアドレス",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "[正常系] サブドメイン付き",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "[正常系] プラス記号付き",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "[異常系] @なし",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "[異常系] ドメインなし",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "[異常系] TLDなし",
			email:   "user@example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthenticateValidator(WithEmail(tt.email))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestNewAuthenticateValidator_PasswordFormat(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "[正常系] 8文字のパスワード",
			password: "pass1234",
			wantErr:  false,
		},
		{
			name:     "[正常系] 24文字のパスワード",
			password: "password12345678901234ab",
			wantErr:  false,
		},
		{
			name:     "[正常系] 記号を含むパスワード",
			password: "pass.word-123",
			wantErr:  false,
		},
		{
			name:     "[異常系] 7文字（短すぎる）",
			password: "pass123",
			wantErr:  true,
		},
		{
			name:     "[異常系] 25文字（長すぎる）",
			password: "password123456789012345abc",
			wantErr:  true,
		},
		{
			name:     "[異常系] 許可されていない記号",
			password: "pass@word123",
			wantErr:  true,
		},
		{
			name:     "[異常系] スペースを含む",
			password: "pass word123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthenticateValidator(WithPassword(tt.password))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestWithEmail(t *testing.T) {
	cfg := &authValidatorConfig{}
	opt := WithEmail("test@example.com")
	opt(cfg)

	assert.Equal(t, "test@example.com", cfg.email)
}

func TestWithPassword(t *testing.T) {
	cfg := &authValidatorConfig{}
	opt := WithPassword("password123")
	opt(cfg)

	assert.Equal(t, "password123", cfg.password)
}
