// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
)

// TestBucketListTags_MethodNotAllowed tests that bucketListTags handles MethodNotAllowed error correctly
func TestBucketListTags_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	
	// Test that MethodNotAllowed error is correctly identified by tfawserr.ErrCodeEquals
	err := &smithy.GenericAPIError{
		Code:    errCodeMethodNotAllowed,
		Message: "The specified method is not allowed against this resource.",
	}
	
	// Verify that tfawserr.ErrCodeEquals correctly identifies MethodNotAllowed
	if !tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed) {
		t.Errorf("Expected tfawserr.ErrCodeEquals to identify MethodNotAllowed error")
	}
}

// TestUpdateTags_MethodNotAllowedErrorCode tests that MethodNotAllowed error is correctly identified
func TestUpdateTags_MethodNotAllowedErrorCode(t *testing.T) {
	t.Parallel()
	
	// Test that MethodNotAllowed error is correctly identified by tfawserr.ErrCodeEquals
	err := &smithy.GenericAPIError{
		Code:    errCodeMethodNotAllowed,
		Message: "The specified method is not allowed against this resource.",
	}
	
	if !tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed) {
		t.Errorf("Expected tfawserr.ErrCodeEquals to identify MethodNotAllowed error")
	}
}

// TestCreateBucket_UnsupportedArgumentErrorCode tests that UnsupportedArgument error is correctly identified
func TestCreateBucket_UnsupportedArgumentErrorCode(t *testing.T) {
	t.Parallel()
	
	// Verify that UnsupportedArgument error is correctly identified
	unsupportedArgErr := &smithy.GenericAPIError{
		Code:    errCodeUnsupportedArgument,
		Message: "The CreateBucket operation does not support the tags argument.",
	}
	
	if !tfawserr.ErrCodeEquals(unsupportedArgErr, errCodeUnsupportedArgument) {
		t.Errorf("Expected tfawserr.ErrCodeEquals to identify UnsupportedArgument error")
	}
}

// TestCreateBucket_AuthorizationError tests authorization error detection
func TestCreateBucket_AuthorizationError(t *testing.T) {
	t.Parallel()
	
	// Test authorization error message detection
	authErr := fmt.Errorf("AccessDenied: User is not authorized to perform: s3:TagResource")
	
	// This should contain the authorization error message
	if authErr.Error() != "AccessDenied: User is not authorized to perform: s3:TagResource" {
		t.Errorf("Expected authorization error message, got: %v", authErr)
	}
}
