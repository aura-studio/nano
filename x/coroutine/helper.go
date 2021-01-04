package coroutine

func interfaceSliceToStringSlice(data []interface{}) []string {
	strs := make([]string, 0)
	for _, i := range data {
		strs = append(strs, i.(string))
	}
	return strs
}
