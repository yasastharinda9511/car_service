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

func IsValidDocumentType(mimeType string) bool {
	validTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}
