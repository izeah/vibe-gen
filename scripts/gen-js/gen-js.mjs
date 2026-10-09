// gen-js derives templates/javascript/<fw> from templates/typescript/<fw> (TypeScript is the single source).
//
//   node gen-js.mjs           rewrite the JavaScript templates
//   node gen-js.mjs --check   fail if the committed JavaScript templates are stale
//
// Rules (authoring conventions they rely on: vibe-gen/README.md, "Writing TypeScript templates"):
//   *.ts/.mts/.cts/.tsx.tmpl   types stripped (sucrase), relative .ts imports → .js, prettier
//   *.vue/.svelte.tmpl         same for <script lang="ts"> blocks; lang="ts" removed
//   package.json.tmpl          TypeScript-only deps and the typecheck script removed, .ts paths in scripts → .js
//   *.d.ts.tmpl, tsconfig*.json.tmpl   dropped
//   anything else              copied verbatim (branch on {{if .IsTypeScript}} inside it when needed)
import { mkdir, mkdtemp, readdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";
import { format } from "prettier";
import { transform } from "sucrase";

const TEMPLATES = join(dirname(fileURLToPath(import.meta.url)), "../../templates");

// Frameworks with no JavaScript variant (decorators do not survive type stripping).
const TS_ONLY = new Set(["nestjs"]);

const TS_ONLY_DEPS = /^(typescript|typescript-eslint|@types\/.+|vue-tsc|svelte-check|@tsconfig\/.+)$/;

// Go template tags must survive the JS tooling: swap them for identifiers and back.
// Only value tags ({{.Var}}) are allowed in code; control tags would turn into stray statements.
function protect(code, file) {
  const tags = [];
  const out = code.replace(/\{\{[\s\S]*?\}\}/g, (tag) => {
    if (!/^\{\{\s*\.\w+\s*\}\}$/.test(tag)) {
      throw new Error(`${file}: only {{.Var}} tags are allowed in TypeScript code, found ${tag}`);
    }
    tags.push(tag);
    return `__VIBE_TMPL_${tags.length - 1}__`;
  });
  return [out, (s) => s.replace(/__VIBE_TMPL_(\d+)__/g, (_, i) => tags[Number(i)])];
}

async function stripTypes(code, file, jsx) {
  const [safe, restore] = protect(code, file);
  const js = transform(safe, {
    transforms: jsx ? ["typescript", "jsx"] : ["typescript"],
    jsxRuntime: "preserve",
    keepUnusedImports: true, // like verbatimModuleSyntax: only `import type` is erased
    disableESTransforms: true,
    filePath: file,
  }).code;
  // An erased `import type` leaves a blank line inside the import block: close the gap.
  const joined = js.replace(/(from\s+["'][^"'\n]+["'];?\n)\n+(?=import\s)/g, "$1");
  const imports = joined.replace(/(["'])(\.{1,2}\/[^"'\n]*?)\.([mc]?)ts(x?)\1/g, (_, q, p, mc, x) => `${q}${p}.${mc}js${x}${q}`);
  return restore(await format(imports, { parser: "babel" }));
}

async function convertSFC(code, file) {
  const blocks = [];
  const re = /<script([^>]*?)\s+lang=["']ts["']([^>]*)>([\s\S]*?)<\/script>/g;
  for (const m of code.matchAll(re)) {
    blocks.push({ m, js: await stripTypes(m[3], file, false) });
  }
  let out = "";
  let last = 0;
  for (const { m, js } of blocks) {
    out += code.slice(last, m.index) + `<script${m[1]}${m[2]}>\n${js}</script>`;
    last = m.index + m[0].length;
  }
  return out + code.slice(last);
}

function convertPackageJSON(code, file) {
  const [safe, restore] = protect(code, file);
  const pkg = JSON.parse(safe);
  for (const field of ["dependencies", "devDependencies", "peerDependencies"]) {
    for (const name of Object.keys(pkg[field] ?? {})) {
      if (TS_ONLY_DEPS.test(name)) delete pkg[field][name];
    }
  }
  if (pkg.scripts) {
    delete pkg.scripts.typecheck;
    for (const [k, v] of Object.entries(pkg.scripts)) {
      pkg.scripts[k] = v.replace(/\.([mc]?)ts(x?)\b/g, ".$1js$2");
    }
  }
  return restore(JSON.stringify(pkg, null, 2) + "\n");
}

// convertFile returns [outputName, content], or null to drop the file.
export async function convertFile(name, code, file = name) {
  if (/\.d\.[mc]?ts\.tmpl$/.test(name) || /^tsconfig.*\.json\.tmpl$/.test(name)) return null;
  let m = name.match(/^(.*)\.([mc]?)ts(x?)\.tmpl$/);
  if (m) return [`${m[1]}.${m[2]}js${m[3]}.tmpl`, await stripTypes(code, file, m[3] === "x")];
  if (/\.(vue|svelte)\.tmpl$/.test(name)) return [name, await convertSFC(code, file)];
  if (name === "package.json.tmpl") return [name, convertPackageJSON(code, file)];
  return [name, code];
}

async function* walk(dir) {
  for (const e of await readdir(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) yield* walk(p);
    else yield p;
  }
}

export async function convertTree(src, out) {
  await rm(out, { recursive: true, force: true });
  for await (const p of walk(src)) {
    const rel = relative(src, p);
    const res = await convertFile(rel.split("/").pop(), await readFile(p, "utf8"), rel);
    if (!res) continue;
    const dest = join(out, dirname(rel), res[0]);
    await mkdir(dirname(dest), { recursive: true });
    await writeFile(dest, res[1]);
  }
}

async function readTree(dir) {
  const files = new Map();
  try {
    for await (const p of walk(dir)) files.set(relative(dir, p), await readFile(p, "utf8"));
  } catch (e) {
    if (e.code !== "ENOENT") throw e;
  }
  return files;
}

async function main() {
  const check = process.argv.includes("--check");
  const tsRoot = join(TEMPLATES, "typescript");
  const frameworks = (await readdir(tsRoot, { withFileTypes: true }).catch(() => []))
    .filter((e) => e.isDirectory() && !TS_ONLY.has(e.name))
    .map((e) => e.name);

  const tmp = check ? await mkdtemp(join(tmpdir(), "gen-js-")) : null;
  const stale = [];
  for (const fw of frameworks) {
    const committed = join(TEMPLATES, "javascript", fw);
    const out = check ? join(tmp, fw) : committed;
    await convertTree(join(tsRoot, fw), out);
    if (!check) {
      console.log(`gen-js: templates/javascript/${fw}`);
      continue;
    }
    const [want, got] = [await readTree(out), await readTree(committed)];
    for (const f of new Set([...want.keys(), ...got.keys()])) {
      if (want.get(f) !== got.get(f)) stale.push(`templates/javascript/${fw}/${f}`);
    }
  }
  if (tmp) await rm(tmp, { recursive: true, force: true });
  if (stale.length) {
    console.error(`gen-js-check: JavaScript templates are stale, run \`make gen-js\`:\n  ${stale.join("\n  ")}`);
    process.exit(1);
  }
  if (check) console.log(`gen-js-check: ${frameworks.length} framework(s) in sync`);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
