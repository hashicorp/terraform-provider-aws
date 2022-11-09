//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -ListTags -UpdateTags -AWSSDKVersion=2
//go:generate go run ./test-fixtures/generate/document_classifier/main.go
//go:generate go run ./test-fixtures/generate/entity_recognizer/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package comprehend
