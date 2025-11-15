package util

func HasPermission(allPermissions []string, permission string) bool {
	for _, perm := range allPermissions {
		if perm == permission {
			return true
		}
	}
	return false
}
