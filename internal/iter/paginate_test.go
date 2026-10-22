// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type pagerOptions struct{}

type pagerOutput struct {
	Items []string
}

// fakePager stands in for an AWS SDK for Go v2 paginator.
type fakePager struct {
	pages  [][]string
	errAt  int // 1-based page to fail on; 0 never fails
	calls  int
	optFns int
}

func (p *fakePager) HasMorePages() bool {
	return p.calls < len(p.pages)
}

func (p *fakePager) NextPage(_ context.Context, optFns ...func(*pagerOptions)) (*pagerOutput, error) {
	p.calls++
	p.optFns += len(optFns)

	if p.errAt == p.calls {
		return nil, errPage
	}

	return &pagerOutput{Items: p.pages[p.calls-1]}, nil
}

var errPage = errors.New("boom")

func pagerItems(o *pagerOutput) []string { return o.Items }

func TestPageItems(t *testing.T) {
	t.Parallel()

	type testCase struct {
		pages     [][]string
		errAt     int
		want      []string
		wantErr   bool
		wantCalls int
	}
	tests := map[string]testCase{
		"no pages": {
			wantCalls: 0,
		},
		"one page": {
			pages:     [][]string{{"a", "b"}},
			want:      []string{"a", "b"},
			wantCalls: 1,
		},
		"three pages, one empty": {
			pages:     [][]string{{"a"}, {}, {"b", "c"}},
			want:      []string{"a", "b", "c"},
			wantCalls: 3,
		},
		"error on first page": {
			pages:     [][]string{{"a"}, {"b"}},
			errAt:     1,
			wantErr:   true,
			wantCalls: 1,
		},
		"error on second page yields the first page's items": {
			pages:     [][]string{{"a"}, {"b"}},
			errAt:     2,
			want:      []string{"a"},
			wantErr:   true,
			wantCalls: 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &fakePager{pages: test.pages, errAt: test.errAt}

			var got []string
			var gotErr error
			for item, err := range PageItems[pagerOptions](context.Background(), "listing things", p, pagerItems) {
				if err != nil {
					gotErr = err

					break
				}
				got = append(got, item)
			}

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("unexpected items (+want, -got): %s", diff)
			}
			if test.wantErr {
				if gotErr == nil {
					t.Fatal("expected an error, got none")
				}
				if !errors.Is(gotErr, errPage) {
					t.Errorf("error does not wrap the paginator error: %v", gotErr)
				}
				if want := "listing things: boom"; gotErr.Error() != want {
					t.Errorf("error = %q, want %q", gotErr, want)
				}
			} else if gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}
			if p.calls != test.wantCalls {
				t.Errorf("NextPage called %d times, want %d", p.calls, test.wantCalls)
			}
		})
	}
}

func TestPageItemsStopsEarly(t *testing.T) {
	t.Parallel()

	p := &fakePager{pages: [][]string{{"a", "b"}, {"c"}}}

	// Breaking out of the first page must not request the second.
	for _, err := range PageItems[pagerOptions](context.Background(), "listing things", p, pagerItems) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		break
	}

	if p.calls != 1 {
		t.Errorf("NextPage called %d times after an early break, want 1", p.calls)
	}
}

func TestPagesPassesOptionFunctions(t *testing.T) {
	t.Parallel()

	p := &fakePager{pages: [][]string{{"a"}, {"b"}}}
	opt := func(*pagerOptions) {}

	pages := slices.Collect(func(yield func([]string) bool) {
		for items, err := range pages(context.Background(), "listing things", p, pagerItems, opt, opt) {
			if err != nil {
				return
			}
			if !yield(items) {
				return
			}
		}
	})

	if len(pages) != 2 {
		t.Errorf("collected %d pages, want 2", len(pages))
	}
	if want := 4; p.optFns != want {
		t.Errorf("option functions passed %d times, want %d", p.optFns, want)
	}
}
