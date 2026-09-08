#!/usr/bin/env node
// PostToolUse(Edit|Write): gofmt the one file that was just written, and record it so the Stop
// hook knows whether this session touched Go code at all.
import { readFileSync, mkdirSync, appendFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { join, basename } from 'node:path';

// Windows paths arrive with OS separators; compare on forward slashes.
function normalize(p) {
	return p.split(String.fromCharCode(92)).join('/');
}

function readInput() {
	try {
		return JSON.parse(readFileSync(0, 'utf8'));
	} catch {
		return null;
	}
}

const input = readInput();
if (!input) process.exit(0);

const file = input.tool_input?.file_path;
if (typeof file !== 'string' || !file.endsWith('.go')) process.exit(0);

const projectDir = process.env.CLAUDE_PROJECT_DIR || input.cwd || process.cwd();
const cacheDir = join(projectDir, '.claude', '.cache');
const session = input.session_id || 'default';

try {
	mkdirSync(cacheDir, { recursive: true });
	appendFileSync(join(cacheDir, `touched-${session}.txt`), file + '\n');
} catch {
	// recording is best-effort; never fail the turn over it
}

const listed = spawnSync('gofmt', ['-l', file], { encoding: 'utf8' });
if (listed.error || listed.status !== 0) process.exit(0); // gofmt missing, or the file does not parse
if (!listed.stdout.trim()) process.exit(0); // already formatted

const written = spawnSync('gofmt', ['-w', file], { encoding: 'utf8' });
if (written.error || written.status !== 0) process.exit(0);

const root = normalize(projectDir);
const full = normalize(file);
const shown = full.startsWith(root) ? full.slice(root.length + 1) : basename(file);

console.log(JSON.stringify({ systemMessage: `gofmt reformatted ${shown} after your edit.` }));
