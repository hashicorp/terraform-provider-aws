// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -UpdateTags -SkipTypesImp -KVTValues
//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -ListTags -TagType TagEntry -KeyValueTagsFunc KeyValueTagsSlice -TagsFunc TagsSlice -GetTagsInFunc getTagsInSlice -SetTagsOutFunc setTagsOutSlice tags_slice_gen.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ssmquicksetup
