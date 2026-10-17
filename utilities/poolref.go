package utilities

import "regexp"

// uuidRE matches standard 8-4-4-4-12 hex UUID form (case-insensitive).
var uuidRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// LooksLikeUUID reports whether ref is a UUID string (virsh-style pool ref).
// UUID-shaped strings always take the UUID lookup path; do not fall back to name.
func LooksLikeUUID(ref string) bool {
	if ref == "" {
		return false
	}
	return uuidRE.MatchString(ref)
}
