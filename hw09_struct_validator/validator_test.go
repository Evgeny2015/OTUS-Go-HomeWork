package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John Doe",
				Age:    30,
				Email:  "johndoe@example.com",
				Role:   UserRole("admin"),
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "12345678901234567890123456789012345",
				Name:   "John Doe",
				Age:    30,
				Email:  "johndoe@example.com",
				Role:   UserRole("admin"),
				Phones: []string{"12345678901"},
			},
			expectedErr: ErrInvalidStringLength,
		},
		{
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			in: App{
				Version: "1.0.0.1",
			},
			expectedErr: ErrInvalidStringLength,
		},
		{
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 201,
				Body: "OK",
			},
			expectedErr: ErrInvalidItem,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			// Place your code here.
			err := Validate(tt.in)

			// If no error is expected
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				return
			}

			// If an error is expected but none occurred
			if err == nil {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			// Check if any of the validation errors match the expected error
			found := false
			for _, vErr := range err.(ValidationErrors) { //nolint:errorlint
				if errors.Is(vErr.Err, tt.expectedErr) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			_ = tt
		})
	}
}

func TestUser(t *testing.T) {
	user := User{
		ID:     "123456789012345678901234567890123456",
		Name:   "John Doe",
		Age:    30,
		Email:  "johndoe@example.com",
		Role:   UserRole("admin"),
		Phones: []string{"12345678901"},
	}

	t.Run("all valid", func(t *testing.T) {
		err := Validate(user)
		require.NoError(t, err)
	})

	t.Run("user id", func(t *testing.T) {
		user.ID = "12345678901234567890123456789"
		err := Validate(user)
		require.ErrorContains(t, err, ErrInvalidStringLength.Error())
	})

	t.Run("user email", func(t *testing.T) {
		user.Email = "john.doe@example.com"
		err := Validate(user)
		require.ErrorContains(t, err, ErrInvalidReqExp.Error())
	})

	t.Run("user role", func(t *testing.T) {
		user.Role = "no"
		err := Validate(user)
		require.ErrorContains(t, err, ErrInvalidItem.Error())
	})

	t.Run("user phones", func(t *testing.T) {
		user.Phones = []string{"12345678901", "1234567890"}
		err := Validate(user)
		fmt.Println(err)
		require.ErrorContains(t, err, ErrInvalidStringLength.Error())
	})

	t.Run("user age", func(t *testing.T) {
		user.Age = 10
		err := Validate(user)
		require.ErrorContains(t, err, ErrValueMin.Error())

		user.Age = 60
		err = Validate(user)
		require.ErrorContains(t, err, ErrValueMax.Error())
	})
}

func TestApp(t *testing.T) {
	app := App{
		Version: "1.0.0",
	}
	t.Run("all valid", func(t *testing.T) {
		err := Validate(app)
		require.NoError(t, err)
	})

	t.Run("app version", func(t *testing.T) {
		app.Version = "1.0.1.0"
		err := Validate(app)
		require.ErrorContains(t, err, ErrInvalidStringLength.Error())
	})
}

func TestResponse(t *testing.T) {
	response := Response{
		Code: 200,
		Body: "OK",
	}

	t.Run("all valid", func(t *testing.T) {
		err := Validate(response)
		require.NoError(t, err)
	})

	t.Run("response code", func(t *testing.T) {
		response.Code = 201
		err := Validate(response)
		require.ErrorContains(t, err, ErrInvalidItem.Error())
	})
}
