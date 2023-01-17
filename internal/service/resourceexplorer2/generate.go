//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -TagInIDElem=ResourceArn -ListTags -ListTagsInIDElem=ResourceArn -ServiceTagsMap -UpdateTags -UntagInTagsElem=TagKeys -KVTValues -SkipTypesImp
//go:generate go run ../../generate/servicepackagedata/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package resourceexplorer2
