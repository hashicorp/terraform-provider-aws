package comprehend

// Takes the result of flatmap.Expand for an array of strings
// and returns a []string
func ExpandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, val)
		}
	}
	return vs
}

// Takes list of strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSet
func FlattenStringList(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}
