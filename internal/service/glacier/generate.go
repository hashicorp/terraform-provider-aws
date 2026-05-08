// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForVault -ListTagsInIDElem=VaultName -ServiceTagsMap -KVTValues -TagOp=AddTagsToVault -TagInIDElem=VaultName -UntagOp=RemoveTagsFromVault -UpdateTags -CreateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package glacier
