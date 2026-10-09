import assert from "node:assert/strict";
import { test } from "node:test";
import { convertFile } from "./gen-js.mjs";

test("strips types, keeps JSX, comments and template tags, rewrites relative .ts imports", async () => {
  const src = `import type { ReactNode } from "react";
import { useState } from "react";
import { api } from "./api.ts";

// Counter shows a button.
export function Counter({ label }: { label: string; icon?: ReactNode }) {
  const [n, setN] = useState<number>(0);
  return <button onClick={() => setN(n + 1)}>{label} {n} {"{{.ProjectSlug}}"} {api}</button>;
}
`;
  const [name, out] = await convertFile("counter.tsx.tmpl", src);
  assert.equal(name, "counter.jsx.tmpl");
  assert.doesNotMatch(out, /ReactNode|<number>|: \{ label/);
  assert.match(out, /from "\.\/api\.js"/);
  assert.match(out, /\/\/ Counter shows a button\./);
  assert.match(out, /\{"\{\{\.ProjectSlug\}\}"\}/);
  assert.match(out, /<button onClick=/);
});

test("value tags in code positions survive", async () => {
  const [, out] = await convertFile("a.ts.tmpl", "export const port: number = {{.Port}};\n");
  assert.equal(out, "export const port = {{.Port}};\n");
});

test("control tags in code are rejected", async () => {
  await assert.rejects(convertFile("a.ts.tmpl", "{{if .IsTypeScript}}\nconst a = 1;\n{{end}}\n"), /only \{\{\.Var\}\} tags/);
});

test("mts/cts map to mjs/cjs", async () => {
  assert.equal((await convertFile("x.config.mts.tmpl", "export default {};\n"))[0], "x.config.mjs.tmpl");
});

test("vue script lang=ts is converted, template left verbatim", async () => {
  const src = `<script setup lang="ts">
const count: number = 1;
</script>

<template>
  <p>{{"{{"}} count {{"}}"}}</p>
</template>
`;
  const [name, out] = await convertFile("App.vue.tmpl", src);
  assert.equal(name, "App.vue.tmpl");
  assert.match(out, /<script setup>\nconst count = 1;\n<\/script>/);
  assert.match(out, /<p>\{\{"\{\{"\}\} count \{\{"\}\}"\}\}<\/p>/);
});

test("package.json drops TypeScript-only deps and typecheck", async () => {
  const src = JSON.stringify({
    name: "{{.ProjectSlug}}",
    scripts: { dev: "node --watch src/main.ts", typecheck: "tsc --noEmit", test: "vitest run" },
    dependencies: { hono: "4.0.0" },
    devDependencies: { typescript: "6.0.0", "@types/node": "24.0.0", vitest: "4.0.0", "typescript-eslint": "8.0.0" },
  });
  const [, out] = await convertFile("package.json.tmpl", src);
  const pkg = JSON.parse(out);
  assert.equal(pkg.name, "{{.ProjectSlug}}");
  assert.deepEqual(pkg.scripts, { dev: "node --watch src/main.js", test: "vitest run" });
  assert.deepEqual(pkg.devDependencies, { vitest: "4.0.0" });
});

test("declaration files and tsconfig are dropped, other files copied", async () => {
  assert.equal(await convertFile("env.d.ts.tmpl", "declare const x: string;"), null);
  assert.equal(await convertFile("tsconfig.node.json.tmpl", "{}"), null);
  assert.deepEqual(await convertFile("Makefile.tmpl", "a\n"), ["Makefile.tmpl", "a\n"]);
});

test("erased type imports leave no gap in the import block", async () => {
  const src = 'import { a } from "./a.ts";\nimport type { B } from "./b.ts";\nimport { c } from "./c.ts";\n\nexport const x: B = a + c;\n';
  const [, out] = await convertFile("x.ts.tmpl", src);
  assert.equal(out, 'import { a } from "./a.js";\nimport { c } from "./c.js";\n\nexport const x = a + c;\n');
});
