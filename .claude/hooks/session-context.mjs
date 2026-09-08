#!/usr/bin/env node
// SessionStart: three facts that would otherwise cost tool calls every session.
import { readFileSync, existsSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';

function readInput() {
	try {
		return JSON.parse(readFileSync(0, 'utf8'));
	} catch {
		return {};
	}
}

const input = readInput();
const projectDir = process.env.CLAUDE_PROJECT_DIR || input.cwd || process.cwd();
const git = (args) => {
	const r = spawnSync('git', args, { cwd: projectDir, encoding: 'utf8' });
	return r.status === 0 ? r.stdout.trim() : '';
};

const lines = [];

const branch = git(['rev-parse', '--abbrev-ref', 'HEAD']);
if (branch) lines.push(`Branch: ${branch}`);

const dirty = git(['status', '--porcelain']);
if (dirty) lines.push(`Uncommitted files: ${dirty.split('\n').length}`);

try {
	const changelog = readFileSync(join(projectDir, 'CHANGELOG.md'), 'utf8');
	const ids = [...changelog.matchAll(/\[(\d{3})\]/g)].map((m) => Number(m[1]));
	if (ids.length) {
		const highest = Math.max(...ids);
		lines.push(
			`CHANGELOG feature IDs: highest is ${String(highest).padStart(3, '0')}, so a new feature is ${String(highest + 1).padStart(3, '0')}.`
		);
	}
} catch {}

const plan = join(projectDir, 'tmp', 'plan.md');
if (existsSync(plan)) lines.push('tmp/plan.md exists — unfinished multi-step work from an earlier session; read it before starting something new.');

if (lines.length === 0) process.exit(0);

console.log(
	JSON.stringify({
		hookSpecificOutput: {
			hookEventName: 'SessionStart',
			additionalContext: lines.join('\n')
		}
	})
);
