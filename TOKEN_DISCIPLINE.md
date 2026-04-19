# Token Discipline For Market Lens

This file documents a required operating rule for future work on this project.

The previous implementation turns used too many tokens because the work included broad file reads, large patch payloads, noisy command outputs, verbose progress narration, and oversized smoke-test responses. That must not be repeated.

## Required Rule

For every future task, reference this file before working and use a token-light workflow.

The goal is:

- Same engineering care.
- Less transcript.
- Less irrelevant tool output.
- Smaller patches.
- Compact summaries.

## What Went Wrong

- Read too much file content instead of targeted ranges.
- Replaced large files in full when smaller patches would have worked.
- Let Docker/build/API outputs dump too much text.
- Printed full JSON smoke-test responses instead of key fields.
- Gave too many progress updates for routine steps.
- Combined too many changes in one turn without compact phase boundaries.

## Required Workflow

Before coding:

- State a compact plan.
- Update README/change log only when required by project workflow.
- Use targeted inspection before reading large files.

During coding:

- Prefer `rg`, `rg --files`, `wc -l`, and narrow `sed` ranges.
- Do not dump entire files into context unless necessary.
- Patch focused sections instead of rewriting large files.
- Split large changes into small phases if they grow.

During verification:

- Keep command output compact.
- For API smoke tests, print only key fields.
- Inspect Docker logs only on failure or when explicitly needed.
- Avoid full JSON responses in the transcript.

During communication:

- Use fewer intermediate updates.
- Report only meaningful state changes, failures, or decisions.
- Keep final summaries concise.

During commits:

- Commit only after explicit user signal.
- Inspect status and diffs first.
- Do not blindly commit everything.
- Group related changes into sensible commits.
- Push only after grouped commits are created and verified.

## Compact Command Patterns

Use this style for API smoke tests:

```bash
curl -s http://127.0.0.1:3000/api/health
```

For larger responses, summarize fields:

```bash
curl -s http://127.0.0.1:3000/api/runs/$RUN_ID | node -e "
let s='';
process.stdin.on('data', d => s += d);
process.stdin.on('end', () => {
  const r = JSON.parse(s);
  console.log({
    status: r.status,
    symbol: r.symbol,
    workers: r.workers?.length,
    analysisId: r.analysis?.id
  });
});
"
```

Use this style for file inspection:

```bash
rg "startAnalysisRun|runStatus|WorkflowRail" .
sed -n '40,120p' frontend/src/App.jsx
```

Avoid broad inspection unless necessary:

```bash
sed -n '1,500p' large-file.js
docker compose logs --tail=200
curl -s http://.../large-response
```

## Default Behavior Going Forward

Unless the user asks for exhaustive output, use compact mode:

- Small plan.
- Minimal targeted reads.
- Focused patches.
- Compact verification.
- Short final answer.

If a task starts getting large, pause and split it into phases instead of consuming the context window.
