package utilities

import "github.com/google/uuid"

func UniqueStringSlice(slice []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func InlineIf(condition bool, trueValue any, falseValue any) any {
	if condition {
		return trueValue
	}
	return falseValue
}

func IsUuidValid(value string) bool {
	return uuid.Validate(value) == nil
}
