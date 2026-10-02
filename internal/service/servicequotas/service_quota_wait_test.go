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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
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

func TestRefreshServiceQuotaRequest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		wait      bool
		status    awstypes.RequestStatus
		applied   float64
		wantID    string
		wantValue float64
	}{
		{name: "pendingWait", wait: true, status: awstypes.RequestStatusPending, applied: 10, wantID: "request-id", wantValue: 10},
		{name: "openedWait", wait: true, status: awstypes.RequestStatusCaseOpened, applied: 10, wantID: "request-id", wantValue: 10},
		{name: "approvedLag", wait: true, status: awstypes.RequestStatusApproved, applied: 10, wantID: "request-id", wantValue: 10},
		{name: "closedLag", wait: true, status: awstypes.RequestStatusCaseClosed, applied: 10, wantID: "request-id", wantValue: 10},
		{name: "enacted", wait: true, status: awstypes.RequestStatusApproved, applied: 20, wantValue: 20},
		{name: "denied", wait: true, status: awstypes.RequestStatusDenied, applied: 10, wantValue: 10},
		{name: "notApproved", wait: true, status: awstypes.RequestStatusNotApproved, applied: 10, wantValue: 10},
		{name: "invalid", wait: true, status: awstypes.RequestStatusInvalidRequest, applied: 10, wantValue: 10},
		{name: "legacyPending", status: awstypes.RequestStatusPending, applied: 10, wantID: "request-id", wantValue: 20},
		{name: "legacyOpened", status: awstypes.RequestStatusCaseOpened, applied: 10, wantID: "request-id", wantValue: 20},
		{name: "legacyApproved", status: awstypes.RequestStatusApproved, applied: 10, wantValue: 10},
		{name: "legacyClosed", status: awstypes.RequestStatusCaseClosed, applied: 10, wantValue: 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := schema.TestResourceDataRaw(t, resourceServiceQuota().SchemaMap(), map[string]any{
				"service_code": "service", "quota_code": "quota", names.AttrValue: tc.applied,
				"wait_for_fulfillment": tc.wait,
			})
			if err := d.Set("request_id", "request-id"); err != nil {
				t.Fatal(err)
			}
			request := &awstypes.RequestedServiceQuotaChange{Status: tc.status, DesiredValue: aws.Float64(20)}
			// Repeated refreshes must not mistake the previous applied value for the target.
			refreshServiceQuotaRequest(d, request)
			refreshServiceQuotaRequest(d, request)
			if got := d.Get(names.AttrValue).(float64); got != tc.wantValue {
				t.Fatalf("expected value %f, got %f", tc.wantValue, got)
			}
			if got := d.Get("request_id").(string); got != tc.wantID {
				t.Fatalf("expected request ID %q, got %q", tc.wantID, got)
			}
			if got := d.Get("request_status").(string); got != string(tc.status) {
				t.Fatalf("expected status %q, got %q", tc.status, got)
			}
		})
	}
}

func TestIncreaseServiceQuotaTimeoutRecovery(t *testing.T) {
	t.Parallel()

	for _, status := range []awstypes.RequestStatus{awstypes.RequestStatusPending, awstypes.RequestStatusApproved, awstypes.RequestStatusCaseClosed} {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()

			d := schema.TestResourceDataRaw(t, resourceServiceQuota().SchemaMap(), map[string]any{
				"service_code": "service", "quota_code": "quota", names.AttrValue: 100.0,
				"wait_for_fulfillment": true,
			})
			client := &fakeServiceQuotaWaitClient{blockQuota: true}
			err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, 25*time.Millisecond)
			if err == nil {
				t.Fatal("expected fulfillment timeout")
			}
			if d.Id() != "service/quota" || d.Get("request_id") != "request-id" {
				t.Fatalf("timeout lost resource or request ID: %q, %v", d.Id(), d.Get("request_id"))
			}
			client.requestedQuota.Status = status
			if err := d.Set(names.AttrValue, 10.0); err != nil {
				t.Fatal(err)
			}
			refreshServiceQuotaRequest(d, client.requestedQuota)
			refreshServiceQuotaRequest(d, client.requestedQuota)
			if d.Get(names.AttrValue) != 10.0 || d.Get("request_id") != "request-id" {
				t.Fatal("refresh masked the unfulfilled quota or lost the request")
			}
			client.blockQuota = false
			client.quota = &awstypes.ServiceQuota{Value: aws.Float64(100)}
			if err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, time.Second); err != nil {
				t.Fatal(err)
			}
			if client.submissions != 1 || d.Get("request_id") != "" {
				t.Fatal("resume submitted a duplicate request or retained a fulfilled request")
			}
		})
	}
}

