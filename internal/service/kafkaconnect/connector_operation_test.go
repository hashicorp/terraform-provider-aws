// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafkaconnect/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestWaitConnectorOperationCompleted(t *testing.T) {
	t.Parallel()

	const operationARN = "arn:aws:kafkaconnect:us-west-2:123456789012:connector-operation/test/operation"

	testCases := []struct {
		name          string
		terminalState awstypes.ConnectorOperationState
		expectError   bool
	}{
		{
			name:          "update complete",
			terminalState: awstypes.ConnectorOperationStateUpdateComplete,
		},
		{
			name:          "update failed",
			terminalState: awstypes.ConnectorOperationStateUpdateFailed,
			expectError:   true,
		},
		{
			name:          "rollback failed",
			terminalState: awstypes.ConnectorOperationStateRollbackFailed,
			expectError:   true,
		},
		{
			name:          "rollback complete",
			terminalState: awstypes.ConnectorOperationStateRollbackComplete,
			expectError:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			conn := newTestKafkaConnectClient(t, operationARN, []awstypes.ConnectorOperationState{
				awstypes.ConnectorOperationStatePending,
				awstypes.ConnectorOperationStateUpdateInProgress,
				testCase.terminalState,
			})

			output, err := waitConnectorOperationCompleted(context.Background(), conn, operationARN, 3*time.Second)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got none")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %s", err)
			}

			if output == nil {
				t.Fatal("expected non-nil operation output, got nil")
			} else if output.ConnectorOperationState != testCase.terminalState {
				t.Fatalf("expected operation state %s, got: %s", testCase.terminalState, output.ConnectorOperationState)
			}

			if err != nil {
				errString := err.Error()
				expectedContains := []string{operationARN}
				expectedContains = append(expectedContains, enum.Slice(testCase.terminalState)...)
				expectedContains = append(expectedContains, enum.Slice(awstypes.ConnectorOperationTypeUpdateConnectorConfiguration)...)
				expectedContains = append(expectedContains,
					"InvalidConnectorConfiguration",
					"connector configuration rejected",
					"VALIDATE_UPDATE=FAILED",
				)

				for _, expected := range expectedContains {
					if !strings.Contains(errString, expected) {
						t.Fatalf("expected error to contain %q, got: %s", expected, errString)
					}
				}
			}
		})
	}
}

func newTestKafkaConnectClient(t *testing.T, operationARN string, states []awstypes.ConnectorOperationState) *kafkaconnect.Client {
	t.Helper()

	var index int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if index >= len(states) {
			index = len(states) - 1
		}

		state := states[index]
		index++

		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"connectorOperationArn":   operationARN,
			"connectorOperationState": state,
			"connectorOperationType":  awstypes.ConnectorOperationTypeUpdateConnectorConfiguration,
			"errorInfo": map[string]string{
				"code":            "InvalidConnectorConfiguration",
				names.AttrMessage: "connector configuration rejected",
			},
			"operationSteps": []map[string]any{
				{
					"stepState": awstypes.ConnectorOperationStepStateFailed,
					"stepType":  awstypes.ConnectorOperationStepTypeValidateUpdate,
				},
			},
		})
	}))
	t.Cleanup(server.Close)

	return kafkaconnect.New(kafkaconnect.Options{
		Region:       "us-west-2", //lintignore:AWSAT003
		BaseEndpoint: aws.String(server.URL),
		Credentials:  credentials.NewStaticCredentialsProvider("AKID", "SECRET", "SESSION"),
	})
}
