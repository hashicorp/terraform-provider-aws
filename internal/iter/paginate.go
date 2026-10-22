// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"context"
	"fmt"
	"iter"
)

// Paginator is an AWS SDK for Go v2 paginator over pages of type O, where OPT
// is the service's Options type. The SDK's generated paginators satisfy it.
type Paginator[O, OPT any] interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*OPT)) (*O, error)
}

// PageItems returns an iterator over the items of every page produced by p.
// The first error is yielded wrapped as "what: err" and iteration stops.
//
// OPT cannot be inferred when no option functions are passed, so give it
// explicitly:
//
//	tfiter.PageItems[ecr.Options](ctx, "listing ECR Repository resources",
//		ecr.NewDescribeRepositoriesPaginator(conn, input),
//		func(o *ecr.DescribeRepositoriesOutput) []awstypes.Repository { return o.Repositories })
//
// The sequence is single-use, because p holds the pagination state.
func PageItems[OPT, O, T any](ctx context.Context, what string, p Paginator[O, OPT], items func(*O) []T, optFns ...func(*OPT)) iter.Seq2[T, error] {
	return ConcatValuesWithError(pages(ctx, what, p, items, optFns...))
}

func pages[OPT, O, T any](ctx context.Context, what string, p Paginator[O, OPT], items func(*O) []T, optFns ...func(*OPT)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		for p.HasMorePages() {
			page, err := p.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("%s: %w", what, err))

				return
			}

			if !yield(items(page), nil) {
				return
			}
		}
	}
}
