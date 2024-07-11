package util

import "fmt"

func ThrowErrorIfEmpty(token string) (string, error) {
	if token == EmptyString {
		return "", fmt.Errorf("string is empty")
	}
	return token, nil
}
