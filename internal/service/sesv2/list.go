// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

func ListConfigurationSetsPages(ctx context.Context, conn *sesv2.Client, in *sesv2.ListConfigurationSetsInput, fn func(*sesv2.ListConfigurationSetsOutput, bool) bool) error {
	for {
		out, err := conn.ListConfigurationSets(ctx, in)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(out.NextToken) == ""
		if !fn(out, lastPage) || lastPage {
			break
		}

		in.NextToken = out.NextToken
	}

	return nil
}

func ListContactListsPages(ctx context.Context, conn *sesv2.Client, in *sesv2.ListContactListsInput, fn func(*sesv2.ListContactListsOutput, bool) bool) error {
	for {
		out, err := conn.ListContactLists(ctx, in)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(out.NextToken) == ""
		if !fn(out, lastPage) || lastPage {
			break
		}

		in.NextToken = out.NextToken
	}

	return nil
}
