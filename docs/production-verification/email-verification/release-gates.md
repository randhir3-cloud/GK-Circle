# Email Verification Release Gates

- [ ] Verification CI passes for an immutable 40-character commit SHA.
- [ ] Staging deploy uses that exact SHA.
- [ ] Staging health and verification smoke checks pass.
- [ ] Rollback is executed in staging and its evidence identifier is recorded.
- [ ] The staging artifact contains the exact SHA and rollback evidence.
- [ ] At least 24 hours have elapsed since the successful staging run completed.
- [ ] Production GitHub environment approval is recorded.
- [ ] Production workflow validates the same SHA before deployment.
- [ ] Production health smoke check passes.
- [ ] Five-milestone production audit remains separate and uses a uniquely tagged QA identity.

Required GitHub environment configuration:

- `RAILWAY_TOKEN` secret in both `staging` and `production` environments.
- `RAILWAY_PROJECT_ID`, environment names, health URLs, and `RAILWAY_SERVICES_JSON` repository/environment variables.
- Required reviewers on the `production` environment.
