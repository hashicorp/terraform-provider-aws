// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testReplicationDomainARN = "arn:aws:es:us-east-1:123456789012:domain/primary" //lintignore:AWSAT003,AWSAT005
	testReplicationOurPeer   = "arn:aws:es:us-east-1:123456789012:domain/ours"    //lintignore:AWSAT003,AWSAT005
	testReplicationPeerB     = "arn:aws:es:us-east-1:123456789012:domain/other-b" //lintignore:AWSAT003,AWSAT005
	testReplicationPeerC     = "arn:aws:es:us-east-1:123456789012:domain/other-c" //lintignore:AWSAT003,AWSAT005
)

func testReplicationPeer(arn string, status awstypes.ReplicationStatus, msg string) awstypes.ReplicationPeer {
	p := awstypes.ReplicationPeer{
		PeerDomainArn: aws.String(arn),
		PeerRole:      awstypes.ReplicationRoleSecondary,
		Status:        status,
	}
	if msg != "" {
		p.StatusMessage = aws.String(msg)
	}

	return p
}

func testReplicationOutput(peers ...awstypes.ReplicationPeer) *opensearch.DescribeDomainReplicationOutput {
	return &opensearch.DescribeDomainReplicationOutput{
		DomainArn: aws.String(testReplicationDomainARN),
		Role:      awstypes.ReplicationRolePrimary,
		Peers:     peers,
	}
}

// testReplicationResponse is one scripted DescribeDomainReplication result: an output, or an error.
type testReplicationResponse struct {
	out *opensearch.DescribeDomainReplicationOutput
	err error
}

// testReplicationClient returns a client whose DescribeDomainReplication calls return the
// scripted responses in order, repeating the last one. No request leaves the process.
func testReplicationClient(t *testing.T, calls *atomic.Int32, responses ...testReplicationResponse) *opensearch.Client {
	t.Helper()

	fake := middleware.InitializeMiddlewareFunc("fakeDescribeDomainReplication", func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
		if _, ok := in.Parameters.(*opensearch.DescribeDomainReplicationInput); !ok {
			return middleware.InitializeOutput{}, middleware.Metadata{}, fmt.Errorf("unexpected operation input %T", in.Parameters)
		}

		n := int(calls.Add(1)) - 1
		r := responses[min(n, len(responses)-1)]

		return middleware.InitializeOutput{Result: r.out}, middleware.Metadata{}, r.err
	})

	return opensearch.New(opensearch.Options{
		Region:      "us-east-1", //lintignore:AWSAT003
		Credentials: aws.AnonymousCredentials{},
		APIOptions: []func(*middleware.Stack) error{
			func(s *middleware.Stack) error {
				return s.Initialize.Add(fake, middleware.Before)
			},
		},
	})
}

func testReplicationNotFound() testReplicationResponse {
	return testReplicationResponse{err: &awstypes.ResourceNotFoundException{Message: aws.String("domain is not replicating")}}
}

func TestFindReplicationPeerByARN_multiplePeers(t *testing.T) {
	t.Parallel()

	ours := testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, "ours")
	b := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusFailed, "b")
	c := testReplicationPeer(testReplicationPeerC, awstypes.ReplicationStatusPendingSetup, "c")

	for name, tc := range map[string]struct {
		out       *opensearch.DescribeDomainReplicationOutput
		wantFound bool
	}{
		"nil output":   {out: nil},
		"no peers":     {out: testReplicationOutput()},
		"ours first":   {out: testReplicationOutput(ours, b, c), wantFound: true},
		"ours middle":  {out: testReplicationOutput(b, ours, c), wantFound: true},
		"ours last":    {out: testReplicationOutput(b, c, ours), wantFound: true},
		"ours absent":  {out: testReplicationOutput(b, c)},
		"ours only":    {out: testReplicationOutput(ours), wantFound: true},
		"prefix match": {out: testReplicationOutput(testReplicationPeer(testReplicationOurPeer+"-2", awstypes.ReplicationStatusActive, ""))},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := findReplicationPeerByARN(tc.out, testReplicationOurPeer)
			if !tc.wantFound {
				if got != nil {
					t.Fatalf("expected no peer, got %s", aws.ToString(got.PeerDomainArn))
				}
				return
			}
			if got == nil {
				t.Fatal("expected our peer, got nil")
			}
			if aws.ToString(got.PeerDomainArn) != testReplicationOurPeer || aws.ToString(got.StatusMessage) != "ours" {
				t.Fatalf("picked wrong peer: %s (%s)", aws.ToString(got.PeerDomainArn), aws.ToString(got.StatusMessage))
			}
		})
	}
}

