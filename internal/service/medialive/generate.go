//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -KVTValues=true -SkipTypesImp=true -ListTags -ServiceTagsMap -TagOp=CreateTags -UntagOp=DeleteTags -UpdateTags
//go:generate go run ../../generate/servicepackagedata/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package medialive
