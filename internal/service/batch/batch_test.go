// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ statecheck.StateCheck = expectAllTagsCheck{}

type expectAllTagsCheck struct {
	stateCheckBase
	knownValue knownvalue.Check
}

func (e expectAllTagsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resource, ok := e.resourceFromState(req, resp)
	if !ok {
		return
	}

	identifier, ok := resource.AttributeValues[names.AttrARN]
	if !ok {
		resp.Error = fmt.Errorf("attribute %q not found in resource %s", names.AttrARN, e.resourceAddress)
	}

	conn := acctest.Provider.Meta().(*conns.AWSClient).BatchClient(ctx)

	tags, err := tfbatch.ListTags(ctx, conn, identifier.(string))
	if err != nil {
		resp.Error = fmt.Errorf("listing tags for %s: %s", e.resourceAddress, err)
		return
	}

	tagsMap := tfmaps.ApplyToAllValues(tags.Map(), func(s string) any {
		return s
	})

	if err := e.knownValue.CheckValue(tagsMap); err != nil {
		resp.Error = fmt.Errorf("checking value for all tags for %s: %s", e.resourceAddress, err)
		return
	}
}

func expectAllTags(resourceAddress string, knownValue knownvalue.Check) statecheck.StateCheck {
	return expectAllTagsCheck{
		stateCheckBase: stateCheckBase{
			resourceAddress: resourceAddress,
		},
		knownValue: knownValue,
	}
}

type stateCheckBase struct {
	resourceAddress string
}

func (e stateCheckBase) resourceFromState(req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) (*tfjson.StateResource, bool) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")

		return nil, false
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")

		return nil, false
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")

		return nil, false
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if e.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", e.resourceAddress)

		return nil, false
	}

	return resource, true
}
