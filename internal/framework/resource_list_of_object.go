// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
func ResourceOptionalComputedListOfObjectsAttribute[T any](ctx context.Context, sizeAtMost int, nestedObjectOptions []fwtypes.NestedObjectOfOption[T], planModifiers ...planmodifier.List) schema.ListAttribute {
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

type resourceListOfObjectsOptions[T any] struct {
	nestedObjectOptions []fwtypes.NestedObjectOfOption[T]
	planModifiers       []planmodifier.List
	validators          []validator.List
}

type ResourceListOfObjectsOptionsFunc[T any] func(*resourceListOfObjectsOptions[T])

// WithNestedObjectOptions sets the list of nested object options.
//
// Use this option to fully overwrite the nested object options list. To preserve
// preexisting items, use WithNestedObjectOptionsAppend instead.
func WithNestedObjectOptions[T any](nestedObjectOptions ...fwtypes.NestedObjectOfOption[T]) ResourceListOfObjectsOptionsFunc[T] {
	return func(o *resourceListOfObjectsOptions[T]) {
		o.nestedObjectOptions = nestedObjectOptions
	}
}

// WithNestedObjectOptionsAppend appends to the list of nested object options.
//
// Use this option to preserve preexisting items in the nested object options list.
func WithNestedObjectOptionsAppend[T any](nestedObjectOptions ...fwtypes.NestedObjectOfOption[T]) ResourceListOfObjectsOptionsFunc[T] {
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
		o.planModifiers = planModifiers
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
		o.validators = validators
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
