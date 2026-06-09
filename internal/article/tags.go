package article

import "strings"

// GetTags returns the article's tags as a string slice.
func (a *Article) GetTags() []string {
	val, ok := a.Frontmatter["tags"]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

// SetTags replaces the tags field with the given values.
// Accepts comma-separated values within each string.
func (a *Article) SetTags(tags ...string) {
	flat := flattenTags(tags)
	if len(flat) > 0 {
		a.Frontmatter["tags"] = flat
	}
}

// AddTags appends tags to the existing tags field.
// Accepts comma-separated values within each string.
// Deduplicates against existing tags.
func (a *Article) AddTags(tags ...string) {
	flat := flattenTags(tags)
	if len(flat) == 0 {
		return
	}

	existing := a.GetTags()
	seen := make(map[string]bool, len(existing))
	for _, t := range existing {
		seen[t] = true
	}

	for _, t := range flat {
		if !seen[t] {
			existing = append(existing, t)
			seen[t] = true
		}
	}

	a.Frontmatter["tags"] = existing
}

// RemoveTags removes specified tags from the tags field.
// Accepts comma-separated values within each string.
// Deletes the tags field entirely if empty after removal.
func (a *Article) RemoveTags(tags ...string) {
	toRemove := flattenTags(tags)
	if len(toRemove) == 0 {
		return
	}

	removeSet := make(map[string]bool, len(toRemove))
	for _, t := range toRemove {
		removeSet[t] = true
	}

	existing := a.GetTags()
	if len(existing) == 0 {
		return
	}

	var kept []string
	for _, t := range existing {
		if !removeSet[t] {
			kept = append(kept, t)
		}
	}

	if len(kept) == 0 {
		delete(a.Frontmatter, "tags")
	} else {
		a.Frontmatter["tags"] = kept
	}
}

// flattenTags splits comma-separated tag strings into individual trimmed tags.
func flattenTags(tags []string) []string {
	var flat []string
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			p := strings.TrimSpace(part)
			if p != "" {
				flat = append(flat, p)
			}
		}
	}
	return flat
}
