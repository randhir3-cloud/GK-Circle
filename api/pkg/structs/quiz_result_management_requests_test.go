package structs

import "testing"

func TestResultReleasePolicyConstants_Centralised(t *testing.T) {
	policies := []ResultReleasePolicy{
		ResultReleasePolicyImmediate,
		ResultReleasePolicyManual,
		ResultReleasePolicyScheduled,
	}
	expected := map[ResultReleasePolicy]string{
		ResultReleasePolicyImmediate: "IMMEDIATE",
		ResultReleasePolicyManual:    "MANUAL",
		ResultReleasePolicyScheduled: "SCHEDULED",
	}
	if len(policies) != len(expected) {
		t.Fatalf("expected %d centralised policies, got %d", len(expected), len(policies))
	}
	for _, policy := range policies {
		if expected[policy] != string(policy) {
			t.Fatalf("policy constant mismatch for %q", policy)
		}
	}
	if AuditPolicyUpdate != "POLICY_UPDATE" {
		t.Fatalf("unexpected AuditPolicyUpdate: %q", AuditPolicyUpdate)
	}
	if AuditManualRelease != "MANUAL_RELEASE" {
		t.Fatalf("unexpected AuditManualRelease: %q", AuditManualRelease)
	}
}
