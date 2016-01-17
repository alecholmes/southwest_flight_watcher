package client

import (
	"fmt"
	"regexp"
	"strings"
)

var airportCodeRegex = regexp.MustCompile("^[A-Z]{3}$")

func normalizeAirportCode(code string) (string, error) {
	normalizedCode := strings.ToUpper(code)
	if airportCodeRegex.MatchString(normalizedCode) {
		return normalizedCode, nil
	} else {
		return "", fmt.Errorf("Invalid airport code %s", code)
	}
}
