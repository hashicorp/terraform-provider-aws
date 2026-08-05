// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// NewResourceComputedListOfObjectSchema returns a new schema.ListAttribute for objects of the specified type.
// The list is Computed-only.
func ResourceComputedListOfObjectsAttribute[T any](ctx context.Context, planModifiers ...planmodifier.List) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType:    fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Computed:      true,
		PlanModifiers: planModifiers,
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}

// ResourceOptionalComputedListOfObjectsAttribute returns a new schema.ListAttribute for objects of the specified type.
// The list is Optional+Computed.
func ResourceOptionalComputedListOfObjectsAttribute[T any](ctx context.Context, sizeAtMost int, nestedObjectOptions []fwtypes.NestedObjectOfOptionsFunc[T], planModifiers ...planmodifier.List) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType:    fwtypes.NewListNestedObjectTypeOf(ctx, nestedObjectOptions...),
		Optional:      true,
		Computed:      true,
		PlanModifiers: planModifiers,
		Validators: []validator.List{
			listvalidator.SizeAtMost(sizeAtMost),
		},
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}

func ResourceOptionalComputedListOfObjectsAttribute2[T any](ctx context.Context, optFns ...ResourceListOfObjectsOptionsFunc[T]) schema.ListAttribute {
	opts := newResourceListOfObjectsOptions(optFns...)

	return schema.ListAttribute{
		CustomType:    fwtypes.NewListNestedObjectTypeOf(ctx, opts.nestedObjectOptions...),
		Optional:      true,
		Computed:      true,
		PlanModifiers: opts.planModifiers,
		Validators:    opts.validators,
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}

// ResourceOptionalComputedSingleNestedObjectAttribute returns a schema attribute that
// is equivalent to a Terraform Plugin SDKv2 Optional+Computed list nested block with a maximum size of 1.
func ResourceOptionalComputedSingleNestedObjectAttribute[T any](ctx context.Context) schema.ListAttribute {
	return ResourceOptionalComputedListOfObjectsAttribute2(ctx, WithValidators[T](listvalidator.SizeAtMost(1)), WithPlanModifiers[T](listplanmodifier.UseStateForUnknown()))
}

// ResourceOptionalComputedForceNewSingleNestedObjectAttribute returns a schema attribute that
// is equivalent to a Terraform Plugin SDKv2 Optional+Computed+ForceNew list nested block with a maximum size of 1.
func ResourceOptionalComputedForceNewSingleNestedObjectAttribute[T any](ctx context.Context) schema.ListAttribute {
	return ResourceOptionalComputedListOfObjectsAttribute2(ctx, WithValidators[T](listvalidator.SizeAtMost(1)), WithPlanModifiers[T](listplanmodifier.RequiresReplaceIfConfigured(), listplanmodifier.UseStateForUnknown()))
}

type resourceListOfObjectsOptions[T any] struct {
	nestedObjectOptions []fwtypes.NestedObjectOfOptionsFunc[T]
	planModifiers       []planmodifier.List
	validators          []validator.List
}

func newResourceListOfObjectsOptions[T any](optFns ...ResourceListOfObjectsOptionsFunc[T]) resourceListOfObjectsOptions[T] {
	var opts resourceListOfObjectsOptions[T]
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

type ResourceListOfObjectsOptionsFunc[T any] func(*resourceListOfObjectsOptions[T])

// WithNestedObjectOptions sets the list of nested object options.
//
// Use this option to fully overwrite the nested object options list. To preserve
// preexisting items, use WithNestedObjectOptionsAppend instead.
func WithNestedObjectOptions[T any](nestedObjectOptions ...fwtypes.NestedObjectOfOptionsFunc[T]) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.nestedObjectOptions = slices.Clone(nestedObjectOptions)
	}
}

// WithNestedObjectOptionsAppend appends to the list of nested object options.
//
// Use this option to preserve preexisting items in the nested object options list.
func WithNestedObjectOptionsAppend[T any](nestedObjectOptions ...fwtypes.NestedObjectOfOptionsFunc[T]) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.nestedObjectOptions = append(o.nestedObjectOptions, nestedObjectOptions...)
	}
}

// WithPlanModifiers sets the list of plan modifiers.
//
// Use this option to fully overwrite the plan modifiers list. To preserve
// preexisting items, use WithPlanModifiersAppend instead.
func WithPlanModifiers[T any](planModifiers ...planmodifier.List) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.planModifiers = slices.Clone(planModifiers)
	}
}

// WithPlanModifiersAppend appends to the list of plan modifiers.
//
// Use this option to preserve preexisting items in the plan modifiers list.
func WithPlanModifiersAppend[T any](planModifiers ...planmodifier.List) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.planModifiers = append(o.planModifiers, planModifiers...)
	}
}

// WithValidators sets the list of validators.
//
// Use this option to fully overwrite the validators list. To preserve
// preexisting items, use WithValidatorsAppend instead.
func WithValidators[T any](validators ...validator.List) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.validators = slices.Clone(validators)
	}
}

// WithValidatorsAppend appends to the list of validators.
//
// Use this option to preserve preexisting items in the validator list.
func WithValidatorsAppend[T any](validators ...validator.List) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.validators = append(o.validators, validators...)
	}
}