func TestIncreaseServiceQuotaAdoption(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		status          awstypes.RequestStatus
		desired         float64
		saved           bool
		race            bool
		wantError       string
		wantSubmissions int
	}{
		{name: "createOpen", status: awstypes.RequestStatusPending, desired: 200},
		{name: "createApproved", status: awstypes.RequestStatusApproved, desired: 200},
		{name: "createClosed", status: awstypes.RequestStatusCaseClosed, desired: 200},
		{name: "updateApproved", status: awstypes.RequestStatusApproved, desired: 200, saved: true},
		{name: "updateClosed", status: awstypes.RequestStatusCaseClosed, desired: 200, saved: true},
		{name: "smallerOpenRequest", status: awstypes.RequestStatusPending, desired: 80, wantError: "does not request"},
		{name: "savedSmallerOpenRequest", status: awstypes.RequestStatusPending, desired: 80, saved: true, wantError: "does not request"},
		{name: "retireDenied", status: awstypes.RequestStatusDenied, desired: 100, saved: true, wantSubmissions: 1},
		{name: "retireNotApproved", status: awstypes.RequestStatusNotApproved, desired: 100, saved: true, wantSubmissions: 1},
		{name: "retireInvalid", status: awstypes.RequestStatusInvalidRequest, desired: 100, saved: true, wantSubmissions: 1},
		{name: "retireSmallerApproved", status: awstypes.RequestStatusApproved, desired: 80, saved: true, wantSubmissions: 1},
		{name: "concurrentRequest", status: awstypes.RequestStatusPending, desired: 200, race: true, wantSubmissions: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := schema.TestResourceDataRaw(t, resourceServiceQuota().SchemaMap(), map[string]any{
				"service_code": "service", "quota_code": "quota", names.AttrValue: 100.0,
				"wait_for_fulfillment": true,
			})
			request := &awstypes.RequestedServiceQuotaChange{
				Id: aws.String("existing"), Status: tc.status, DesiredValue: aws.Float64(tc.desired),
				ServiceCode: aws.String("service"), QuotaCode: aws.String("quota"),
			}
			page := &servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
				RequestedQuotas: []awstypes.RequestedServiceQuotaChange{*request},
			}
			client := &fakeServiceQuotaWaitClient{
				requestedQuota: request, quota: &awstypes.ServiceQuota{Value: aws.Float64(150)},
				historyPages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{page},
			}
			if tc.saved {
				d.SetId("service/quota")
				if err := d.Set("request_id", "existing"); err != nil {
					t.Fatal(err)
				}
			}
			if tc.race {
				client.historyPages = []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{{}, page}
				client.submitErr = &awstypes.ResourceAlreadyExistsException{Message: aws.String("Only one open service quota increase request is allowed per quota")}
			}
			err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, time.Second)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
				}
			} else if err != nil || d.Id() != "service/quota" {
				t.Fatalf("expected successful adoption, got %q, %v", d.Id(), err)
			}
			if client.submissions != tc.wantSubmissions {
				t.Fatalf("expected %d submissions, got %d", tc.wantSubmissions, client.submissions)
			}
		})
	}
}

func TestIncreaseServiceQuotaWithoutWaiting(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceServiceQuota().SchemaMap(), map[string]any{
		"service_code": "service", "quota_code": "quota", names.AttrValue: 100.0,
	})
	client := &fakeServiceQuotaWaitClient{}
	if err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, time.Second); err != nil {
		t.Fatal(err)
	}
	if client.submissions != 1 || len(client.historyTokens) != 0 || len(client.quotaMarkers) != 0 {
		t.Fatal("non-waiting increase searched history or polled fulfillment")
	}
}

