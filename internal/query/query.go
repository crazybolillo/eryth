package query

import (
	"net/url"
	"strconv"
)

func GetIntOr(values url.Values, name string, defaultValue int) (int, error) {
	value := values.Get(name)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, err
	}

	return intValue, nil
}
