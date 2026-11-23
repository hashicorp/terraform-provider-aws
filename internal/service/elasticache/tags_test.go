// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestWaitTagsPropagated(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		tags         map[string]string
		responses    []map[string]string
		ignoreConfig *tftags.IgnoreConfig
		withContext  bool
		wantError    bool
		minCalls     int32
	}{
		"no tags": {
			responses: []map[string]string{{}},
		},
		"empty user tag value with system tags": {
			tags:      map[string]string{names.AttrKey: ""},
			responses: []map[string]string{{names.AttrKey: "", "aws:cloudformation:stack-id": "stack"}},
		},
		"only ignored and system tags remain": {
			responses: []map[string]string{{"ignored": "external", "aws:cloudformation:stack-id": "stack"}},
			ignoreConfig: &tftags.IgnoreConfig{
				Keys: tftags.New(t.Context(), []string{"ignored"}),
			},
			withContext: true,
		},
		"all user tags removed but still present": {
			responses: []map[string]string{{names.AttrKey: names.AttrValue, "aws:cloudformation:stack-id": "stack"}},
			wantError: true,
		},
		"wait for added user tags": {
			tags: map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{
				{"aws:cloudformation:stack-id": "stack"},
				{names.AttrKey: names.AttrValue, "aws:cloudformation:stack-id": "stack"},
			},
			minCalls: 3,
		},
		"wait for changed user tags": {
			tags: map[string]string{names.AttrKey: "new"},
			responses: []map[string]string{
				{names.AttrKey: "old", "aws:cloudformation:stack-id": "stack"},
				{names.AttrKey: "new", "aws:cloudformation:stack-id": "stack"},
			},
			minCalls: 3,
		},
		"matching user tags": {
			tags:      map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{{names.AttrKey: names.AttrValue}},
		},
		"system tags without tagging context": {
			tags:      map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{{names.AttrKey: names.AttrValue, "aws:cloudformation:stack-id": "stack"}},
		},
		"system tags with tagging context": {
			tags:        map[string]string{names.AttrKey: names.AttrValue},
			responses:   []map[string]string{{names.AttrKey: names.AttrValue, "aws:cloudformation:stack-id": "stack"}},
			withContext: true,
		},
		"only system tags remain": {
			responses: []map[string]string{{"aws:cloudformation:stack-id": "stack"}},
		},
		"ignored tags": {
			tags: map[string]string{names.AttrKey: names.AttrValue, "ignored": "old", "external:key": "old"},
			responses: []map[string]string{{
				names.AttrKey: names.AttrValue, "ignored": "new", "external:key": "new", "aws:cloudformation:stack-id": "stack",
			}},
			ignoreConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(t.Context(), []string{"ignored"}),
				KeyPrefixes: tftags.New(t.Context(), []string{"external:"}),
			},
			withContext: true,
		},
		"wait for removed user tags": {
			tags: map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{
				{names.AttrKey: names.AttrValue, "removed": "old", "aws:cloudformation:stack-id": "stack"},
				{names.AttrKey: names.AttrValue, "aws:cloudformation:stack-id": "stack"},
			},
			minCalls: 3,
		},
		"removed user tag still present": {
			tags:      map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{{names.AttrKey: names.AttrValue, "removed": "old", "aws:cloudformation:stack-id": "stack"}},
			wantError: true,
		},
		"changed user tag not propagated": {
			tags:      map[string]string{names.AttrKey: "new"},
			responses: []map[string]string{{names.AttrKey: "old", "aws:cloudformation:stack-id": "stack"}},
			wantError: true,
		},
		"missing user tag": {
			tags:      map[string]string{names.AttrKey: names.AttrValue},
			responses: []map[string]string{{"aws:cloudformation:stack-id": "stack"}},
			wantError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Errorf("parsing request: %v", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if action := r.Form.Get("Action"); action != "ListTagsForResource" {
					t.Errorf("unexpected action: %s", action)
					http.Error(w, "unexpected action", http.StatusBadRequest)
					return
				}

				response := testcase.responses[min(int(calls.Add(1))-1, len(testcase.responses)-1)]
				var output struct {
					XMLName xml.Name `xml:"ListTagsForResourceResponse"`
					Tags    []struct {
						Key   string
						Value string
					} `xml:"ListTagsForResourceResult>TagList>Tag"`
				}
				for k, v := range response {
					output.Tags = append(output.Tags, struct {
						Key   string
						Value string
					}{Key: k, Value: v})
				}
				w.Header().Set("Content-Type", "text/xml")
				if err := xml.NewEncoder(w).Encode(output); err != nil {
					t.Errorf("encoding response: %v", err)
				}
			}))
			defer server.Close()

			conn := elasticache.New(elasticache.Options{
				Region:       "us-west-2", //lintignore:AWSAT003
				BaseEndpoint: aws.String(server.URL),
				Credentials:  aws.AnonymousCredentials{},
			})
			timeout := 10 * time.Second
			if testcase.wantError {
				// A successful waiter needs two polls at least one second apart.
				timeout = 3 * time.Second
			}
			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			defer cancel()
			if testcase.withContext {
				ctx = tftags.NewContext(ctx, nil, testcase.ignoreConfig, nil)
			}

			err := waitTagsPropagated(ctx, conn, "test-resource", tftags.New(ctx, testcase.tags))
			if testcase.wantError {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("expected deadline while user tags differ, got %v", err)
				}
			} else if err != nil {
				t.Errorf("waiting for tag propagation: %v", err)
			}
			minCalls := max(testcase.minCalls, 2)
			if got := calls.Load(); got < minCalls {
				t.Errorf("ListTagsForResource calls = %d, want at least %d", got, minCalls)
			}
		})
	}
}
