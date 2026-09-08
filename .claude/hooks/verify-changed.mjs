#!/usr/bin/env node
// Stop: if this session edited Go files, prove the backend still builds and vets before the turn
// ends. Exit 2 keeps the conversation going so the failure gets fixed instead of reported as done.
// Gives up after two blocked attempts so a genuinely stuck build can't trap the session.
import { readFileSync, existsSync, rmSync, writeFileSync, statSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';

const MAX_ATTEMPTS = 2;

function readInput() {
	try {
		return JSON.parse(readFileSync(0, 'utf8'));
	} catch {
		return {};
	}
}

const input = readInput();
const projectDir = process.env.CLAUDE_PROJECT_DIR || input.cwd || process.cwd();
const session = input.session_id || 'default';
const cacheDir = join(projectDir, '.claude', '.cache');
const touchedFile = join(cacheDir, `touched-${session}.txt`);
const attemptsFile = join(cacheDir, `attempts-${session}.txt`);

if (!existsSync(touchedFile)) process.exit(0);

const touched = readFileSync(touchedFile, 'utf8')
	.split('\n')
	.map((l) => l.trim())
	.filter((l) => l.endsWith(".go") && normalize(l).includes("apps/api/"));

const clear = () => {
	rmSync(touchedFile, { force: true });
	rmSync(attemptsFile, { force: true });
};

if (touched.length === 0) {
	clear();
	process.exit(0);
}

const attempts = existsSync(attemptsFile) ? Number(readFileSync(attemptsFile, 'utf8').trim()) || 0 : 0;
if (attempts >= MAX_ATTEMPTS) {
	clear();
	console.error(
		'go build/vet still failing after ' +
			MAX_ATTEMPTS +
			' attempts. Not blocking again — report the remaining failure to the user explicitly.'
	);
	process.exit(1); // non-blocking notice
}

const apiDir = join(projectDir, 'apps', 'api');
let inLinkedWorktree = false;
try {
	inLinkedWorktree = statSync(join(projectDir, '.git')).isFile();
} catch {}

const env = { ...process.env, ...(inLinkedWorktree ? { GOWORK: 'off' } : {}) };
const run = (args) => spawnSync('go', args, { cwd: apiDir, encoding: 'utf8', env });

for (const args of [['build', './...'], ['vet', './...']]) {
	const result = run(args);
	if (result.error) process.exit(0); // no Go toolchain here; nothing to enforce
	if (result.status !== 0) {
		writeFileSync(attemptsFile, String(attempts + 1));
		const output = (result.stderr || result.stdout || '').trim().split('\n').slice(0, 40).join('\n');
		console.error(
			`\`go ${args.join(' ')}\` failed in apps/api after this session's Go edits. Fix it before finishing:\n\n${output}`
		);
		process.exit(2); // blocks the stop; stderr goes to Claude
	}
}

clear();
console.log(JSON.stringify({ systemMessage: 'go build ./... and go vet ./... pass in apps/api.' }));

// Windows paths arrive with OS separators; compare on forward slashes.
function normalize(p) {
	return p.split(String.fromCharCode(92)).join("/");
}
