// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apicall_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/aws/smithy-go/middleware"
	"github.com/hashicorp/terraform-provider-aws/internal/conns/apicall"
)

// TestEndToEnd_PinpointClient drives a real pinpoint.Client built via
// NewFromConfig and verifies the recorder captures the operation.
//
// It guards against future smithy-go or codegen changes that would invalidate
// the assumption that Initialize.After sees a fully-populated service ctx.
func TestEndToEnd_PinpointClient(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, `{"ApplicationResponse":{"Id":"x","Arn":"arn:aws:mobiletargeting:us-east-1:000000000000:apps/x","Name":"x"}}`); err != nil {
			t.Errorf("write response: %s", err)
		}
	}))
	t.Cleanup(server.Close)

	rec := apicall.NewRecorder()
	ctx := apicall.NewContext(context.Background(), rec)

	cfg := aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(server.URL),
		Credentials:  credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		APIOptions:   []func(*middleware.Stack) error{apicall.Middleware()},
	}

	client := pinpoint.NewFromConfig(cfg)
	input := pinpoint.GetAppInput{ApplicationId: aws.String("x")}
	if _, err := client.GetApp(ctx, &input); err != nil {
		t.Fatalf("GetApp: %v", err)
	}

	calls := rec.Calls()
	if len(calls) != 1 {
		t.Fatalf("len(Calls()) = %d, want 1; calls=%+v", len(calls), calls)
	}
	if calls[0].Service != "Pinpoint" || calls[0].Operation != "GetApp" || calls[0].Err != nil {
		t.Errorf("Calls[0] = %+v, want Pinpoint.GetApp with no error", calls[0])
	}
	if calls[0].Duration <= 0 {
		t.Errorf("Duration = %s, want > 0", calls[0].Duration)
	}
}