func TestDomainReplicationSetFromDescribe_multiplePeers(t *testing.T) {
	t.Parallel()

	ours := awstypes.ReplicationPeer{
		PeerDomainArn: aws.String(testReplicationOurPeer),
		PeerRole:      awstypes.ReplicationRoleSecondary,
		Status:        awstypes.ReplicationStatusPendingRemoval,
		StatusMessage: aws.String("ours"),
	}
	b := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusFailed, "b failed")
	c := testReplicationPeer(testReplicationPeerC, awstypes.ReplicationStatusActive, "")

	for name, out := range map[string]*opensearch.DescribeDomainReplicationOutput{
		"ours first": testReplicationOutput(ours, b, c),
		"ours last":  testReplicationOutput(b, c, ours),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			data := domainReplicationResourceModel{PeerDomainARN: fwtypes.ARNValue(testReplicationOurPeer)}
			data.setFromDescribe(ctx, out)

			want := map[string]attr.Value{
				"domain_arn":            types.StringValue(testReplicationDomainARN),
				names.AttrRole:          fwtypes.StringEnumValue(awstypes.ReplicationRolePrimary),
				"peer_role":             fwtypes.StringEnumValue(awstypes.ReplicationRoleSecondary),
				names.AttrStatus:        fwtypes.StringEnumValue(awstypes.ReplicationStatusPendingRemoval),
				names.AttrStatusMessage: types.StringValue("ours"),
			}
			got := map[string]attr.Value{
				"domain_arn":            data.DomainARN,
				names.AttrRole:          data.Role,
				"peer_role":             data.PeerRole,
				names.AttrStatus:        data.Status,
				names.AttrStatusMessage: data.StatusMessage,
			}
			for k, w := range want {
				if !got[k].Equal(w) {
					t.Errorf("%s: got %s, want %s", k, got[k], w)
				}
			}
		})
	}

	t.Run("no status message on our peer", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		data := domainReplicationResourceModel{PeerDomainARN: fwtypes.ARNValue(testReplicationOurPeer)}
		data.setFromDescribe(ctx, testReplicationOutput(b, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, "")))

		if !data.StatusMessage.IsNull() {
			t.Errorf("status_message: got %s, want null (must not borrow another peer's message)", data.StatusMessage)
		}
	})
}

func TestFindDomainReplicationByTwoPartKey_multiplePeers(t *testing.T) {
	t.Parallel()

	activeB := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, "")
	for name, tc := range map[string]struct {
		resp         testReplicationResponse
		wantNotFound bool
	}{
		"not replicating":                {resp: testReplicationNotFound(), wantNotFound: true},
		"empty peer list":                {resp: testReplicationResponse{out: testReplicationOutput()}, wantNotFound: true},
		"other peers only":               {resp: testReplicationResponse{out: testReplicationOutput(activeB)}, wantNotFound: true},
		"ours detached, other active":    {resp: testReplicationResponse{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusDetached, ""))}, wantNotFound: true},
		"ours active, other detached":    {resp: testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusDetached, ""), testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, ""))}},
		"ours ejected, other active":     {resp: testReplicationResponse{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusEjected, ""))}},
		"ours failed, other active":      {resp: testReplicationResponse{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusFailed, "boom"))}},
		"ours pending setup among peers": {resp: testReplicationResponse{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusPendingSetup, ""), testReplicationPeer(testReplicationPeerC, awstypes.ReplicationStatusFailed, ""))}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var calls atomic.Int32

			out, err := findDomainReplicationByTwoPartKey(ctx, testReplicationClient(t, &calls, tc.resp), "primary", testReplicationOurPeer)
			if tc.wantNotFound {
				if !retry.NotFound(err) {
					t.Fatalf("expected not-found, got out=%v err=%v", out, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if findReplicationPeerByARN(out, testReplicationOurPeer) == nil {
				t.Fatal("returned output does not list our peer")
			}
		})
	}
}

