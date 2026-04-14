
---
name: dependabot-batch-remediation
description: >
  Use when you have multiple open Dependabot PRs (Go modules or other ecosystems)
  that you want to bundle into a single safe integration branch, validate end-to-end,
  and ship as one Helm/release increment. Covers high-risk version-jump handling,
  smoke-test deployment, and release tagging.
metadata:
  category: technique
  triggers: >
    dependabot PRs, batch dependency update, combine PRs, integration branch,
    go mod tidy, helm release, smoke test, grpc upgrade, aws-sdk upgrade
  repo: kube-s3-operator
  last_used: v0.2.7
---

# Dependabot Batch Remediation

Safely combine N open Dependabot PRs into one integration branch, validate with a
Helm smoke test, and ship as a single release.

> **Why batch instead of merge individually?**
> Each Dependabot PR is tested in isolation. Merging them sequentially can surface
> interaction bugs only visible when all updates coexist. A single integration branch
> surfaces those conflicts once, in a controlled environment.

---

## When to Use

- You have 2+ queued Dependabot PRs you want in the same release
- One PR has a major/risky version jump (e.g., gRPC minor→minor with many commits)
- You need Helm chart + image tag bump to accompany the dependency updates
- You want a smoke test gate before touching `main`

**NOT for:**
- Single Dependabot PR with no dependency interaction risk → merge directly
- Non-Go ecosystems without a `go.mod` workflow

---

## Variables (fill before running)

```bash
# ── Repo & branch ──────────────────────────────────────────────
REPO_ROOT="$(git rev-parse --show-toplevel)"
INTEGRATION_BRANCH="release/next-helm-deps"   # change per release cycle
GO_MODULE_DIR="code"                           # subdirectory containing go.mod

# ── PR numbers & local branch names (edit per batch) ──────────
# Format: "pr-<number>-<short-slug>"
PR_LOW_RISK=("pr-43-smithy-go" "pr-42-aws-config" "pr-44-go-modules-group")
PR_HIGH_RISK="pr-40-grpc"      # largest version jump; merged last

# ── Image ──────────────────────────────────────────────────────
IMG_REPO="victorbecerra/kube-s3-controller"
IMG_TAG="deps-integration-$(date +%Y%m%d%H%M)"

# ── Smoke test ─────────────────────────────────────────────────
TEST_NS="s3-operator-smoke"
HELM_RELEASE="s3-operator-smoke"
CHART_PATH="./code/charts/kube-s3-operator"

# ── Release ────────────────────────────────────────────────────
NEXT_VERSION="0.2.8"           # bump from current Chart.yaml version
```

---

## Phase 1 — Create the Integration Branch

```bash
cd "${REPO_ROOT}"
git fetch origin --prune

# Fetch each Dependabot PR head into a local tracking branch
git fetch origin pull/44/head:pr-44-go-modules-group
git fetch origin pull/43/head:pr-43-smithy-go
git fetch origin pull/42/head:pr-42-aws-config
git fetch origin pull/40/head:pr-40-grpc

# Fresh integration branch from main
git checkout -B "${INTEGRATION_BRANCH}" origin/main
```

---

## Phase 2 — High-Risk PR Isolation Gate

> Run this **before** merging the risky PR into the integration branch.
> If it fails here, exclude it from this batch and ship the rest.

```bash
# Trial branch — test the risky PR alone
git checkout -B "trial/${PR_HIGH_RISK}" origin/main
git merge --no-ff "${PR_HIGH_RISK}" -m "test: isolate ${PR_HIGH_RISK}"

cd "${GO_MODULE_DIR}"
go mod tidy
go test ./...
cd "${REPO_ROOT}"

# Clean up trial branch when done
git checkout "${INTEGRATION_BRANCH}"
git branch -D "trial/${PR_HIGH_RISK}"
```

**Gate decision:**
| Outcome | Action |
|---------|--------|
| ✅ Tests pass | Proceed — include in integration branch |
| ❌ Tests fail | Exclude `${PR_HIGH_RISK}`, proceed with low-risk PRs only |

---

## Phase 3 — Merge All PRs (low-risk first, high-risk last)

```bash
git checkout "${INTEGRATION_BRANCH}"

# Low-risk first (order: smallest → largest change footprint)
git merge --no-ff pr-43-smithy-go       -m "merge: PR #43 smithy-go 1.24.2->1.24.3"
git merge --no-ff pr-42-aws-config      -m "merge: PR #42 aws-sdk-go-v2/config 1.32.13->1.32.14"
git merge --no-ff pr-44-go-modules-group -m "merge: PR #44 go_modules group bump"

# Baseline validation at mid-point
cd "${GO_MODULE_DIR}" && go mod tidy && go test ./... && cd "${REPO_ROOT}"

# High-risk last
git merge --no-ff pr-40-grpc -m "merge: PR #40 grpc 1.72.2->1.79.3"

# Final tidy after all merges
cd "${GO_MODULE_DIR}" && go mod tidy && cd "${REPO_ROOT}"

git push origin "${INTEGRATION_BRANCH}"
```

---

## Phase 4 — Build & Push Operator Image

```bash
cd "${GO_MODULE_DIR}"
make docker-build IMG="${IMG_REPO}:${IMG_TAG}"
docker push "${IMG_REPO}:${IMG_TAG}"
cd "${REPO_ROOT}"

echo "Image pushed: ${IMG_REPO}:${IMG_TAG}"
```

---

## Phase 5 — Helm Smoke Test

### 5a. Prepare namespace and AWS secret

