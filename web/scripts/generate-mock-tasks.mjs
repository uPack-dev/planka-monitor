#!/usr/bin/env node
import { mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import {
  MIN_DESIGN_TASK_COUNT,
  generateMockTasks,
} from '../app/utils/mockTasks.mjs';

const options = parseArgs(process.argv.slice(2));
const count = Number(options.count || options.n || MIN_DESIGN_TASK_COUNT);
const tasks = generateMockTasks({
  count,
  from: options.from,
  to: options.to,
  now: options.now,
  selectedDate: options.date || options.selectedDate,
});
const payload = {
  generatedAt: new Date().toISOString(),
  count: tasks.length,
  items: tasks,
};
const json = `${JSON.stringify(payload, null, 2)}\n`;

if (options.out) {
  const outputPath = resolve(process.cwd(), options.out);
  await mkdir(dirname(outputPath), { recursive: true });
  await writeFile(outputPath, json);
  console.log(`Generated ${tasks.length} mock tasks: ${outputPath}`);
} else {
  process.stdout.write(json);
}

function parseArgs(args) {
  const parsed = {};

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (!arg.startsWith('--')) continue;

    const [rawKey, inlineValue] = arg.slice(2).split('=');
    const key = rawKey.trim();
    const value = inlineValue ?? args[index + 1];

    if (inlineValue === undefined && value && !value.startsWith('--')) {
      index += 1;
    }

    parsed[key] = value === undefined || value.startsWith('--') ? true : value;
  }

  return parsed;
}
