// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -SkipTypesImp -ListTags -ListTagsOp=ListQueueTags -ListTagsInIDElem=QueueUrl -ServiceTagsMap -KVTValues -TagOp=TagQueue -TagInIDElem=QueueUrl -UntagOp=UntagQueue -UpdateTags -CreateTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package sqs
