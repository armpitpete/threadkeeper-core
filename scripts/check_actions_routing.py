from pathlib import Path
import re
import sys

path = Path(__file__).resolve().parents[1] / ".github" / "workflows" / "conformance.yml"
lines = path.read_text(encoding="utf-8").splitlines()
OWNED = "[self-hosted, Linux, ARM64, oracle-ci, threadkeeper-core]"
expected = {
    "test": "ubuntu-latest",
    "test-owned": OWNED,
    "windows-git-environment-isolation": "windows-latest",
}

jobs = {}
current = None
in_jobs = False
for line in lines:
    if line == "jobs:":
        in_jobs = True
        continue
    if not in_jobs:
        continue
    if line and not line.startswith(" "):
        break
    m = re.match(r"^  ([A-Za-z0-9_-]+):\s*$", line)
    if m:
        current = m.group(1)
        jobs[current] = []
    elif current is not None:
        jobs[current].append(line)

errors = []
for name, runner in expected.items():
    block = jobs.get(name)
    if block is None:
        errors.append(f"missing job: {name}")
        continue
    runs = [x.strip().split(":",1)[1].strip() for x in block if x.strip().startswith("runs-on:")]
    if runs != [runner]:
        errors.append(f"{name}: expected runs-on {runner}, got {runs}")
    if any("${{" in x for x in runs):
        errors.append(f"{name}: dynamic runs-on forbidden")

if "test" in jobs and not any("if: github.event_name == 'pull_request'" in x for x in jobs["test"]):
    errors.append("test: hosted Linux exception must be pull_request-only")
if "test-owned" in jobs and not any("if: github.event_name == 'push'" in x for x in jobs["test-owned"]):
    errors.append("test-owned: owned Linux job must be push-only")
if set(jobs) != set(expected):
    errors.append("job set differs from routing policy: " + repr(sorted(jobs)))

if errors:
    print("\n".join(errors), file=sys.stderr)
    raise SystemExit(1)
print("PASS: Threadkeeper Core Actions routing policy")
