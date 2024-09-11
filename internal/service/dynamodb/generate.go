// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -GetTag -ListTags -ListTagsOp=ListTagsOfResource -ServiceTagsSlice -UpdateTags -ParentNotFoundErrCode=ResourceNotFoundException
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListBackups -InputPaginator=ExclusiveStartBackupArn -OutputPaginator=LastEvaluatedBackupArn -- list_backups_pages_gen.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dynamodb