func TestIncreaseServiceQuotaMalformedResponse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value *float64
	}{
		{name: "missingValue"},
		{name: "smallerValue", value: aws.Float64(80)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := schema.TestResourceDataRaw(t, resourceServiceQuota().SchemaMap(), map[string]any{
				"service_code": "service", "quota_code": "quota", names.AttrValue: 100.0,
				"wait_for_fulfillment": true,
			})
			client := &fakeServiceQuotaWaitClient{
				submittedQuota: &awstypes.RequestedServiceQuotaChange{
					Id: aws.String("accepted"), DesiredValue: tc.value, Status: awstypes.RequestStatusPending,
				},
			}
			if err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, time.Second); err == nil {
				t.Fatal("expected malformed response error")
			}
			if d.Id() != "service/quota" || d.Get("request_id") != "accepted" {
				t.Fatal("accepted request association was lost")
			}
			client.requestedQuota.Status = awstypes.RequestStatusApproved
			if tc.value == nil {
				if err := increaseServiceQuota(context.Background(), client, d, "service", "quota", 100, time.Second); err == nil || client.submissions != 1 {
					t.Fatal("retry must retain a known request with missing metadata, not submit another")
				}
			}
		})
	}
}

func TestServiceQuotaValueDiff(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		wait bool
		old  string
		new  string
		want bool
	}{
		{name: "oversatisfied", wait: true, old: "150", new: "100", want: true},
		{name: "fulfilled", wait: true, old: "100", new: "100", want: true},
		{name: "unfulfilled", wait: true, old: "10", new: "100"},
		{name: "legacy", old: "150", new: "100"},
		{name: "unknown", wait: true, old: "150", new: "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := resourceServiceQuota()
			d := schema.TestResourceDataRaw(t, r.SchemaMap(), map[string]any{
				"service_code": "service", "quota_code": "quota", names.AttrValue: 100.0,
				"wait_for_fulfillment": tc.wait,
			})
			if got := r.SchemaMap()[names.AttrValue].DiffSuppressFunc("value", tc.old, tc.new, d); got != tc.want {
				t.Fatalf("expected suppression %v, got %v", tc.want, got)
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

func TestFindServiceQuotaRequestByQuota(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		pages     []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput
		wantID    string
		wantError string
	}{
		{name: "empty", wantError: "notFound"},
		{name: "paginatedOpen", pages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
			{NextToken: aws.String("next")},
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Id: aws.String("open"), Status: awstypes.RequestStatusPending}}},
		}, wantID: "open"},
		{name: "newestApproved", pages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Id: aws.String("old"), Status: awstypes.RequestStatusApproved, DesiredValue: aws.Float64(100), Created: aws.Time(time.Unix(1, 0))}}, NextToken: aws.String("next")},
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Id: aws.String("new"), Status: awstypes.RequestStatusApproved, DesiredValue: aws.Float64(100), Created: aws.Time(time.Unix(2, 0))}}},
		}, wantID: "new"},
		{name: "openBeforeHistorical", pages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Id: aws.String("approved"), Status: awstypes.RequestStatusApproved, DesiredValue: aws.Float64(100)}}, NextToken: aws.String("next")},
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Id: aws.String("open"), Status: awstypes.RequestStatusCaseOpened}}},
		}, wantID: "open"},
		{name: "smallerApprovedIgnored", pages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
			{RequestedQuotas: []awstypes.RequestedServiceQuotaChange{{Status: awstypes.RequestStatusApproved, DesiredValue: aws.Float64(80)}}},
		}, wantError: "notFound"},
		{name: "repeatedToken", pages: []*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{
			{NextToken: aws.String("next")}, {NextToken: aws.String("next")},
		}, wantError: "repeated pagination token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeServiceQuotaWaitClient{historyPages: tc.pages}
			request, err := findServiceQuotaRequestByQuota(context.Background(), client, "service", "quota", 100)
			if tc.wantError == "notFound" {
				if !retry.NotFound(err) {
					t.Fatalf("expected not-found error, got %v", err)
				}
			} else if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected %q error, got %v", tc.wantError, err)
				}
			} else if err != nil || request == nil || aws.ToString(request.Id) != tc.wantID {
				t.Fatalf("expected request %q, got %v, %v", tc.wantID, request, err)
			}
			if len(client.historyTokens) > 1 && client.historyTokens[1] != "next" {
				t.Fatal("history pagination token was not forwarded")
			}
		})
	}
}

func TestFindServiceQuotaRequestByQuotaCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &fakeServiceQuotaWaitClient{}
	_, err := findServiceQuotaRequestByQuota(ctx, client, "service", "quota", 100)
	if !errors.Is(err, context.Canceled) || len(client.historyTokens) != 0 {
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
