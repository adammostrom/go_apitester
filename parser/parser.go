package parser

// Takes a slice of paths, and appends them into a recursive map
func setPath(path []string, val any) map[string]any {

	body := map[string]any{}

	current := body

	for i, key := range path {
		if i == len(path)-1 {
			body[key] = val
			break
		}
		next := map[string]any{}
		current[key] = next

		current = next

	}
	return body
}
