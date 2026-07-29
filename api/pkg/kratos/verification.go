package kratos

import (
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
)

// IsEmailVerified reports whether the primary email address on the Kratos
// identity is confirmed verified. Comparison is case-insensitive and
// whitespace-normalised. Fails closed for all ambiguous input — anything
// other than an explicit verified=true on the matching primary address
// returns false.
//
// Fail-closed cases:
//   - Identity.Traits.Email is empty
//   - VerifiableAddresses is nil or empty
//   - No address matches the primary email (case-insensitive, trimmed)
//   - The matching address has Verified = false
//   - Zero-value config.KratosUserDetails (all fields empty)
func IsEmailVerified(user config.KratosUserDetails) bool {
	primary := strings.ToLower(strings.TrimSpace(user.Identity.Traits.Email))
	if primary == "" {
		return false
	}

	for _, addr := range user.Identity.VerifiableAddresses {
		candidate := strings.ToLower(strings.TrimSpace(addr.Value))
		if candidate == primary {
			// Only explicit true counts — zero value of bool is false.
			return addr.Verified
		}
	}

	// No matching address found.
	return false
}
