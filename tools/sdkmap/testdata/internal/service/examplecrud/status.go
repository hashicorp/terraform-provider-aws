// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusWidget(conn *examplecrud.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findWidgetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
