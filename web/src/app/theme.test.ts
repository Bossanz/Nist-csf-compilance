import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/app/globals.css"), "utf8");
const layout = readFileSync(resolve(process.cwd(), "src/app/layout.tsx"), "utf8");

test("carries the Versotis brand palette and type system into the workspace", () => {
  expect(css).toContain("--qid-pink: #eb147c;");
  expect(css).toContain("--qid-purple: #6a32de;");
  expect(css).toContain("--qid-bg: #0b0914;");
  expect(css).toContain("--qid-surface: #13101f;");
  expect(css).toContain("--qid-action-gradient: linear-gradient(135deg, var(--qid-pink) 0%, var(--qid-purple) 100%);");
  expect(css).toContain('--qid-font-heading: "Space Grotesk", system-ui, sans-serif;');
  expect(css).toContain('--qid-font-body: "Inter", system-ui, sans-serif;');
  expect(css).toContain('--qid-font-mono: "JetBrains Mono", ui-monospace, monospace;');
  expect(css).toContain("font-family: var(--qid-font-body);");
  expect(css).toContain("background: var(--qid-action-gradient);");
  expect(css).toContain("--qid-accent: #eb147c;");
  expect(css).toContain("--qid-secondary: #6a32de;");
  expect(css).toMatch(/\.section-index, \.eyebrow \{\s*color: var\(--qid-accent\);/);
  expect(css).toMatch(/\.outcome-code \{ color: var\(--qid-accent\);/);
});

test("loads the brand fonts and records the new visual direction", () => {
  expect(layout).toContain('<link rel="preconnect" href="https://fonts.googleapis.com" />');
  expect(layout).toContain("family=JetBrains+Mono");
  expect(layout).toContain("family=Inter");
  expect(layout).toContain("family=Space+Grotesk");
  expect(layout).toContain("QID v3 Compliance Workspace");
  expect(layout).toContain("dark-first workspace system");
  expect(layout).toContain('data-theme="dark"');
  expect(layout).not.toContain("Clean Editorial Casefile");
});

test("provides a Versotis dark theme for the reading workspace", () => {
  expect(css).toContain("color-scheme: dark;");
  expect(css).toContain("color-scheme: light;");
  expect(css).toContain(':root[data-theme="dark"]');
  expect(css).toContain("--qid-bg: #0b0914;");
  expect(css).toContain("--qid-surface-2: #1c182d;");
  expect(css).toContain("--qid-surface-3: #26213c;");
  expect(css).toContain("--qid-text: #f8f8fc;");
  expect(css).toContain("--qid-muted: #a19db5;");
  expect(css).toContain("--qid-focus-ring:");
});

test("keeps report print tokens readable when dark theme is active", () => {
  const printBlock = css.slice(css.lastIndexOf("@media print"));
  expect(printBlock).toContain(':root[data-theme="dark"]');
  expect(printBlock).toContain("--qid-surface: #ffffff;");
  expect(printBlock).toContain("--muted-strong: #1e293b;");
});

test("keeps the Assignment group muted when another workspace surface is active", () => {
  expect(css).toMatch(/\.sidebar-group-toggle \{\s*color: var\(--qid-muted\);\s*\}/);
});
