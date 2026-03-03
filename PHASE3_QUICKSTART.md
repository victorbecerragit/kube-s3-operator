# Phase 3 Quick Start Guide

**Get Started with Phase 3 in 5 minutes**

---

## What is Phase 3?

Phase 3 implements **automated testing** through GitHub Actions. Every time you push code or create a PR, tests run automatically.

---

## Files You Need to Know

### Core Workflows (Run Automatically)

1. **`.github/workflows/tests.yml`** - Full test suite
   - When: Push to main/develop, PR to main/develop
   - Duration: 4-5 minutes
   - Jobs: Unit tests → KUTTL tests, E2E tests → Summary

2. **`.github/workflows/pr-checks.yml`** - Fast PR validation
   - When: PR to main/develop
   - Duration: < 3 minutes
   - Jobs: Lint → Quick tests → Manifest check

### Documentation (Read These)

3. **`.github/WORKFLOWS.md`** - Complete workflow reference
   - What: Explains every job and step
   - Read when: You're confused about workflows

4. **`docs/PHASE3_IMPLEMENTATION.md`** - Full implementation guide
   - What: Technical details and architecture
   - Read when: You want to understand how it works

5. **`PHASE3_DEPLOYMENT_CHECKLIST.md`** - Deploy Phase 3
   - What: Step-by-step deployment guide
   - Read when: Deploying Phase 3

6. **`PHASE3_SUMMARY.md`** - High-level overview
   - What: Executive summary of Phase 3
   - Read when: You need a quick overview

---

## Timeline: From Code to Deployment

### Day 1: Understanding (30 min)
```
1. Read this file (5 min)
2. Review .github/WORKFLOWS.md (15 min)
3. Check PHASE3_SUMMARY.md (10 min)
```

### Day 1: Local Testing (30 min)
```bash
cd code/
make test                      # Unit tests
gofmt -l . && go vet ./...    # Lint check
make manifests generate       # Manifest check
```

### Day 2: Deployment (2 hours)
```
Follow PHASE3_DEPLOYMENT_CHECKLIST.md
1. Verify files (30 min)
2. Test locally (30 min)
3. Create PR and verify (30 min)
4. Configure branch protection (30 min)
```

### Day 3: Monitoring (30 min)
```
1. Create test PR
2. Watch pr-checks.yml run
3. Monitor execution times
4. Train team
```

---

## How to Use Phase 3

### As a Developer

**Before Committing:**
```bash
# Run quick checks locally
cd code/
make test
gofmt -l . && go vet ./...
make manifests generate
```

**After Pushing:**
```
1. Go to GitHub → Actions tab
2. See your workflow running
3. Check status in 2-3 minutes
4. If failed: Download artifact → Debug → Fix → Push again
```

### As a Reviewer

**On PR:**
```
1. See pr-checks.yml status at bottom of PR
2. All green ✅ = safe to review code
3. Any red ❌ = wait for developer to fix
```

### As a Maintainer

**Daily:**
```
1. Check GitHub Actions dashboard
2. Look for failed workflows
3. Review artifacts if needed
4. Monitor performance trends
```

---

## Common Questions

### Q: When do workflows run?
**A:** 
- `pr-checks.yml`: Immediately on every PR (< 3 min)
- `tests.yml`: On push to main/develop or merge to main (~5 min)

### Q: Why are tests blocked on PR?
**A:** 
GitHub's branch protection requires tests to pass before merge. This ensures code quality.

### Q: How do I debug a failure?
**A:** 
1. Click failed job in GitHub Actions
2. Expand steps to see error
3. For KUTTL: Download `kuttl-cluster-state` artifact
4. Run locally: `kind create cluster && make install && kubectl kuttl test`

### Q: What about caching?
**A:** 
Go modules cache automatically. First run: 30s, future runs: 5s (if no go.sum change).

### Q: Can I skip workflows?
**A:** 
No - they're required. But you can run them locally first to catch issues.

### Q: How long are artifacts kept?
**A:** 
7 days. Download them before they expire.

---

## Troubleshooting

### Workflow Won't Start
**Solution:** Check you pushed to correct branch (main/develop)

### Always Cache Miss
**Solution:** Normal if go.sum changes often. This is expected.

### Tests Timeout
**Solution:** Check WORKFLOWS.md timeout adjustment guide

### Can't See Artifacts
**Solution:** Download from GitHub Actions → [Run] → Artifacts

---

## Key Files Reference

| File | Purpose | Read When |
|------|---------|-----------|
| `.github/workflows/tests.yml` | Full test suite | Debugging tests |
| `.github/workflows/pr-checks.yml` | PR validation | Understanding fast checks |
| `.github/WORKFLOWS.md` | Workflow reference | Need details |
| `docs/PHASE3_IMPLEMENTATION.md` | Full implementation | Want to learn |
| `PHASE3_DEPLOYMENT_CHECKLIST.md` | Deploy guide | Ready to deploy |
| `PHASE3_SUMMARY.md` | Executive summary | Need overview |
| `docs/testing.md` | All testing info | Understanding tests |

---

## Quick Reference

### Workflow Status
```
GitHub → Repository → Actions tab → See all runs and status
```

### View Specific Run
```
GitHub → Actions → [Workflow Name] → [Run Number] → See details
```

### Download Artifacts
```
GitHub → Actions → [Failed Run] → Artifacts → Download
```

### Run Tests Locally
```bash
cd code/
make test                    # Unit tests
make install                 # Install CRDs (needs Kind)
kubectl kuttl test test/kuttl/  # KUTTL tests
```

---

## Deployment Checklist

- [ ] Read this file
- [ ] Read PHASE3_SUMMARY.md
- [ ] Read PHASE3_DEPLOYMENT_CHECKLIST.md
- [ ] Run local tests
- [ ] Create test PR
- [ ] Verify pr-checks.yml passes
- [ ] Merge to develop
- [ ] Verify tests.yml passes on main
- [ ] Configure branch protection
- [ ] Train team

---

## Next Steps

1. **Read:** `PHASE3_SUMMARY.md` (10 min)
2. **Understand:** `.github/WORKFLOWS.md` (20 min)
3. **Deploy:** `PHASE3_DEPLOYMENT_CHECKLIST.md` (2 hours)
4. **Monitor:** First week daily checks

---

## Need Help?

- **Confused about workflows?** → Read `.github/WORKFLOWS.md`
- **Want implementation details?** → Read `docs/PHASE3_IMPLEMENTATION.md`
- **Ready to deploy?** → Follow `PHASE3_DEPLOYMENT_CHECKLIST.md`
- **Test failing?** → Download artifact and review
- **Questions?** → Check section troubleshooting above

---

## What's Next?

After Phase 3 deployment:
- Phase 4: Add Slack notifications
- Phase 5: Security scanning
- Phase 6: Auto-releases

---

**Phase 3: Ready to deploy! 🚀**

Start with `PHASE3_DEPLOYMENT_CHECKLIST.md`
