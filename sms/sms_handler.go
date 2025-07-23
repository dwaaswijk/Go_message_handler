package sms

import (
	"strings"
)

// ValidatePhone checks if the phone number is in a valid E.164 format
func ValidatePhone(phone string) bool {
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	return strings.HasPrefix(phone, "+")
}

// maskPhone obfuscates the phone number for logging
func maskPhone(phone string) string {
	if len(phone) > 4 {
		return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
	}
	return "****"
}
