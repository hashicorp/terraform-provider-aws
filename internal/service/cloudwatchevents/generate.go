//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=ListEventBuses,ListRules,ListTargetsByRule -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceARN -ServiceTagsSlice=yes -TagInIDElem=ResourceARN -UpdateTags=yes

package cloudwatchevents
