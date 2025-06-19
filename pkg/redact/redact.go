package redact

func RedactString(input string) string {
	if len(input) <= 4 {
		return "****"
	}

	lastFour := input[len(input)-4:]

	return "****" + lastFour
}

