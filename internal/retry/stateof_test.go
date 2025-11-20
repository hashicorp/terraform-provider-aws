// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"errors"
	"iter"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
)

//
// Based on https://github.com/hashicorp/terraform-plugin-sdk/helper/retry/state_test.go.
//

func FailedStateRefreshFuncOf() StateRefreshFuncOf[*value, string] {
	return func(context.Context) (*value, string, error) {
		return nil, "", errors.New("failed")
	}
}

func TimeoutStateRefreshFuncOf() StateRefreshFuncOf[*value, string] {
	return func(ctx context.Context) (*value, string, error) {
		select {
		case <-ctx.Done():
			return nil, "", &aws.RequestCanceledError{Err: ctx.Err()}
		case <-time.After(5 * time.Second):
		}
		return &value{val: "value"}, "pending", nil
	}
}

func SuccessfulStateRefreshFuncOf() StateRefreshFuncOf[*value, string] {
	return func(context.Context) (*value, string, error) {
		return &value{val: "value"}, "running", nil
	}
}

func InconsistentStateRefreshFuncOf() StateRefreshFuncOf[*int, string] {
	sequence := []string{
		"done", "replicating",
		"done", "done", "done",
		"replicating",
		"done", "done", "done",
		"replicating", "replicating", "replicating", "replicating", "replicating",
	}

	r := NewStateGenerator(sequence)

	return func(context.Context) (*int, string, error) {
		idx, s, err := r.NextState()
		if err != nil {
			return nil, "", err
		}

		return &idx, s, nil
	}
}

func UnknownPendingStateRefreshFuncOf() StateRefreshFuncOf[*int, string] {
	sequence := []string{
		"unknown1", "unknown2", "done",
	}

	r := NewStateGenerator(sequence)

	return func(context.Context) (*int, string, error) {
		idx, s, err := r.NextState()
		if err != nil {
			return nil, "", err
		}

		return &idx, s, nil
	}
}

func TestWaitForStateOf_inconsistent_positive(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*int, string]{
		Pending:                   []string{"replicating"},
		Target:                    []string{"done"},
		Refresh:                   InconsistentStateRefreshFuncOf(),
		Timeout:                   90 * time.Millisecond,
		PollInterval:              10 * time.Millisecond,
		ContinuousTargetOccurence: 3,
	}

	idx, err := conf.WaitForStateContext(t.Context())

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if idx == nil {
		t.Fatal("Expected index 4, was nil")
	}
	if *idx != 4 {
		t.Fatalf("Expected index 4, given %d", *idx)
	}
}

func TestWaitForStateOf_inconsistent_negative(t *testing.T) {
	t.Parallel()

	refreshCount := int64(0)
	f := InconsistentStateRefreshFuncOf()
	refresh := func(ctx context.Context) (*int, string, error) {
		atomic.AddInt64(&refreshCount, 1)
		return f(ctx)
	}

	conf := &StateChangeConfOf[*int, string]{
		Pending:                   []string{"replicating"},
		Target:                    []string{"done"},
		Refresh:                   refresh,
		Timeout:                   85 * time.Millisecond,
		PollInterval:              10 * time.Millisecond,
		ContinuousTargetOccurence: 4,
	}

	_, err := conf.WaitForStateContext(t.Context())

	if err == nil {
		t.Fatal("Expected timeout error. No error returned.")
	}

	// we can't guarantee the exact number of refresh calls in the tests by
	// timing them, but we want to make sure the test at least went through the
	// required states.
	if atomic.LoadInt64(&refreshCount) < 6 {
		t.Fatal("refreshed called too few times")
	}

	expectedErr := "timeout while waiting for state to become 'done'"
	if !strings.HasPrefix(err.Error(), expectedErr) {
		t.Fatalf("error prefix doesn't match.\nExpected: %q\nGiven: %q\n", expectedErr, err.Error())
	}
}

func TestWaitForStateOf_timeout(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*value, string]{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: TimeoutStateRefreshFuncOf(),
		Timeout: 1 * time.Second,
	}

	obj, err := conf.WaitForStateContext(t.Context())

	if err == nil {
		t.Fatal("Expected timeout error. No error returned.")
	}

	expectedErr := "timeout while waiting for state to become 'running' (timeout: 1s)"
	if !strings.HasPrefix(err.Error(), expectedErr) {
		t.Fatalf("Errors don't match.\nExpected: %q\nGiven: %q\n", expectedErr, err.Error())
	}

	if obj != nil {
		t.Fatalf("should not return obj")
	}
}

func TestWaitForStateOf_success(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*value, string]{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: SuccessfulStateRefreshFuncOf(),
		Timeout: 200 * time.Second,
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if obj == nil {
		t.Fatalf("should return obj")
	}
}

func TestWaitForStateOf_successUnknownPending(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*int, string]{
		Target:  []string{"done"},
		Refresh: UnknownPendingStateRefreshFuncOf(),
		Timeout: 200 * time.Second,
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if obj == nil {
		t.Fatalf("should return obj")
	}
}

