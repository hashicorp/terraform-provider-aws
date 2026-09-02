// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

type serviceQuotaWaitContextKey string

type fakeServiceQuotaWaitClient struct {
	requestedQuota    *awstypes.RequestedServiceQuotaChange
	quota             *awstypes.ServiceQuota
	requestedQuotas   []awstypes.RequestedServiceQuotaChange
	quotaErr          error
	requestedContexts []context.Context
	quotaContexts     []context.Context
	historyContexts   []context.Context
}

func (f *fakeServiceQuotaWaitClient) GetRequestedServiceQuotaChange(ctx context.Context, _ *servicequotas.GetRequestedServiceQuotaChangeInput, _ ...func(*servicequotas.Options)) (*servicequotas.GetRequestedServiceQuotaChangeOutput, error) {
	f.requestedContexts = append(f.requestedContexts, ctx)

	return &servicequotas.GetRequestedServiceQuotaChangeOutput{RequestedQuota: f.requestedQuota}, nil
}

func (f *fakeServiceQuotaWaitClient) GetServiceQuota(ctx context.Context, _ *servicequotas.GetServiceQuotaInput, _ ...func(*servicequotas.Options)) (*servicequotas.GetServiceQuotaOutput, error) {
	f.quotaContexts = append(f.quotaContexts, ctx)
	if f.quotaErr != nil {
		return nil, f.quotaErr
	}

	return &servicequotas.GetServiceQuotaOutput{Quota: f.quota}, nil
}

func (f *fakeServiceQuotaWaitClient) ListRequestedServiceQuotaChangeHistoryByQuota(ctx context.Context, _ *servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput, _ ...func(*servicequotas.Options)) (*servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput, error) {
	f.historyContexts = append(f.historyContexts, ctx)

	return &servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaOutput{RequestedQuotas: f.requestedQuotas}, nil
}

func TestStatusServiceQuotaRequestUsesRefreshContext(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{
		requestedQuota: &awstypes.RequestedServiceQuotaChange{
			Id:     aws.String("request-id"),
			Status: awstypes.RequestStatusPending,
		},
	}
	refreshContext := context.WithValue(context.Background(), serviceQuotaWaitContextKey("request"), "refresh")

	_, state, err := statusServiceQuotaRequest(client, "request-id")(refreshContext)

	if err != nil {
		t.Fatalf("statusServiceQuotaRequest returned an error: %s", err)
	}
	if state != string(awstypes.RequestStatusPending) {
		t.Fatalf("expected pending state, got %q", state)
	}
	if got := client.requestedContexts[0].Value(serviceQuotaWaitContextKey("request")); got != "refresh" {
		t.Fatalf("expected refresh context to reach AWS finder, got %v", got)
	}
}

func TestStatusServiceQuotaValueUsesRefreshContext(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{
		quota: &awstypes.ServiceQuota{Value: aws.Float64(10)},
	}
	refreshContext := context.WithValue(context.Background(), serviceQuotaWaitContextKey("quota"), "refresh")

	_, state, err := statusServiceQuotaValue(client, "service", "quota", 20)(refreshContext)

	if err != nil {
		t.Fatalf("statusServiceQuotaValue returned an error: %s", err)
	}
	if state != "pending" {
		t.Fatalf("expected pending state, got %q", state)
	}
	if got := client.quotaContexts[0].Value(serviceQuotaWaitContextKey("quota")); got != "refresh" {
		t.Fatalf("expected refresh context to reach AWS finder, got %v", got)
	}
}

func TestStatusServiceQuotaRequestRejectsClosedRequest(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{
		requestedQuota: &awstypes.RequestedServiceQuotaChange{
			Status: awstypes.RequestStatusCaseClosed,
		},
	}

	_, _, err := statusServiceQuotaRequest(client, "request-id")(context.Background())

	if err == nil || !strings.Contains(err.Error(), "closed without approval") {
		t.Fatalf("expected closed request error, got %v", err)
	}
}

func TestValidateApprovedServiceQuotaValueRejectsPartialApproval(t *testing.T) {
	t.Parallel()

	if err := validateApprovedServiceQuotaValue(100, 80); err == nil {
		t.Fatal("expected partial approval to be rejected")
	}
}

func TestFindOpenServiceQuotaRequestByQuotaReturnsOpenRequest(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{
		requestedQuotas: []awstypes.RequestedServiceQuotaChange{
			{Id: aws.String("closed"), Status: awstypes.RequestStatusDenied},
			{Id: aws.String("open"), Status: awstypes.RequestStatusCaseOpened},
		},
	}

	request, err := findOpenServiceQuotaRequestByQuota(context.Background(), client, "service", "quota")

	if err != nil {
		t.Fatalf("findOpenServiceQuotaRequestByQuota returned an error: %s", err)
	}
	if got := aws.ToString(request.Id); got != "open" {
		t.Fatalf("expected open request ID, got %q", got)
	}
}

func TestFindOpenServiceQuotaRequestByQuotaReturnsNotFound(t *testing.T) {
	t.Parallel()

	_, err := findOpenServiceQuotaRequestByQuota(context.Background(), &fakeServiceQuotaWaitClient{}, "service", "quota")

	var notFound *retry.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected provider not-found error, got %T: %v", err, err)
	}
}

func TestFindServiceQuotaReturnsProviderNotFound(t *testing.T) {
	t.Parallel()

	client := &fakeServiceQuotaWaitClient{quotaErr: &awstypes.NoSuchResourceException{}}
	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String("quota"),
		ServiceCode: aws.String("service"),
	}

	_, err := findServiceQuota(context.Background(), client, input)

	var notFound *retry.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected provider not-found error, got %T: %v", err, err)
	}
}
