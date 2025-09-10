package util

func IsValidImageType(mimeType string) bool {
	validTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}
