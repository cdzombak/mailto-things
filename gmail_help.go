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

// PartCID returns the content of the Content-ID header of the given
// Gmail message part, if it has one; else the empty string.
func PartCID(part *gmail.MessagePart) string {
	for _, h := range part.Headers {
		if strings.ToLower(h.Name) == "content-id" {
			v := h.Value
			v = strings.TrimPrefix(v, "<")
			v = strings.TrimSuffix(v, ">")
			return v
		}
	}
	return ""
}
