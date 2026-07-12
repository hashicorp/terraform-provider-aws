// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// updateTags wraps the generated updateTagsBase with additional retry logic
// for UserNotFoundFault and UserGroupNotFoundFault, which occur transiently
// when ElastiCache users/user groups are in a brief internal "modifying" state
// during tag operations.
// See: https://github.com/hashicorp/terraform-provider-aws/issues/47638
func updateTags(ctx context.Context, conn *elasticache.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*elasticache.Options)) error {
	_, err := tfresource.RetryWhen[any](ctx, 5*time.Minute,
		func(ctx context.Context) (any, error) {
			return nil, updateTagsBase(ctx, conn, identifier, oldTagsMap, newTagsMap, optFns...)
		},
		func(err error) (bool, error) {
			// Retry when a user is not available for tagging (transient ~5s state).
			if errs.IsAErrorMessageContains[*awstypes.UserNotFoundFault](err, "is not available for tagging") {
				return true, err
			}
			// Retry when a user group is not available for tagging (same race condition).
			if errs.IsA[*awstypes.UserGroupNotFoundFault](err) {
				return true, err
			}
			return false, err
		},
	)
	return err
}

// UpdateTags updates elasticache service tags.
// It is called from outside this package (by the tagging interceptor).
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient).ElastiCacheClient(ctx), identifier, oldTags, newTags)
}
