// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

type serviceQuotaWaitContextKey string

type fakeServiceQuotaWaitClient struct {
	requestedQuota *awstypes.RequestedServiceQuotaChange
	quota          *awstypes.ServiceQuota
	quotaErr       error
	requestErr     error
	submitErr      error
	submittedQuota *awstypes.RequestedServiceQuotaChange
	historyPages   []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput
	historyTokens  []string
	requestMarkers []any
	quotaMarkers   []any
	quotaDeadlines []time.Time
	requestIDs     []string
	submissions    int
	blockQuota     bool
}

func (f *fakeServiceQuotaWaitClient) GetRequestedServiceQuotaChange(ctx context.Context, input *servicequotas.GetRequestedServiceQuotaChangeInput, _ ...func(*servicequotas.Options)) (*servicequotas.GetRequestedServiceQuotaChangeOutput, error) {
	f.requestMarkers = append(f.requestMarkers, ctx.Value(serviceQuotaWaitContextKey("refresh")))
	f.requestIDs = append(f.requestIDs, aws.ToString(input.RequestId))
	if f.requestErr != nil {
		return nil, f.requestErr
	}
	return &servicequotas.GetRequestedServiceQuotaChangeOutput{RequestedQuota: f.requestedQuota}, nil
}

func (f *fakeServiceQuotaWaitClient) GetServiceQuota(ctx context.Context, _ *servicequotas.GetServiceQuotaInput, _ ...func(*servicequotas.Options)) (*servicequotas.GetServiceQuotaOutput, error) {
	f.quotaMarkers = append(f.quotaMarkers, ctx.Value(serviceQuotaWaitContextKey("refresh")))
	if deadline, ok := ctx.Deadline(); ok {
		f.quotaDeadlines = append(f.quotaDeadlines, deadline)
	}
	if f.blockQuota {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if f.quotaErr != nil {
		return nil, f.quotaErr
	}
	return &servicequotas.GetServiceQuotaOutput{Quota: f.quota}, nil
}

func (f *fakeServiceQuotaWaitClient) ListRequestedServiceQuotaChangeHistoryByQuota(ctx context.Context, input *servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput, _ ...func(*servicequotas.Options)) (*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	page := len(f.historyTokens)
	f.historyTokens = append(f.historyTokens, aws.ToString(input.NextToken))
	if page >= len(f.historyPages) {
		return &servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{}, nil
	}
	return f.historyPages[page], nil
}

func (f *fakeServiceQuotaWaitClient) RequestServiceQuotaIncrease(ctx context.Context, input *servicequotas.RequestServiceQuotaIncreaseInput, _ ...func(*servicequotas.Options)) (*servicequotas.RequestServiceQuotaIncreaseOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.submissions++
	if f.submitErr != nil {
		return nil, f.submitErr
	}
	request := f.submittedQuota
	if request == nil {
		request = &awstypes.RequestedServiceQuotaChange{
			Id: aws.String("request-id"), ServiceCode: input.ServiceCode,
			QuotaCode: input.QuotaCode, DesiredValue: input.DesiredValue,
			Status: awstypes.RequestStatusPending,
		}
	}
	f.requestedQuota = request
	return &servicequotas.RequestServiceQuotaIncreaseOutput{RequestedQuota: request}, nil
}

func TestStatusServiceQuotaFulfillment(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		status     awstypes.RequestStatus
		applied    float64
		requestErr error
		quotaErr   error
		wantState  string
		wantError  string
	}{
		{name: "pending", status: awstypes.RequestStatusPending, applied: 10, wantState: "pending"},
		{name: "caseOpened", status: awstypes.RequestStatusCaseOpened, applied: 10, wantState: "pending"},
		{name: "approvedNotApplied", status: awstypes.RequestStatusApproved, applied: 10, wantState: "pending"},
		{name: "partiallyApplied", status: awstypes.RequestStatusApproved, applied: 80, wantState: "pending"},
		{name: "closedNotApplied", status: awstypes.RequestStatusCaseClosed, applied: 10, wantState: "pending"},
		{name: "closedApplied", status: awstypes.RequestStatusCaseClosed, applied: 100, wantState: "updated"},
		{name: "pendingAlreadyApplied", status: awstypes.RequestStatusPending, applied: 100, wantState: "updated"},
		{name: "largerRequest", status: awstypes.RequestStatusApproved, applied: 150, wantState: "updated"},
		{name: "denied", status: awstypes.RequestStatusDenied, applied: 10, wantError: "denied"},
		{name: "notApproved", status: awstypes.RequestStatusNotApproved, applied: 10, wantError: "not approved"},
		{name: "invalid", status: awstypes.RequestStatusInvalidRequest, applied: 10, wantError: "invalid"},
		{name: "unknown", status: awstypes.RequestStatus("UNKNOWN"), applied: 10, wantError: "unexpected quota request status"},
		{name: "requestMissing", applied: 10, requestErr: &awstypes.NoSuchResourceException{}, wantState: "pending"},
		{name: "appliedDespiteMissingRequest", applied: 100, requestErr: &awstypes.NoSuchResourceException{}, wantState: "updated"},
		{name: "quotaMissing", status: awstypes.RequestStatusPending, quotaErr: &awstypes.NoSuchResourceException{}, wantState: "pending"},
		{name: "quotaReadError", quotaErr: errors.New("quota read failed"), wantError: "quota read failed"},
		{name: "requestReadError", applied: 10, requestErr: errors.New("request read failed"), wantError: "request read failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeServiceQuotaWaitClient{
				requestedQuota: &awstypes.RequestedServiceQuotaChange{Status: tc.status, DesiredValue: aws.Float64(200)},
				quota:          &awstypes.ServiceQuota{Value: aws.Float64(tc.applied)},
				quotaErr:       tc.quotaErr, requestErr: tc.requestErr,
			}
			ctx := context.WithValue(context.Background(), serviceQuotaWaitContextKey("refresh"), "marker")
			_, state, err := statusServiceQuotaFulfillment(client, "service", "quota", "request-id", 100)(ctx)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
				}
			} else if err != nil || state != tc.wantState {
				t.Fatalf("expected state %q without error, got %q, %v", tc.wantState, state, err)
			}
			if client.quotaMarkers[0] != "marker" {
				t.Fatal("quota finder did not receive the refresh context")
			}
			if state == "updated" && len(client.requestIDs) != 0 {
				t.Fatal("satisfied quota should not depend on request status")
			}
			if len(client.requestMarkers) > 0 && client.requestMarkers[0] != "marker" {
				t.Fatal("request finder did not receive the refresh context")
			}
		})
	}
}

