import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const frontend = resolve(here, '..');
const repo = resolve(here, '../..');
const src = resolve(repo, 'data/full.tsv');
const outNames = resolve(frontend, 'src/lib/domains-names.txt');
const outAvail = resolve(frontend, 'src/lib/domains-avail.txt');

const rows = [];
for (const line of readFileSync(src, 'utf8').split('\n')) {
	if (!line) continue;
	const i = line.indexOf('\t');
	rows.push([line.slice(0, i), line.slice(i + 1)]);
}
rows.sort((a, b) => (a[0] < b[0] ? -1 : a[0] > b[0] ? 1 : 0));

const names = new Array(rows.length);
const avail = new Array(rows.length);
for (let i = 0; i < rows.length; i++) {
	names[i] = rows[i][0];
	avail[i] = rows[i][1] === 'available' ? '1' : '0';
}

writeFileSync(outNames, names.join('\n'));
writeFileSync(outAvail, avail.join(''));

const availCount = avail.reduce((n, c) => n + (c === '1' ? 1 : 0), 0);
console.log(`wrote ${rows.length} names (${availCount} available)`);
