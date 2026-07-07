package utils

import (
	"fmt"
	"net"
	"strings"
)

// ValidateEmailDomain checks that the domain part of email has DNS records
// capable of receiving mail: an MX record, falling back to A/AAAA per the
// implicit-MX rule in RFC 5321. It only proves the domain is real and mail-
// configured, not that the specific mailbox exists.
func ValidateEmailDomain(email string) error {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return fmt.Errorf("invalid email address %q", email)
	}
	domain := email[at+1:]

	if mxRecords, err := net.LookupMX(domain); err == nil && len(mxRecords) > 0 {
		return nil
	}
	if _, err := net.LookupHost(domain); err == nil {
		return nil
	}
	return fmt.Errorf("email domain %q does not resolve to a mail server", domain)
}
