package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"

	appErr "github.com/takehiro1111/gin-todo/backend/internal/errors"
)

type authValidatorConfig struct {
	email    string
	password string
}

type AuthValidatorOption func(*authValidatorConfig)

func WithEmail(email string) AuthValidatorOption {
	return func(c *authValidatorConfig) {
		c.email = email
	}
}

func WithPassword(password string) AuthValidatorOption {
	return func(c *authValidatorConfig) {
		c.password = password
	}
}

func emailValidate(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true
	}
	match, _ := regexp.MatchString(".+@.+\\..+", email)
	return match
}

func passwordValidate(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9.?/!-]{8,24}$", password)
	return match
}

type authValidateTarget struct {
	Email    string `validate:"omitempty,emailCustom"`
	Password string `validate:"omitempty,passwordCustom"`
}

func NewAuthenticateValidator(opts ...AuthValidatorOption) error {
	// opts = [
	//   func(c) { c.email = "test@example.com" },
	//   func(c) { c.password = "password123" },
	// ]

	cfg := &authValidatorConfig{}
	for _, opt := range opts {
		// WithEmail or WithPasswordがここで初めて実行され、configが設定される。
		opt(cfg)
	}

	if cfg.email == "" && cfg.password == "" {
		return appErr.ErrInvalidInput
	}

	validate := validator.New()
	validate.RegisterValidation("emailCustom", emailValidate)
	validate.RegisterValidation("passwordCustom", passwordValidate)

	target := authValidateTarget{
		Email:    cfg.email,
		Password: cfg.password,
	}

	if err := validate.Struct(target); err != nil {
		return appErr.ErrInvalidInput
	}

	return nil
}
