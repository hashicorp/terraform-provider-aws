// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=ListTagsOfResource -ServiceTagsSlice -UpdateTags -ParentNotFoundErrCode=ResourceNotFoundException
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dynamodb
