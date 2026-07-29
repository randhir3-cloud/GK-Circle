package kratos

import (
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
)

// makeUser builds a minimal KratosUserDetails for testing.
func makeUser(primaryEmail string, addresses ...struct {
	value    string
	verified bool
}) config.KratosUserDetails {
	user := config.KratosUserDetails{}
	user.Identity.Traits.Email = primaryEmail
	for _, a := range addresses {
		user.Identity.VerifiableAddresses = append(
			user.Identity.VerifiableAddresses,
			struct {
				Value    string `json:"value"`
				Verified bool   `json:"verified"`
			}{Value: a.value, Verified: a.verified},
		)
	}
	return user
}

func TestIsEmailVerified_ExplicitTrue(t *testing.T) {
	user := makeUser("user@example.com", struct {
		value    string
		verified bool
	}{"user@example.com", true})
	if !IsEmailVerified(user) {
		t.Fatal("expected verified=true for explicit verified primary address")
	}
}

func TestIsEmailVerified_ExplicitFalse(t *testing.T) {
	user := makeUser("user@example.com", struct {
		value    string
		verified bool
	}{"user@example.com", false})
	if IsEmailVerified(user) {
		t.Fatal("expected verified=false when primary address has verified=false")
	}
}

func TestIsEmailVerified_EmptyAddresses(t *testing.T) {
	user := makeUser("user@example.com")
	if IsEmailVerified(user) {
		t.Fatal("expected verified=false when VerifiableAddresses is empty")
	}
}

func TestIsEmailVerified_NoMatchingAddress(t *testing.T) {
	user := makeUser("primary@example.com", struct {
		value    string
		verified bool
	}{"other@example.com", true})
	if IsEmailVerified(user) {
		t.Fatal("expected verified=false when no address matches primary email")
	}
}

func TestIsEmailVerified_EmptyPrimaryEmail(t *testing.T) {
	user := makeUser("", struct {
		value    string
		verified bool
	}{"user@example.com", true})
	if IsEmailVerified(user) {
		t.Fatal("expected verified=false when primary email is empty")
	}
}

func TestIsEmailVerified_ZeroValue(t *testing.T) {
	if IsEmailVerified(config.KratosUserDetails{}) {
		t.Fatal("expected verified=false for zero-value KratosUserDetails")
	}
}

func TestIsEmailVerified_PrimaryEmailHasWhitespace(t *testing.T) {
	// Traits.Email has leading/trailing whitespace; address does not.
	user := makeUser("  user@example.com  ", struct {
		value    string
		verified bool
	}{"user@example.com", true})
	if !IsEmailVerified(user) {
		t.Fatal("expected verified=true when primary email matches after whitespace normalisation")
	}
}

func TestIsEmailVerified_CaseInsensitive(t *testing.T) {
	// Traits.Email mixed case; address lower case.
	user := makeUser("User@Example.COM", struct {
		value    string
		verified bool
	}{"user@example.com", true})
	if !IsEmailVerified(user) {
		t.Fatal("expected verified=true for case-insensitive match")
	}
}

func TestIsEmailVerified_VerifiedSecondaryUnverifiedPrimary(t *testing.T) {
	// Secondary address is verified; primary is not.
	user := makeUser("primary@example.com", struct {
		value    string
		verified bool
	}{"other@example.com", true}, struct {
		value    string
		verified bool
	}{"primary@example.com", false})
	if IsEmailVerified(user) {
		t.Fatal("expected verified=false when primary address is unverified (verified secondary should not count)")
	}
}

func TestIsEmailVerified_VerifiedPrimaryUnverifiedSecondary(t *testing.T) {
	// Primary address is verified; secondary is not.
	user := makeUser("primary@example.com", struct {
		value    string
		verified bool
	}{"primary@example.com", true}, struct {
		value    string
		verified bool
	}{"other@example.com", false})
	if !IsEmailVerified(user) {
		t.Fatal("expected verified=true when primary address is verified")
	}
}
