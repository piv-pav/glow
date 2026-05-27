package article

import (
	"fmt"
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
