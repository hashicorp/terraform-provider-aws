// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

var (
	testBackoffOpts = backoff.WithDelay(backoff.ZeroDelay) // Use 0 delay for testing
)

type ErrorGenerator struct {
	position      int
	errorSequence []error
}

func (g *ErrorGenerator) NextError() (int, error) {
	var err error
	p := g.position
	if len(g.errorSequence)-1 >= p {
		err = g.errorSequence[p]
	} else {
		return -1, errors.New("no more errors available")
	}

	g.position += 1

	return p, err
}

func NewErrorGenerator(sequence []error) *ErrorGenerator {
	g := &ErrorGenerator{}
	g.errorSequence = sequence

	return g
}

func NewOpFunc(sequence []error, d time.Duration) retry.OpFunc[any] {
	g := NewErrorGenerator(sequence)

	return func(ctx context.Context) (any, error) {
		if d > 0 {
			time.Sleep(d)
		}
		idx, err := g.NextError()
		if err != nil {
			return nil, err
		}
		return idx, nil
	}
}

func UntilFoundOpFunc() retry.OpFunc[any] {
	sequence := []error{
		tfresource.NewEmptyResultError(nil),
		tfresource.NewEmptyResultError(nil),
		tfresource.NewEmptyResultError(nil),
		nil,
		tfresource.NewEmptyResultError(nil),
		nil,
		nil,
		nil,
	}

	return NewOpFunc(sequence, 0)
}

func UntilFoundSleepOpFunc() retry.OpFunc[any] {
	sequence := []error{
		tfresource.NewEmptyResultError(nil),
		tfresource.NewEmptyResultError(nil),
		tfresource.NewEmptyResultError(nil),
		nil,
		tfresource.NewEmptyResultError(nil),
		nil,
		nil,
		nil,
	}

	return NewOpFunc(sequence, 1*time.Second)
}

func UntilNotFoundOpFunc() retry.OpFunc[any] {
	sequence := []error{
		nil,
		nil,
		nil,
		tfresource.NewEmptyResultError(nil),
	}

	return NewOpFunc(sequence, 0)
}

func UntilNotFoundFailureOpFunc() retry.OpFunc[any] {
	sequence := []error{
		nil,
		nil,
		nil,
		nil,
	}

	return NewOpFunc(sequence, 0)
}

func TestUntilFoundN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		InARow  int
		WantErr bool
	}{
		{
			Name:   "Once",
			InARow: 1,
		},
		{
			Name:   "Two in a row",
			InARow: 2,
		},
		{
			Name:    "Four in a row",
			InARow:  4,
			WantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			_, err := retry.Operation(UntilFoundOpFunc()).UntilFoundN(testCase.InARow).Run(t.Context(), 2*time.Minute, testBackoffOpts)
			if gotErr := err != nil; gotErr != testCase.WantErr {
				t.Errorf("err = %v, want error presence = %v", err, testCase.WantErr)
			}
		})
	}
}
func TestUntilFoundN_timeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		InARow  int
		WantErr bool
	}{
		{
			Name:    "Once",
			InARow:  1,
			WantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			_, err := retry.Operation(UntilFoundSleepOpFunc()).UntilFoundN(testCase.InARow).Run(t.Context(), 2*time.Second, testBackoffOpts)
			if gotErr := err != nil; gotErr != testCase.WantErr {
				t.Errorf("err = %v, want error presence = %v", err, testCase.WantErr)
			}
		})
	}
}

func TestUntilNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		Op      retry.OpFunc[any]
		WantErr bool
	}{
		{
			Name: "Not found",
			Op:   UntilNotFoundOpFunc(),
		},
		{
			Name:    "Always found",
			Op:      UntilNotFoundFailureOpFunc(),
			WantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			_, err := retry.Operation(testCase.Op).UntilNotFound().Run(t.Context(), 2*time.Minute, testBackoffOpts)
			if gotErr := err != nil; gotErr != testCase.WantErr {
				t.Errorf("err = %v, want error presence = %v", err, testCase.WantErr)
			}
		})
	}
}
