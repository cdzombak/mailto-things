package main

import (
	"strings"

	"google.golang.org/api/gmail/v1"
)

// MessageSubject returns the content of the Subject header of the given
// Gmail message, if it has one; else the empty string.
func MessageSubject(message *gmail.Message) string {
	for _, h := range message.Payload.Headers {
		if strings.ToLower(h.Name) == "subject" {
			return h.Value
		}
	}
	return ""
}
