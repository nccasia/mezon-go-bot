package helper

import (
	"strings"
)

func ExtractMessage(message string) (string, []string) {
	message = strings.ReplaceAll(message, "\n", " ")
	message = strings.TrimSpace(message[len("*"):])

	args := strings.Fields(message)
	if len(args) > 0 {
		command := strings.ToLower(args[0])
		return command, args[1:]
	}
	return "", []string{}
}
