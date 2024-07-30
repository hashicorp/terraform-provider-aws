// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func Schema(in *tfprotov6.Schema) *tfplugin6.Schema {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Schema{
		Block:   Schema_Block(in.Block),
		Version: in.Version,
	}

	return resp
}

func Schema_Block(in *tfprotov6.SchemaBlock) *tfplugin6.Schema_Block {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Schema_Block{
		Attributes:      Schema_Attributes(in.Attributes),
		BlockTypes:      Schema_NestedBlocks(in.BlockTypes),
		Deprecated:      in.Deprecated,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Version:         in.Version,
	}

	return resp
}

func Schema_Attribute(in *tfprotov6.SchemaAttribute) *tfplugin6.Schema_Attribute {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Schema_Attribute{
		Computed:        in.Computed,
		Deprecated:      in.Deprecated,
		Description:     in.Description,
		DescriptionKind: StringKind(in.DescriptionKind),
		Name:            in.Name,
		NestedType:      Schema_Object(in.NestedType),
		Optional:        in.Optional,
		Required:        in.Required,
		Sensitive:       in.Sensitive,
		Type:            CtyType(in.Type),
	}

	return resp
}

func Schema_Attributes(in []*tfprotov6.SchemaAttribute) []*tfplugin6.Schema_Attribute {
	resp := make([]*tfplugin6.Schema_Attribute, 0, len(in))

	for _, a := range in {
		resp = append(resp, Schema_Attribute(a))
	}

	return resp
}

func Schema_NestedBlock(in *tfprotov6.SchemaNestedBlock) *tfplugin6.Schema_NestedBlock {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Schema_NestedBlock{
		Block:    Schema_Block(in.Block),
		MaxItems: in.MaxItems,
		MinItems: in.MinItems,
		Nesting:  Schema_NestedBlock_NestingMode(in.Nesting),
		TypeName: in.TypeName,
	}

	return resp
}

func Schema_NestedBlocks(in []*tfprotov6.SchemaNestedBlock) []*tfplugin6.Schema_NestedBlock {
	resp := make([]*tfplugin6.Schema_NestedBlock, 0, len(in))

	for _, b := range in {
		resp = append(resp, Schema_NestedBlock(b))
	}

	return resp
}

func Schema_NestedBlock_NestingMode(in tfprotov6.SchemaNestedBlockNestingMode) tfplugin6.Schema_NestedBlock_NestingMode {
	return tfplugin6.Schema_NestedBlock_NestingMode(in)
}

func Schema_Object_NestingMode(in tfprotov6.SchemaObjectNestingMode) tfplugin6.Schema_Object_NestingMode {
	return tfplugin6.Schema_Object_NestingMode(in)
}

func Schema_Object(in *tfprotov6.SchemaObject) *tfplugin6.Schema_Object {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Schema_Object{
		Attributes: Schema_Attributes(in.Attributes),
		Nesting:    Schema_Object_NestingMode(in.Nesting),
	}

	return resp
}
