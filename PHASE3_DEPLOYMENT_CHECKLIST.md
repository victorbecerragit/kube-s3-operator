# Phase 3: KUTTL for CI/CD - Deployment Checklist

**Purpose:** Final verification checklist before deploying Phase 3 to production

**Date Started:** March 3, 2026  
**Target Completion:** March 4, 2026

---

## Pre-Deployment Verification

### ✅ File Creation Verification

- [ ] `.github/workflows/tests.yml` exists and is valid
  ```bash
  # Verify file exists
  ls -la .github/workflows/tests.yml
  
  # Verify YAML syntax
  yamllint .github/workflows/tests.yml
  ```

- [ ] `.github/workflows/pr-checks.yml` exists and is valid
  ```bash
  # Verify file exists
  ls -la .github/workflows/pr-checks.yml
  
  # Verify YAML syntax
  yamllint .github/workflows/pr-checks.yml
  ```

- [ ] `.github/WORKFLOWS.md` exists and is complete
  ```bash
  # Verify file exists
  ls -la .github/WORKFLOWS.md
  
  # Verify content
  wc -l .github/WORKFLOWS.md  # Should be 400+ lines
  ```

- [ ] `docs/PHASE3_IMPLEMENTATION.md` exists
  ```bash
  # Verify file exists
  ls -la docs/PHASE3_IMPLEMENTATION.md
  ```

### ✅ Documentation Verification

- [ ] All workflow files have descriptive comments
- [ ] README.md mentions CI/CD workflows
- [ ] WORKFLOWS.md explains all triggers
- [ ] PHASE3_IMPLEMENTATION.md is complete
- [ ] All code comments are accurate

---

## Local Testing (Before Merge)

### ✅ Unit Tests Pass Locally

```bash
cd code/
make test
# Expected: PASS ✓
```

- [ ] All unit tests pass
- [ ] No new warnings introduced
- [ ] Coverage reports generated

### ✅ Manifest Generation Works

```bash
cd code/
make manifests generate
git status  # Should show no changes
# Expected: nothing to commit
```

- [ ] `make manifests` generates correctly
- [ ] No uncommitted manifests
- [ ] All CRDs up to date

### ✅ Build Succeeds

```bash
cd code/
make build
# Expected: Build successful
```

- [ ] Docker image builds cleanly
- [ ] No build warnings
- [ ] Binary works locally

---

## Git Repository Setup

### ✅ Create Feature Branch

```bash
git checkout -b phase3/ci-cd-workflows
git pull origin develop  # Update from latest develop
```

- [ ] On feature branch
- [ ] Branch up to date

### ✅ Stage Files for Commit

```bash
git add .github/workflows/tests.yml
git add .github/workflows/pr-checks.yml
git add .github/WORKFLOWS.md
git add docs/PHASE3_IMPLEMENTATION.md
git add PHASE3_DEPLOYMENT_CHECKLIST.md
git status  # Review what's staged
```

- [ ] All new files staged
- [ ] No unintended files staged
- [ ] Status shows clean

### ✅ Commit with Standard Message

```bash
git commit -m "Phase 3: Implement KUTTL for CI/CD

- Add tests.yml with unit, KUTTL, and E2E jobs
- Add pr-checks.yml for fast PR feedback
- Add comprehensive workflow documentation
- Implement Go module caching for performance
- Add artifact collection on test failures
- Configure concurrency and branch protection"
```

- [ ] Commit message follows convention
- [ ] All Phase 3 files included
- [ ] Commit history clean

---

## GitHub Actions Configuration

### ✅ Enable Workflows

**Location:** GitHub Repository → Settings → Actions → General

```bash
# Check current state
https://github.com/victorbecerra/kube-s3-operator/settings/actions
```

- [ ] Workflows are enabled (not disabled)
- [ ] "Allow all actions and reusable workflows" selected
- [ ] Not in beta or restricted mode

### ✅ Configure Permissions

**Location:** Settings → Actions → General → Workflow permissions

```
✅ Read and write permissions
✅ Allow GitHub Actions to create and approve pull requests
```

- [ ] Read/write enabled
- [ ] PR creation enabled (for automated checks)

### ✅ Configure Secrets (If Needed)

**Location:** Settings → Secrets and variables → Actions

Current secrets needed:
- [ ] `CODECOV_TOKEN` (for coverage reporting)
  ```bash
  # Get from: https://codecov.io/account/kube-s3-operator
  # Scope: Can read coverage reports
  ```

---

## First Workflow Test

### ✅ Push Feature Branch

```bash
git push origin phase3/ci-cd-workflows
```

- [ ] Push succeeds without errors
- [ ] GitHub shows branch created

### ✅ GitHub Actions Status

**Check:** GitHub Repository → Actions tab

```bash
# Manual verification
https://github.com/victorbecerra/kube-s3-operator/actions
```

