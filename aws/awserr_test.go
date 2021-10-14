package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestIsAwsErr(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Code     string
		Message  string
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "nil error code",
			Err:  nil,
			Code: "test",
		},
		{
			Name:    "nil error message",
			Err:     nil,
			Message: "test",
		},
		{
			Name:    "nil error code and message",
			Err:     nil,
			Code:    "test",
			Message: "test",
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name: "other error code",
			Err:  errors.New("test"),
			Code: "test",
		},
		{
			Name:    "other error message",
			Err:     errors.New("test"),
			Message: "test",
		},
		{
			Name:    "other error code and message",
			Err:     errors.New("test"),
			Code:    "test",
			Message: "test",
		},
		{
			Name:     "awserr error matching code and no message",
			Err:      awserr.New("TestCode", "TestMessage", nil),
			Code:     "TestCode",
			Expected: true,
		},
		{
			Name:     "awserr error matching code and matching message exact",
			Err:      awserr.New("TestCode", "TestMessage", nil),
			Code:     "TestCode",
			Message:  "TestMessage",
			Expected: true,
		},
		{
			Name:     "awserr error matching code and matching message contains",
			Err:      awserr.New("TestCode", "TestMessage", nil),
			Code:     "TestCode",
			Message:  "Message",
			Expected: true,
		},
		{
			Name:    "awserr error matching code and non-matching message",
			Err:     awserr.New("TestCode", "TestMessage", nil),
			Code:    "TestCode",
			Message: "NotMatching",
		},
		{
			Name: "awserr error no code",
			Err:  awserr.New("TestCode", "TestMessage", nil),
		},
		{
			Name:    "awserr error no code and matching message exact",
			Err:     awserr.New("TestCode", "TestMessage", nil),
			Message: "TestMessage",
		},
		{
			Name: "awserr error non-matching code",
			Err:  awserr.New("TestCode", "TestMessage", nil),
			Code: "NotMatching",
		},
		{
			Name:    "awserr error non-matching code and message exact",
			Err:     awserr.New("TestCode", "TestMessage", nil),
			Message: "TestMessage",
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name: "wrapped other error code",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
			Code: "test",
		},
		{
			Name:    "wrapped other error message",
			Err:     fmt.Errorf("test: %w", errors.New("test")),
			Message: "test",
		},
		{
			Name:    "wrapped other error code and message",
			Err:     fmt.Errorf("test: %w", errors.New("test")),
			Code:    "test",
			Message: "test",
		},
		{
			Name:     "wrapped awserr error matching code and no message",
			Err:      fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Code:     "TestCode",
			Expected: true,
		},
		{
			Name:     "wrapped awserr error matching code and matching message exact",
			Err:      fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Code:     "TestCode",
			Message:  "TestMessage",
			Expected: true,
		},
		{
			Name:     "wrapped awserr error matching code and matching message contains",
			Err:      fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Code:     "TestCode",
			Message:  "Message",
			Expected: true,
		},
		{
			Name:    "wrapped awserr error matching code and non-matching message",
			Err:     fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Code:    "TestCode",
			Message: "NotMatching",
		},
		{
			Name: "wrapped awserr error no code",
			Err:  fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
		},
		{
			Name:    "wrapped awserr error no code and matching message exact",
			Err:     fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Message: "TestMessage",
		},
		{
			Name: "wrapped awserr error non-matching code",
			Err:  fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Code: "NotMatching",
		},
		{
			Name:    "wrapped awserr error non-matching code and message exact",
			Err:     fmt.Errorf("test: %w", awserr.New("TestCode", "TestMessage", nil)),
			Message: "TestMessage",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := tfawserr.ErrMessageContains(testCase.Err, testCase.Code, testCase.Message)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
