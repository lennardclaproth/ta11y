#!/usr/bin/env node
// PreToolUse(Bash|PowerShell): a bare `go` command in a linked git worktree resolves the repo
// go.work to the main checkout and fails. Prefix GOWORK=off when we can do so safely, and
// otherwise just say so rather than rewriting a command we don't fully understand.
import { readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';

const GO_CMD = /^\s*go\s+(build|vet|test|list|run)\b/;

function readInput() {
	try {
		return JSON.parse(readFileSync(0, 'utf8'));
	} catch {
		return null;
	}
}

function inLinkedWorktree(dir) {
	// A linked worktree has .git as a file containing "gitdir: ..."; the main checkout has a directory.
	try {
		return statSync(join(dir, '.git')).isFile();
	} catch {
		return false;
	}
}

const input = readInput();
if (!input) process.exit(0);

const projectDir = process.env.CLAUDE_PROJECT_DIR || input.cwd || process.cwd();
const command = input.tool_input?.command;

if (typeof command !== 'string' || !GO_CMD.test(command)) process.exit(0);
if (/GOWORK/.test(command)) process.exit(0);
if (!inLinkedWorktree(projectDir)) process.exit(0);

const prefix = input.tool_name === 'PowerShell' ? "$env:GOWORK='off'; " : 'GOWORK=off ';

console.log(
	JSON.stringify({
		hookSpecificOutput: {
			hookEventName: 'PreToolUse',
			updatedInput: { ...input.tool_input, command: prefix + command }
		},
		systemMessage: 'Linked worktree: prefixed GOWORK=off so go.work does not resolve to the main checkout.'
	})
);