- [ ] Workflows are visible
- [ ] No errors in workflow syntax
- [ ] `pr-checks.yml` shows as available

### ✅ Create Test Pull Request

```bash
# On GitHub:
1. Go to Repository → "Compare & pull request"
2. Base: develop
3. Compare: phase3/ci-cd-workflows
4. Click "Create pull request"
5. Title: "Phase 3: KUTTL for CI/CD Workflows"
6. Description: Paste PR details
```

- [ ] PR created successfully
- [ ] Tests are triggered automatically

### ✅ Monitor PR Checks Workflow

**Expected Behavior:**

| Stage | Status | Duration |
|-------|--------|----------|
| pr-checks.yml starts | 🟡 In Progress | ~3 min |
| lint-and-format | ✅ PASS | ~30s |
| quick-unit-tests | ✅ PASS | ~45s |
| manifest-validation | ✅ PASS | ~30s |
| pr-check-summary | ✅ PASS | ~10s |

**Actions to take:**
- [ ] Wait for pr-checks.yml to complete
- [ ] All checks should pass (4/4 green)
- [ ] If any fail, debug and commit fix

### ✅ Monitor Full Tests Workflow

**Expected Behavior (on regular commit, not PR):**

| Stage | Status | Duration |
|-------|--------|----------|
| unit-tests | ✅ PASS | ~45s |
| kuttl-tests | ✅ PASS | ~2 min |
| e2e-tests | ✅ PASS | ~2 min |
| test-summary | ✅ PASS | ~10s |

---

## Artifact Verification

### ✅ Verify Artifact Collection

**When a test passes:**
- [ ] No artifacts collected (normal)
- [ ] No storage overhead

**When a test would fail (intentional test):**
```bash
# Make a deliberate failure to test artifact collection
# 1. Modify a test to fail
# 2. Commit and push
# 3. Check artifacts in Actions dashboard
# 4. Download and verify cluster state
# 5. Revert the failure
```

- [ ] Artifacts are collected on failure
- [ ] Artifact download works
- [ ] Cluster state is useful for debugging

---

## Documentation Review

### ✅ README.md Updates

**Check:** code/README.md mentions CI/CD

```markdown
## Testing

### Automated Testing (CI/CD)
All tests run automatically via GitHub Actions workflows:
- **PR Checks** (pr-checks.yml): Fast feedback on PRs (~3 min)
- **Full Tests** (tests.yml): Comprehensive suite on merge (~5 min)

See [WORKFLOWS.md](.github/WORKFLOWS.md) for details.
```

- [ ] README links to WORKFLOWS.md
- [ ] CI/CD status documented
- [ ] Examples clear and accurate

### ✅ WORKFLOWS.md Complete

**Check:** .github/WORKFLOWS.md is comprehensive

```bash
# Verify content
grep -c "##" .github/WORKFLOWS.md  # Should have ~10+ sections
grep -c "```" .github/WORKFLOWS.md # Should have examples
```

- [ ] All workflows documented
- [ ] All jobs explained
- [ ] Examples provided
- [ ] Troubleshooting complete

### ✅ PHASE3_IMPLEMENTATION.md Complete

**Check:** docs/PHASE3_IMPLEMENTATION.md

- [ ] Overview section complete
- [ ] All jobs documented
- [ ] Performance metrics included
- [ ] Debugging guide present
- [ ] Next steps outlined

---

## Performance Baseline

### ✅ Record Execution Times

Run and record times for:

**PR Checks Workflow:**
- [ ] Lint & Format: _____ seconds
- [ ] Quick Unit Tests: _____ seconds
- [ ] Manifest Validation: _____ seconds
- [ ] **Total:** _____ seconds (target: <180s)

**Full Tests Workflow:**
- [ ] Unit Tests: _____ seconds
- [ ] KUTTL Tests: _____ seconds
- [ ] E2E Tests: _____ seconds
- [ ] **Total:** _____ seconds (target: <300s)

**Cache Performance:**
- [ ] First run (no cache): _____ minutes
- [ ] Second run (with cache): _____ minutes
- [ ] **Improvement:** _____ % faster

---

## Branch Protection Configuration

### ✅ Configure Branch Protection for `main`

**Location:** Settings → Branches → main

```yaml
Require status checks to pass before merging:
  ✅ Require status checks to pass before merging
  ✅ Require branches to be up to date before merging
  
Status checks to require:
  ☐ Unit Tests (unit-tests) - Check after first successful run
  ☐ KUTTL Tests (kuttl-tests) - Check after first successful run  
  ☐ E2E Tests (e2e-tests) - Check after first successful run
  ☐ PR Check Summary - for PR merges

Require reviews:
  ✅ Require a pull request before merging
  ✅ Require approval reviews before merging