func TestStatusDomainReplication_multiplePeers(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		resp       testReplicationResponse
		wantNil    bool
		wantStatus string
		wantErr    string
	}{
		"ours active, other failed": {
			resp:       testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusFailed, "b failed"), testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, ""))},
			wantStatus: string(awstypes.ReplicationStatusActive),
		},
		"ours failed, others active": {
			resp:       testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, "b fine"), testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusFailed, "ours failed"), testReplicationPeer(testReplicationPeerC, awstypes.ReplicationStatusActive, "c fine"))},
			wantStatus: string(awstypes.ReplicationStatusFailed),
			wantErr:    "ours failed",
		},
		"ours detached, others active": {
			resp:    testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, ""), testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusDetached, ""))},
			wantNil: true,
		},
		"ours absent, others active": {
			resp:    testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, ""))},
			wantNil: true,
		},
		"ours pending removal, other pending setup": {
			resp:       testReplicationResponse{out: testReplicationOutput(testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusPendingSetup, ""), testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusPendingRemoval, ""))},
			wantStatus: string(awstypes.ReplicationStatusPendingRemoval),
		},
		"not replicating": {
			resp:    testReplicationNotFound(),
			wantNil: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var calls atomic.Int32

			got, status, err := statusDomainReplication(testReplicationClient(t, &calls, tc.resp), "primary", testReplicationOurPeer)(ctx)

			if tc.wantNil {
				if got != nil || status != "" || err != nil {
					t.Fatalf("expected gone (nil, \"\", nil), got (%v, %q, %v)", got, status, err)
				}
				return
			}
			if status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", status, tc.wantStatus)
			}
			if tc.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error: got %v, want it to contain %q", err, tc.wantErr)
				}
				if err != nil && (strings.Contains(err.Error(), "b fine") || strings.Contains(err.Error(), "c fine")) {
					t.Errorf("error uses another peer's message: %s", err)
				}
			}
		})
	}
}

func TestWaitDomainReplicationActive_multiplePeers(t *testing.T) {
	t.Parallel()

	pendingOurs := testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusPendingSetup, "")
	activeOurs := testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, "")
	activeB := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, "")
	failedB := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusFailed, "b failed")

	for name, tc := range map[string]struct {
		responses []testReplicationResponse
		wantCalls int32
		wantErr   string
	}{
		"other active does not end the wait early": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(activeB, pendingOurs)},
				{out: testReplicationOutput(activeB, pendingOurs)},
				{out: testReplicationOutput(activeB, activeOurs)},
			},
			wantCalls: 3,
		},
		"other failed does not fail ours": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(failedB, pendingOurs)},
				{out: testReplicationOutput(activeOurs, failedB)},
			},
			wantCalls: 2,
		},
		"ours failed fails fast with our message": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(activeB, pendingOurs)},
				{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusFailed, "ours failed"))},
			},
			wantCalls: 2,
			wantErr:   "ours failed",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var calls atomic.Int32

			out, err := waitDomainReplicationActive(ctx, testReplicationClient(t, &calls, tc.responses...), "primary", testReplicationOurPeer, 30*time.Second)

			if got := calls.Load(); got != tc.wantCalls {
				t.Errorf("describe calls: got %d, want %d", got, tc.wantCalls)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error: got %v, want it to contain %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if p := findReplicationPeerByARN(out, testReplicationOurPeer); p == nil || p.Status != awstypes.ReplicationStatusActive {
				t.Fatalf("returned output does not show our peer ACTIVE: %+v", out)
			}
		})
	}
}

func TestWaitDomainReplicationRemoved_multiplePeers(t *testing.T) {
	t.Parallel()

	activeB := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusActive, "")
	detachedB := testReplicationPeer(testReplicationPeerB, awstypes.ReplicationStatusDetached, "")
	removingOurs := testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusPendingRemoval, "")

	for name, tc := range map[string]struct {
		responses []testReplicationResponse
		wantCalls int32
		wantErr   string
	}{
		"ours detached while other active": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(activeB, removingOurs)},
				{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusDetached, ""))},
			},
			wantCalls: 2,
		},
		"ours dropped from list while other active": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusActive, ""), activeB)},
				{out: testReplicationOutput(activeB)},
			},
			wantCalls: 2,
		},
		"other detached does not end the wait early": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(detachedB, removingOurs)},
				{out: testReplicationOutput(removingOurs)},
				{out: testReplicationOutput(testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusDetached, ""))},
			},
			wantCalls: 3,
		},
		"ours ejected then domain not replicating": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusEjected, ""))},
				testReplicationNotFound(),
			},
			wantCalls: 2,
		},
		"ours failed during removal": {
			responses: []testReplicationResponse{
				{out: testReplicationOutput(activeB, removingOurs)},
				{out: testReplicationOutput(activeB, testReplicationPeer(testReplicationOurPeer, awstypes.ReplicationStatusFailed, "ours failed"))},
			},
			wantCalls: 2,
			wantErr:   "ours failed",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var calls atomic.Int32

			_, err := waitDomainReplicationRemoved(ctx, testReplicationClient(t, &calls, tc.responses...), "primary", testReplicationOurPeer, 30*time.Second)

			if got := calls.Load(); got != tc.wantCalls {
				t.Errorf("describe calls: got %d, want %d", got, tc.wantCalls)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error: got %v, want it to contain %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
