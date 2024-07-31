// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func AttributePath(in *tftypes.AttributePath) *tfplugin5.AttributePath {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.AttributePath{
		Steps: AttributePath_Steps(in.Steps()),
	}

	return resp
}

func AttributePaths(in []*tftypes.AttributePath) []*tfplugin5.AttributePath {
	resp := make([]*tfplugin5.AttributePath, 0, len(in))

	for _, a := range in {
		resp = append(resp, AttributePath(a))
	}

	return resp
}

func AttributePath_Step(step tftypes.AttributePathStep) *tfplugin5.AttributePath_Step {
	if step == nil {
		return nil
	}

	switch step := step.(type) {
	case tftypes.AttributeName:
		return &tfplugin5.AttributePath_Step{
			Selector: &tfplugin5.AttributePath_Step_AttributeName{
				AttributeName: string(step),
			},
		}
	case tftypes.ElementKeyInt:
		return &tfplugin5.AttributePath_Step{
			Selector: &tfplugin5.AttributePath_Step_ElementKeyInt{
				ElementKeyInt: int64(step),
			},
		}
	case tftypes.ElementKeyString:
		return &tfplugin5.AttributePath_Step{
			Selector: &tfplugin5.AttributePath_Step_ElementKeyString{
				ElementKeyString: string(step),
			},
		}
	case tftypes.ElementKeyValue:
		// The protocol has no equivalent of an ElementKeyValue, so this
		// returns nil for the step to signal a step we cannot convey back
		// to Terraform.
		return nil
	}

	// It is not currently possible to create tftypes.AttributePathStep
	// implementations outside the tftypes package and these implementations
	// should rarely change, if ever, since they are critical to how
	// Terraform understands attribute paths. If this panic was reached, it
	// implies that a new step type was introduced and needs to be
	// implemented as a new case above or that this logic needs to be
	// otherwise changed to handle some new attribute path system.
	panic(fmt.Sprintf("unimplemented tftypes.AttributePathStep type: %T", step))
}

func AttributePath_Steps(in []tftypes.AttributePathStep) []*tfplugin5.AttributePath_Step {
	resp := make([]*tfplugin5.AttributePath_Step, 0, len(in))

	for _, step := range in {
		s := AttributePath_Step(step)

		// In the face of a ElementKeyValue or missing step, Terraform has no
		// way to represent the attribute path, so only return the prefix.
		if s == nil {
			return resp
		}

		resp = append(resp, s)
	}

	return resp
}
