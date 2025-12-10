// Copyright IBM Corp. 2014, 2025
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

type value struct {
	val string
}

func FailedStateRefreshFunc() StateRefreshFunc {
	return func(context.Context) (any, string, error) {
		return nil, "", errors.New("failed")
	}
}

func TimeoutStateRefreshFunc() StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		select {
		case <-ctx.Done():
			return nil, "", &aws.RequestCanceledError{Err: ctx.Err()}
		case <-time.After(5 * time.Second):
		}
		return &value{val: "value"}, "pending", nil
	}
}

func SuccessfulStateRefreshFunc() StateRefreshFunc {
	return func(context.Context) (any, string, error) {
		return &value{val: "value"}, "running", nil
	}
}

type StateGenerator struct {
	position      int
	stateSequence []string
}

func (r *StateGenerator) NextState() (int, string, error) {
	p := r.position
	if p > len(r.stateSequence)-1 {
		return -1, "", errors.New("No more states available")
	}

	r.position++
	return p, r.stateSequence[p], nil
}

func NewStateGenerator(sequence []string) *StateGenerator {
	r := &StateGenerator{}
	r.stateSequence = sequence

	return r
}

func InconsistentStateRefreshFunc() StateRefreshFunc {
	sequence := []string{
		"done", "replicating",
		"done", "done", "done",
		"replicating",
		"done", "done", "done",
		"replicating", "replicating", "replicating", "replicating", "replicating",
	}

	r := NewStateGenerator(sequence)

	return func(context.Context) (any, string, error) {
		idx, s, err := r.NextState()
		if err != nil {
			return nil, "", err
		}

		return idx, s, nil
	}
}

func UnknownPendingStateRefreshFunc() StateRefreshFunc {
	sequence := []string{
		"unknown1", "unknown2", "done",
	}

	r := NewStateGenerator(sequence)

	return func(context.Context) (any, string, error) {
		idx, s, err := r.NextState()
		if err != nil {
			return nil, "", err
		}

		return idx, s, nil
	}
}

func TestWaitForState_inconsistent_positive(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending:                   []string{"replicating"},
		Target:                    []string{"done"},
		Refresh:                   InconsistentStateRefreshFunc(),
		Timeout:                   90 * time.Millisecond,
		PollInterval:              10 * time.Millisecond,
		ContinuousTargetOccurence: 3,
	}

	idx, err := conf.WaitForStateContext(t.Context())

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if idx != 4 {
		t.Fatalf("Expected index 4, given %d", idx.(int))
	}
}

func TestWaitForState_inconsistent_negative(t *testing.T) {
	t.Parallel()

	refreshCount := int64(0)
	f := InconsistentStateRefreshFunc()
	refresh := func(ctx context.Context) (any, string, error) {
		atomic.AddInt64(&refreshCount, 1)
		return f(ctx)
	}

	conf := &StateChangeConf{
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

func TestWaitForState_timeout(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: TimeoutStateRefreshFunc(),
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

func TestWaitForState_success(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: SuccessfulStateRefreshFunc(),
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

func TestWaitForState_successUnknownPending(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Target:  []string{"done"},
		Refresh: UnknownPendingStateRefreshFunc(),
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

func TestWaitForState_successEmpty(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{},
		Refresh: func(context.Context) (any, string, error) {
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

func TestWaitForState_failureEmpty(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending:        []string{"pending", "incomplete"},
		Target:         []string{},
		NotFoundChecks: 1,
		Refresh: func(context.Context) (any, string, error) {
			return 42, "pending", nil
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

func TestWaitForState_failure(t *testing.T) {
	t.Parallel()

	conf := &StateChangeConf{
		Pending: []string{"pending", "incomplete"},
		Target:  []string{"running"},
		Refresh: FailedStateRefreshFunc(),
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

func TestWaitForState_NotFound_NotFoundChecks(t *testing.T) {
	t.Parallel()

	const notFoundCheckCount = 5

	expectedErr := &NotFoundError{
		Retries: notFoundCheckCount + 1,
	}

	conf := &StateChangeConf{
		Pending:        []string{"pending", "incomplete"},
		Target:         []string{"running"},
		PollInterval:   10 * time.Millisecond,
		Timeout:        100 * time.Millisecond,
		NotFoundChecks: notFoundCheckCount,
		Refresh: func(context.Context) (any, string, error) {
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

type refresh[T any, S ~string] struct {
	obj   T
	state S
	err   error
}

func inconsistentResultStateRefreshFunc(t *testing.T) (StateRefreshFunc, func()) {
	t.Helper()

	sequence := []refresh[any, string]{
		{nil, "", nil}, // 1
		{&value{val: "value"}, "pending", nil},
		{nil, "", nil}, {nil, "", nil}, // 2
		{&value{val: "value"}, "pending", nil},
		{nil, "", nil}, {nil, "", nil}, {nil, "", nil}, // 3
	}

	next, stop := iter.Pull(slices.Values(sequence))

	return func(context.Context) (any, string, error) {
		v, _ := next()

		return v.obj, v.state, v.err
	}, stop
}

func TestWaitForState_EmptyTarget_ContinuousTargetOccurence(t *testing.T) {
	t.Parallel()

	const continuousTargetOccurence = 3
	const expectedCount = 8

	var count atomic.Int32

	inner, stop := inconsistentResultStateRefreshFunc(t)
	defer stop()

	refresh := func(ctx context.Context) (any, string, error) {
		count.Add(1)
		return inner(ctx)
	}

	conf := &StateChangeConf{
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

func TestWaitForStateContext_cancel(t *testing.T) {
	t.Parallel()

	// make this refresh func block until we cancel it
	ctx, cancel := context.WithCancel(context.Background())
	refresh := func(context.Context) (any, string, error) {
		<-ctx.Done()
		return nil, "pending", nil
	}
	conf := &StateChangeConf{
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