func TestWaitForStateOf_successEmpty(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*value, string]{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{},
		Refresh: func(context.Context) (*value, string, error) {
			return nil, "", nil
		},
		Timeout: 200 * time.Second,
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if obj != nil {
		t.Fatalf("obj should be nil")
	}
}

func TestWaitForStateOf_failureEmpty(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*value, string]{
		Pending:        []string{"pending", "incomplete"},
		Target:         []string{},
		NotFoundChecks: 1,
		Refresh: func(context.Context) (*value, string, error) {
			return &value{val: "forty-two"}, "pending", nil
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      100 * time.Millisecond,
	}

	_, err := conf.WaitForStateContext(t.Context())
	if err == nil {
		t.Fatal("Expected timeout error. Got none.")
	}
	expectedErr := "timeout while waiting for resource to be gone (last state: 'pending', timeout: 100ms)"
	if err.Error() != expectedErr {
		t.Fatalf("Errors don't match.\nExpected: %q\nGiven: %q\n", expectedErr, err.Error())
	}
}

func TestWaitForStateOf_failure(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConfOf[*value, string]{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: FailedStateRefreshFuncOf(),
		Timeout: 200 * time.Second,
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if err == nil {
		t.Fatal("Expected error. No error returned.")
	}
	expectedErr := "failed"
	if err.Error() != expectedErr {
		t.Fatalf("Errors don't match.\nExpected: %q\nGiven: %q\n", expectedErr, err.Error())
	}
	if obj != nil {
		t.Fatalf("should not return obj")
	}
}

func TestWaitForStateOf_NotFound_NotFoundChecks(t *testing.T) {
	t.Parallel()

	const notFoundCheckCount = 5

	expectedErr := &NotFoundError{
		Retries: notFoundCheckCount + 1,
	}

	conf := &StateChangeConfOf[*value, string]{
		Pending:        []string{"pending", "incomplete"},
		Target:         []string{"running"},
		PollInterval:   10 * time.Millisecond,
		Timeout:        100 * time.Millisecond,
		NotFoundChecks: notFoundCheckCount,
		Refresh: func(context.Context) (*value, string, error) {
			return nil, "", nil
		},
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if obj != nil {
		t.Errorf("should not return obj")
	}
	if err == nil {
		t.Fatal("Expected error. No error returned.")
	}

	if !cmp.Equal(expectedErr, err) {
		t.Errorf("Errors don't match.\nExpected: %q\nGiven: %q\n", expectedErr, err)
	}
}

func inconsistentResultStateRefreshFuncOf(t *testing.T) (StateRefreshFuncOf[*value, string], func()) {
	t.Helper()

	sequence := []refresh[*value, string]{
		{nil, "", nil}, // 1
		{&value{val: "value"}, "pending", nil},
		{nil, "", nil}, {nil, "", nil}, // 2
		{&value{val: "value"}, "pending", nil},
		{nil, "", nil}, {nil, "", nil}, {nil, "", nil}, // 3
	}

	next, stop := iter.Pull(slices.Values(sequence))

	return func(context.Context) (*value, string, error) {
		v, _ := next()

		return v.obj, v.state, v.err
	}, stop
}

func TestWaitForStateOf_EmptyTarget_ContinuousTargetOccurence(t *testing.T) {
	t.Parallel()

	const continuousTargetOccurence = 3
	const expectedCount = 8

	var count atomic.Int32

	inner, stop := inconsistentResultStateRefreshFuncOf(t)
	defer stop()

	refresh := func(ctx context.Context) (*value, string, error) {
		count.Add(1)
		return inner(ctx)
	}

	conf := &StateChangeConfOf[*value, string]{
		Pending:                   []string{"pending", "incomplete"},
		Target:                    []string{},
		Timeout:                   100 * time.Millisecond,
		PollInterval:              10 * time.Millisecond,
		ContinuousTargetOccurence: continuousTargetOccurence,
		Refresh:                   refresh,
	}

	obj, err := conf.WaitForStateContext(t.Context())
	if obj != nil {
		t.Errorf("should not return obj")
	}
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if v := count.Load(); v != expectedCount {
		t.Errorf("Expected %d refresh calls, got %d", expectedCount, v)
	}
}

func TestWaitForStateContextOf_cancel(t *testing.T) {
	t.Parallel()

	// make this refresh func block until we cancel it
	ctx, cancel := context.WithCancel(context.Background())
	refresh := func(context.Context) (*value, string, error) {
		<-ctx.Done()
		return nil, "pending", nil
	}
	conf := &StateChangeConfOf[*value, string]{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: refresh,
		Timeout: 10 * time.Second,
	}

	var err error

	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		_, err = conf.WaitForStateContext(ctx)
	}()

	// make sure WaitForState is blocked
	select {
	case <-waitDone:
		t.Fatal("WaitForState returned too early")
	case <-time.After(10 * time.Millisecond):
	}

	// unlock the refresh function
	cancel()
	// make sure WaitForState returns
	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("WaitForState didn't return after refresh finished")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected canceled context error, got: %s", err)
	}
}