```bash
kubectl create namespace "${TEST_NS}" 2>/dev/null || true

# Idempotent secret apply (won't fail if already exists)
kubectl -n "${TEST_NS}" create secret generic aws-secret \
  --from-literal=aws-access-key-id="${AWS_ACCESS_KEY_ID}" \
  --from-literal=aws-secret-access-key="${AWS_SECRET_ACCESS_KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -
```

> **Tip — pre-existing secret:** If `aws-secret` was created outside Helm, set
> `--set awsCredentials.createSecret=false` to prevent Helm ownership conflicts.

### 5b. Deploy the operator

```bash
helm upgrade --install "${HELM_RELEASE}" "${CHART_PATH}" \
  -n "${TEST_NS}" \
  --set controllerManager.manager.image.repository="${IMG_REPO}" \
  --set controllerManager.manager.image.tag="${IMG_TAG}" \
  --set awsCredentials.createSecret=false   # remove if Helm should own the secret
```

### 5c. Validate controller health

```bash
DEPLOY="$(kubectl -n ${TEST_NS} get deploy -l app.kubernetes.io/name=kube-s3-operator -o name)"
kubectl -n "${TEST_NS}" rollout status "${DEPLOY}" --timeout=180s
kubectl -n "${TEST_NS}" logs "${DEPLOY}" --tail=200
```

### 5d. Smoke-reconcile an S3Bucket CR

```bash
export CR_NAME="smoke-bucket-$(date +%s)"
export BUCKET_NAME="smoke-${USER}-$(date +%s)"

sed \
  -e "s/namespace: s3-acme/namespace: ${TEST_NS}/" \
  -e "s/name: my-bucket-test-acme$/name: ${CR_NAME}/" \
  -e "s/name: my-bucket-test-acme-2025/name: ${BUCKET_NAME}/" \
  code/config/samples/s3_v1alpha1_s3bucket.yaml | kubectl apply -f -

# Watch for Ready
kubectl -n "${TEST_NS}" get s3bucket -w

# Wait gate (non-blocking — || true lets the script continue for log inspection)
kubectl -n "${TEST_NS}" wait \
  --for=jsonpath='{.status.conditions[?(@.type=="Ready")].status}'=True \
  s3bucket/"${CR_NAME}" --timeout=300s || true

# AWS SDK sanity check in logs
kubectl -n "${TEST_NS}" logs "${DEPLOY}" --since=10m \
  | grep -Ei "reconcil|bucket|aws|error"
```

**Known AWS S3 lifecycle constraints to watch for:**

| Rule type | Constraint |
|-----------|-----------|
| `STANDARD_IA` transition | Minimum **30 days** |
| Expiration | Must be **> transition days** |

---

## Phase 6 — Merge, Bump, Tag, Push

### 6a. Merge Dependabot PRs on GitHub

Preferred: merge the integration PR → close the 4 Dependabot PRs as "superseded".

If policy requires individual merges, use this order (least → most risk):

```bash
gh pr merge 43 --squash --delete-branch
gh pr merge 42 --squash --delete-branch
gh pr merge 44 --squash --delete-branch
gh pr merge 40 --squash --delete-branch   # high-risk last
```

### 6b. Bump Helm chart version on `main`

```bash
git checkout main && git pull --ff-only origin main

# In-place version bump
sed -i "s/^version: .*/version: ${NEXT_VERSION}/"         code/charts/kube-s3-operator/Chart.yaml
sed -i "s/^appVersion: .*/appVersion: \"${NEXT_VERSION}\"/" code/charts/kube-s3-operator/Chart.yaml
sed -i "s/^\([[:space:]]*tag:\).*/\1 \"${NEXT_VERSION}\"/" code/charts/kube-s3-operator/values.yaml

git add code/charts/kube-s3-operator/Chart.yaml \
        code/charts/kube-s3-operator/values.yaml
git commit -m "chore(release): bump chart/app/image tag to ${NEXT_VERSION}"
git push origin main
```

### 6c. Tag and push

```bash
git tag -a "v${NEXT_VERSION}" -m "Release v${NEXT_VERSION}: Dependabot PRs batch + Helm bump"
git push origin "v${NEXT_VERSION}"
```

---

## Release Checklist

```
[ ] Phase 1 — integration branch created from main
[ ] Phase 2 — high-risk PR tested in isolation (pass/exclude decision made)
[ ] Phase 3 — all PRs merged into integration branch; go mod tidy clean
[ ] Phase 4 — operator image built and pushed
[ ] Phase 5a — test namespace and aws-secret ready
[ ] Phase 5b — Helm deploy succeeded (no ownership errors)
[ ] Phase 5c — controller pod Running, logs clean
[ ] Phase 5d — S3Bucket CR reached Ready state
[ ] Phase 6a — Dependabot PRs merged/closed on GitHub
[ ] Phase 6b — Chart.yaml / values.yaml version bumped
[ ] Phase 6c — git tag pushed, GitHub Release created
```

---

## Rollback

| Scenario | Rollback |
|----------|---------|
| Smoke test fails after gRPC merge | `git revert` the grpc merge commit; push; rebuild image without it |
| Helm deploy broken | `helm rollback ${HELM_RELEASE} -n ${TEST_NS}` |
| `main` already bumped but smoke failed | Revert version bump commit; `git push origin main --force-with-lease` |
| Image pushed but broken | Re-push previous tag: `docker tag ${IMG_REPO}:prev ${IMG_REPO}:${IMG_TAG} && docker push` |

