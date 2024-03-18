// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func Schema(in *tfprotov5.Schema) *tfplugin5.Schema {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Schema{
		Block:   Schema_Block(in.Block),
		Version: in.Version,
	}

	return resp
}

func Schema_Block(in *tfprotov5.SchemaBlock) *tfplugin5.Schema_Block {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Schema_Block{
		Attributes:      Schema_Attributes(in.Attributes),
		BlockTypes:      Schema_NestedBlocks(in.BlockTypes),
		Deprecated:      in.Deprecated,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Version:         in.Version,
	}

	return resp
}

func Schema_Attribute(in *tfprotov5.SchemaAttribute) *tfplugin5.Schema_Attribute {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Schema_Attribute{
		Computed:        in.Computed,
		Deprecated:      in.Deprecated,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Name:            in.Name,
		Optional:        in.Optional,
		Required:        in.Required,
		Sensitive:       in.Sensitive,
		Type:            CtyType(in.Type),
	}

	return resp
}

func Schema_Attributes(in []*tfprotov5.SchemaAttribute) []*tfplugin5.Schema_Attribute {
	resp := make([]*tfplugin5.Schema_Attribute, 0, len(in))

	for _, a := range in {
		resp = append(resp, Schema_Attribute(a))
	}

	return resp
}

func Schema_NestedBlock(in *tfprotov5.SchemaNestedBlock) *tfplugin5.Schema_NestedBlock {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Schema_NestedBlock{
		Block:    Schema_Block(in.Block),
		MaxItems: in.MaxItems,
		MinItems: in.MinItems,
		Nesting:  Schema_NestedBlock_NestingMode(in.Nesting),
		TypeName: in.TypeName,
	}

	return resp
}

func Schema_NestedBlocks(in []*tfprotov5.SchemaNestedBlock) []*tfplugin5.Schema_NestedBlock {
	resp := make([]*tfplugin5.Schema_NestedBlock, 0, len(in))

	for _, b := range in {
		resp = append(resp, Schema_NestedBlock(b))
	}

	return resp
}

func Schema_NestedBlock_NestingMode(in tfprotov5.SchemaNestedBlockNestingMode) tfplugin5.Schema_NestedBlock_NestingMode {
	return tfplugin5.Schema_NestedBlock_NestingMode(in)
}
