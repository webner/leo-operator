package utils

// ValueOrDefault returns the string value if not empty otherwise the def string
func ValueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}
