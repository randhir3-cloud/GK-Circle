package cli

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

const testIdentityID = "8e5ca187-848d-43f6-a011-7c20d9890d71"
const testRunID = "e2e12345"

type fakeQAAdmin struct {
	identity qaKratosIdentity
	getErr   error
	deleted  bool
}

func (f *fakeQAAdmin) GetIdentity(context.Context, string) (qaKratosIdentity, error) {
	return f.identity, f.getErr
}

func (f *fakeQAAdmin) DeleteIdentity(context.Context, string) error {
	f.deleted = true
	return nil
}

type fakeQAUsers struct {
	user models.User
	err  error
}

func (f fakeQAUsers) GetUserByKratosID(string) (models.User, error) {
	return f.user, f.err
}

type fakeQADeleter struct {
	deleted bool
}

func (f *fakeQADeleter) DeleteUserDataById(string, string) error {
	f.deleted = true
	return nil
}

func taggedIdentity() qaKratosIdentity {
	identity := qaKratosIdentity{ID: testIdentityID}
	identity.Traits.Email = "qa+" + testRunID + "@example.test"
	return identity
}

func TestVerificationCleanupDryRunDoesNotDelete(t *testing.T) {
	admin := &fakeQAAdmin{identity: taggedIdentity()}
	deleter := &fakeQADeleter{}
	runner := verificationCleanupRunner{
		admin: admin,
		users: fakeQAUsers{user: models.User{
			ID:    "application-user",
			Email: "qa+" + testRunID + "@example.test",
		}},
		deleter: deleter,
	}

	if err := runner.Run(context.Background(), testIdentityID, testRunID, false); err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if admin.deleted || deleter.deleted {
		t.Fatal("dry-run performed a deletion")
	}
}

func TestVerificationCleanupUsesApplicationService(t *testing.T) {
	admin := &fakeQAAdmin{identity: taggedIdentity()}
	deleter := &fakeQADeleter{}
	runner := verificationCleanupRunner{
		admin: admin,
		users: fakeQAUsers{user: models.User{
			ID:    "application-user",
			Email: "qa+" + testRunID + "@example.test",
		}},
		deleter: deleter,
	}

	if err := runner.Run(context.Background(), testIdentityID, testRunID, true); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !deleter.deleted {
		t.Fatal("application service was not used")
	}
	if admin.deleted {
		t.Fatal("runner duplicated the Kratos deletion")
	}
}

func TestVerificationCleanupDeletesKratosOnlyIdentity(t *testing.T) {
	admin := &fakeQAAdmin{identity: taggedIdentity()}
	runner := verificationCleanupRunner{
		admin:   admin,
		users:   fakeQAUsers{err: sql.ErrNoRows},
		deleter: &fakeQADeleter{},
	}

	if err := runner.Run(context.Background(), testIdentityID, testRunID, true); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !admin.deleted {
		t.Fatal("Kratos-only identity was not deleted")
	}
}

func TestVerificationCleanupRejectsMismatchedRunTag(t *testing.T) {
	admin := &fakeQAAdmin{identity: taggedIdentity()}
	runner := verificationCleanupRunner{
		admin:   admin,
		users:   fakeQAUsers{err: sql.ErrNoRows},
		deleter: &fakeQADeleter{},
	}

	err := runner.Run(context.Background(), testIdentityID, "another123", true)
	if err == nil {
		t.Fatal("mismatched run tag was accepted")
	}
	if admin.deleted {
		t.Fatal("mismatched identity was deleted")
	}
}

func TestVerificationCleanupIsIdempotentWhenBothRecordsAreAbsent(t *testing.T) {
	runner := verificationCleanupRunner{
		admin:   &fakeQAAdmin{getErr: errKratosIdentityNotFound},
		users:   fakeQAUsers{err: sql.ErrNoRows},
		deleter: &fakeQADeleter{},
	}

	if err := runner.Run(context.Background(), testIdentityID, testRunID, true); err != nil {
		t.Fatalf("idempotent cleanup failed: %v", err)
	}
}

func TestEmailContainsExactRunTag(t *testing.T) {
	if !emailContainsExactRunTag("qa+"+testRunID+"@example.test", testRunID) {
		t.Fatal("valid run tag was rejected")
	}
	if emailContainsExactRunTag("qa."+testRunID+"@example.test", testRunID) {
		t.Fatal("non-plus run tag was accepted")
	}
	if emailContainsExactRunTag("qa+"+testRunID+"-extra@example.test", testRunID) {
		t.Fatal("run tag was accepted as a partial suffix")
	}
}

func TestValidateQAAdminURLRejectsPublicProductionHosts(t *testing.T) {
	for _, rawURL := range []string{
		"https://gkcircle.com/kratos-admin",
		"https://kratos.gkcircle.com",
		"https://kratos-production.up.railway.app",
	} {
		if _, err := validateQAAdminURL(rawURL, "production"); err == nil ||
			!strings.Contains(err.Error(), "private Kratos Admin API") {
			t.Fatalf("public production admin URL was accepted: %s", rawURL)
		}
	}

	if _, err := validateQAAdminURL("http://kratos.railway.internal:4434", "production"); err != nil {
		t.Fatalf("private production admin URL was rejected: %v", err)
	}
	if _, err := validateQAAdminURL("http://localhost:4434", "testing"); err != nil {
		t.Fatalf("local testing admin URL was rejected: %v", err)
	}
}
