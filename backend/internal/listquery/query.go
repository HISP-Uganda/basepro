package listquery

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseInt(raw string, defaultValue int, min int, max int, fieldName string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", fieldName)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
	}

	return value, nil
}

func ParseSort(raw string) (field string, order string, err error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", "", nil
	}

	parts := strings.Split(value, ":")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("sort must use field:asc|desc")
	}

	field = strings.TrimSpace(parts[0])
	if field == "" {
		return "", "", fmt.Errorf("sort field is required")
	}
	if len(parts) == 1 {
		return field, "asc", nil
	}

	order = strings.ToLower(strings.TrimSpace(parts[1]))
	if order != "asc" && order != "desc" {
		return "", "", fmt.Errorf("sort order must be asc or desc")
	}
	return field, order, nil
}

func ParseFilter(raw string) (field string, value string, err error) {
	filter := strings.TrimSpace(raw)
	if filter == "" {
		return "", "", nil
	}

	parts := strings.Split(filter, ":")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("filter must use field:value")
	}
	if len(parts) == 1 {
		value = strings.TrimSpace(parts[0])
		if value == "" {
			return "", "", nil
		}
		return "", value, nil
	}

	field = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])
	if field == "" || value == "" {
		return "", "", fmt.Errorf("filter must use field:value")
	}
	return field, value, nil
}

func ResolveSearch(q string, filterField string, filterValue string, allowedFilterFields map[string]struct{}) string {
	search := strings.TrimSpace(q)
	if search != "" {
		return search
	}
	if strings.TrimSpace(filterValue) == "" {
		return ""
	}
	if filterField == "" {
		return strings.TrimSpace(filterValue)
	}
	if _, ok := allowedFilterFields[filterField]; ok {
		return strings.TrimSpace(filterValue)
	}
	return ""
}
