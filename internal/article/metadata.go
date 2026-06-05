package article

import (
	"fmt"
	"strings"
)

// SetMetadata sets a scalar metadata field (overwrites if exists)
// Note: 'created' field is protected and cannot be overwritten
func (a *Article) SetMetadata(key string, value string) {
	if key == "created" {
		// Protect created timestamp from being overwritten
		if _, exists := a.Metadata["created"]; exists {
			return
		}
	}
	a.Metadata[key] = value
}

// AddMetadata adds value to array field (creates array if doesn't exist)
func (a *Article) AddMetadata(key string, values ...string) error {
	existing, exists := a.Metadata[key]
	
	if !exists {
		// Create new array
		a.Metadata[key] = values
		return nil
	}

	// Handle existing value
	switch v := existing.(type) {
	case []interface{}:
		// Already an array of interface{}
		for _, val := range values {
			v = append(v, val)
		}
		a.Metadata[key] = v
	case []string:
		// Array of strings
		v = append(v, values...)
		a.Metadata[key] = v
	case string:
		// Convert scalar to array
		arr := []string{v}
		arr = append(arr, values...)
		a.Metadata[key] = arr
	default:
		return fmt.Errorf("field %s has unsupported type for add operation", key)
	}

	return nil
}

// DeleteMetadata removes value from array or deletes entire field
// Note: 'created' field is protected and cannot be deleted
func (a *Article) DeleteMetadata(key string, value string) error {
	if key == "created" {
		return fmt.Errorf("cannot delete protected field: created")
	}

	existing, exists := a.Metadata[key]
	
	if !exists {
		return fmt.Errorf("field %s does not exist", key)
	}

	// If no value specified, delete entire field
	if value == "" {
		delete(a.Metadata, key)
		return nil
	}

	// Remove from array
	switch v := existing.(type) {
	case []interface{}:
		newArr := make([]interface{}, 0)
		for _, item := range v {
			if str, ok := item.(string); ok && str != value {
				newArr = append(newArr, item)
			} else if !ok {
				newArr = append(newArr, item)
			}
		}
		if len(newArr) == 0 {
			delete(a.Metadata, key)
		} else {
			a.Metadata[key] = newArr
		}
	case []string:
		newArr := make([]string, 0)
		for _, item := range v {
			if item != value {
				newArr = append(newArr, item)
			}
		}
		if len(newArr) == 0 {
			delete(a.Metadata, key)
		} else {
			a.Metadata[key] = newArr
		}
	case string:
		// If it's a scalar and matches, delete it
		if v == value {
			delete(a.Metadata, key)
		} else {
			return fmt.Errorf("value %s not found in field %s", value, key)
		}
	default:
		return fmt.Errorf("field %s has unsupported type for delete operation", key)
	}

	return nil
}

// GetMetadataString returns metadata field as string
func (a *Article) GetMetadataString(key string) (string, bool) {
	if val, ok := a.Metadata[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetMetadataArray returns metadata field as string array
func (a *Article) GetMetadataArray(key string) ([]string, bool) {
	if val, ok := a.Metadata[key]; ok {
		switch v := val.(type) {
		case []string:
			return v, true
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result, true
		}
	}
	return nil, false
}

// SetTags replaces the tags field with the given values.
// Accepts comma-separated values within each string.
func (a *Article) SetTags(tags ...string) {
	var flat []string
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			t = strings.TrimSpace(part)
			if t != "" {
				flat = append(flat, t)
			}
		}
	}
	if len(flat) > 0 {
		a.Metadata["tags"] = flat
	}
}

// AddTags appends tags to the existing tags field.
// Accepts comma-separated values within each string.
func (a *Article) AddTags(tags ...string) {
	var flat []string
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			t = strings.TrimSpace(part)
			if t != "" {
				flat = append(flat, t)
			}
		}
	}
	if len(flat) == 0 {
		return
	}

	existing, ok := a.GetMetadataArray("tags")
	if ok {
		a.Metadata["tags"] = append(existing, flat...)
	} else {
		a.Metadata["tags"] = flat
	}
}

// RemoveTags removes specified tags from the tags field.
// Accepts comma-separated values within each string.
// Deletes the tags field entirely if empty after removal.
func (a *Article) RemoveTags(tags ...string) {
	var toRemove []string
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			t = strings.TrimSpace(part)
			if t != "" {
				toRemove = append(toRemove, t)
			}
		}
	}
	if len(toRemove) == 0 {
		return
	}

	removeSet := make(map[string]bool, len(toRemove))
	for _, t := range toRemove {
		removeSet[t] = true
	}

	existing, ok := a.GetMetadataArray("tags")
	if !ok {
		return
	}

	var kept []string
	for _, t := range existing {
		if !removeSet[t] {
			kept = append(kept, t)
		}
	}

	if len(kept) == 0 {
		delete(a.Metadata, "tags")
	} else {
		a.Metadata["tags"] = kept
	}
}
