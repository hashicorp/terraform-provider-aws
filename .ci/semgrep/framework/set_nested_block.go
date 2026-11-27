// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type model struct{}

func setNestedBlockCustomTypeCases(ctx context.Context) {
	// ruleid: set-nested-block-custom-type
	_ = schema.SetNestedBlock{}

	// ok: set-nested-block-custom-type
	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
	}

	// ruleid: set-nested-block-nil-custom-type
	_ = schema.SetNestedBlock{
		CustomType: nil,
	}
}

func setNestedBlockObjectCustomTypeCases(ctx context.Context) {
	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		// ruleid: set-nested-block-object-custom-type
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{},
		},
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		// ok: set-nested-block-object-custom-type
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[model](ctx),
			Attributes: map[string]schema.Attribute{},
		},
	}

	_ = schema.ListNestedBlock{
		// ok: set-nested-block-object-custom-type
		NestedObject: schema.NestedBlockObject{},
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		// ruleid: set-nested-block-object-custom-type
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"child": schema.ListNestedBlock{
					// ok: set-nested-block-object-custom-type
					NestedObject: schema.NestedBlockObject{},
				},
			},
		},
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[model](ctx),
			Blocks: map[string]schema.Block{
				"child": schema.SetNestedBlock{
					CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
					// ruleid: set-nested-block-object-custom-type
					NestedObject: schema.NestedBlockObject{},
				},
			},
		},
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		// ruleid: set-nested-block-object-nil-custom-type
		NestedObject: schema.NestedBlockObject{
			CustomType: nil,
		},
	}

	_ = datasourceschema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		Validators: []validator.Set{},
		// ruleid: set-nested-block-object-custom-type
		NestedObject: datasourceschema.NestedBlockObject{
			Attributes: map[string]datasourceschema.Attribute{},
		},
	}

	_ = datasourceschema.SetNestedBlock{
		// ok: set-nested-block-object-custom-type
		NestedObject: datasourceschema.NestedBlockObject{
			Attributes: map[string]datasourceschema.Attribute{},
			CustomType: fwtypes.NewObjectTypeOf[model](ctx),
		},
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		// ruleid: set-nested-block-object-custom-type
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"child": schema.StringAttribute{CustomType: types.StringType},
			},
		},
	}

	_ = schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[model](ctx),
			Blocks: map[string]schema.Block{
				"child": schema.SetNestedBlock{
					CustomType: fwtypes.NewSetNestedObjectTypeOf[model](ctx),
					// ruleid: set-nested-block-object-nil-custom-type
					NestedObject: schema.NestedBlockObject{
						CustomType: nil,
					},
				},
			},
		},
	}
}