```

- [ ] Status checks selected
- [ ] Dismiss stale reviews on push (if preferred)
- [ ] Include administrators (if needed)

### ✅ Configure Branch Protection for `develop`

Same as main, but with relaxed constraints if preferred:
- [ ] Require at least 1 approval
- [ ] Allow status checks or manual override

---

## Team Communication

### ✅ Notify Team

```markdown
# Announcement

Phase 3: KUTTL for CI/CD is now live! 🚀

## What Changed
- Automatic tests on every PR
- Fast feedback (3 min) before full suite
- Artifact collection for debugging

## What to Know
- PRs won't merge until tests pass
- See WORKFLOWS.md for details
- Local test: `make test`

## Questions?
See docs/PHASE3_IMPLEMENTATION.md
```

- [ ] Message sent to team channel
- [ ] Team read documentation
- [ ] Questions addressed

---

## Rollback Plan (If Needed)

### ✅ Quick Rollback

If workflows cause issues:

```bash
# Option 1: Disable workflows (temporary)
git checkout develop
rm .github/workflows/tests.yml
rm .github/workflows/pr-checks.yml
git push origin develop

# Option 2: Revert entire Phase 3
git revert <commit-hash>
git push origin develop
```

- [ ] Rollback procedure documented
- [ ] Team knows how to execute
- [ ] Impact minimal (workflows only, no code changes)

---

## Final Sign-Off

### ✅ Technical Review

- [ ] All files created correctly
- [ ] No syntax errors
- [ ] Tests pass locally
- [ ] Documentation complete

### ✅ Functional Review

- [ ] PR checks work (<3 min)
- [ ] Full tests work (<5 min)
- [ ] Artifacts collected properly
- [ ] Cache functioning

### ✅ Documentation Review

- [ ] README updated
- [ ] WORKFLOWS.md comprehensive
- [ ] PHASE3_IMPLEMENTATION.md complete
- [ ] Team understands changes

### ✅ Approval

- [ ] Code review approved
- [ ] Tests pass
- [ ] Ready to merge

---

## Merge to Main

### ✅ Merge PR

```bash
# On GitHub:
1. All checks pass ✅
2. All reviews approved ✅
3. Click "Squash and merge"
4. Confirmation message
5. PR closed
```

- [ ] PR merged to develop
- [ ] Workflows execute on main automatically
- [ ] Status checks show green

### ✅ Verify Deployment

```bash
# Check main branch has updates
git checkout main
git pull origin main
ls .github/workflows/  # Should see tests.yml and pr-checks.yml

# Verify workflows on main
https://github.com/victorbecerra/kube-s3-operator/actions
# Should show tests.yml running
```

- [ ] All files on main
- [ ] Workflows execute automatically
- [ ] Status badges green
- [ ] No errors in Actions

---

## Post-Deployment Monitoring

### ✅ First Week Monitoring

Daily checklist (7 days):

- [ ] **Day 1:** All workflows pass on initial merge
- [ ] **Day 2:** PR checks working on new PRs
- [ ] **Day 3:** Monitor performance vs baseline
- [ ] **Day 4:** Check cache hit rates
- [ ] **Day 5:** Review any failures and artifacts
- [ ] **Day 6:** Verify branch protection working
- [ ] **Day 7:** Summary report on success

### ✅ Performance Benchmarking

```bash
# After 1 week, compare:
- Avg PR check time: _____ min
- Avg full test time: _____ min
- Cache hit rate: _____ %
- Failure rate: _____ %
- Most common issues: _____
```

---

## Completion

### ✅ Phase 3 Complete When:

- [x] All workflow files created and committed
- [x] Documentation written and reviewed
- [x] First PR passes all checks
- [x] Team trained on new workflows
- [x] Branch protection configured
- [x] Artifacts working for debugging
- [x] Performance baseline recorded
- [x] Rollback plan documented

### ✅ Sign-off

**Team Lead:** _________________ **Date:** _________  
**Code Reviewer:** _____________ **Date:** _________  
**Approver:** __________________ **Date:** _________

---

## Phase 3 Status

🟢 **READY FOR DEPLOYMENT**

**Next Phase:** Phase 4 - Observability & Alerts

---

## Quick Reference

### During Deployment
```bash
# Check workflow syntax
yamllint .github/workflows/*.yml

# Test locally
cd code && make test && make manifests generate

# Push and verify
git push origin phase3/ci-cd-workflows
# Check GitHub Actions dashboard
```

### Debugging Workflow Issues
```bash
# View workflow logs
https://github.com/victorbecerra/kube-s3-operator/actions

# Test locally
kind create cluster --name test
cd code && make install
kubectl kuttl test test/kuttl/
```

### Getting Artifacts
```
GitHub → Actions → [Failed Run] → Artifacts → Download
```

---

**This checklist ensures Phase 3 is properly deployed and functioning.**
