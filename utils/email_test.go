package utils

import "testing"

func TestValidateEmailDomainRejectsMalformed(t *testing.T) {
	for _, email := range []string{"", "no-at-sign", "trailing-at@"} {
		if err := ValidateEmailDomain(email); err == nil {
			t.Errorf("ValidateEmailDomain(%q) = nil, want error", email)
		}
	}
}

// These two hit real DNS and need network access — skip if it's unavailable
// rather than failing the whole suite.
func TestValidateEmailDomainRealDomain(t *testing.T) {
	if err := ValidateEmailDomain("someone@gmail.com"); err != nil {
		t.Skipf("network/DNS unavailable, skipping: %v", err)
	}
}

func TestValidateEmailDomainRejectsNonexistentDomain(t *testing.T) {
	if ValidateEmailDomain("someone@gmail.com") != nil {
		t.Skip("network/DNS unavailable, skipping")
	}
	if err := ValidateEmailDomain("someone@definitely-not-a-real-domain-xyz123456789.com"); err == nil {
		t.Error("ValidateEmailDomain accepted a domain with no DNS records")
	}
}