func TestWaitServiceQuotaFulfilledDeadline(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		parentTimeout time.Duration
		waitTimeout   time.Duration
	}{
		{name: "waitDeadline", parentTimeout: time.Second, waitTimeout: 25 * time.Millisecond},
		{name: "parentDeadline", parentTimeout: 25 * time.Millisecond, waitTimeout: time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), tc.parentTimeout)
			defer cancel()
			parentDeadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("missing parent deadline")
			}
			client := &fakeServiceQuotaWaitClient{blockQuota: true}
			start := time.Now()
			if err := waitServiceQuotaFulfilled(ctx, client, "service", "quota", "request-id", 100, tc.waitTimeout); err == nil {
				t.Fatal("expected timeout")
			}
			if len(client.quotaDeadlines) != 1 || client.quotaDeadlines[0].After(parentDeadline) || client.quotaDeadlines[0].After(start.Add(tc.waitTimeout+100*time.Millisecond)) {
				t.Fatalf("poll did not honor the shared deadline: %v", client.quotaDeadlines)
			}
		})
	}
}

func TestWaitServiceQuotaFulfilledCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &fakeServiceQuotaWaitClient{}
	err := waitServiceQuotaFulfilled(ctx, client, "service", "quota", "request-id", 100, time.Second)
	if !errors.Is(err, context.Canceled) || len(client.quotaMarkers) != 0 {
		t.Fatalf("expected cancellation without API calls, got %v", err)
	}
}

func TestFindServiceQuotaReturnsProviderNotFound(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{quotaErr: &awstypes.NoSuchResourceException{}}
	_, err := findServiceQuota(context.Background(), client, &servicequotas.GetServiceQuotaInput{
		QuotaCode: aws.String("quota"), ServiceCode: aws.String("service"),
	})
	if !retry.NotFound(err) {
		t.Fatalf("expected provider not-found error, got %T: %v", err, err)
	}
}
