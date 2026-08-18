import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/app/globals.css"), "utf8");
const layout = readFileSync(resolve(process.cwd(), "src/app/layout.tsx"), "utf8");

test("carries the Versotis brand palette and type system into the workspace", () => {
  expect(css).toContain("--magenta: #e6007a;");
  expect(css).toContain("--purple: #7b2cbf;");
  expect(css).toContain("--navy: #0b1d3a;");
  expect(css).toContain("--accent-grad: linear-gradient(135deg, var(--magenta) 0%, var(--purple) 100%);");
  expect(css).toContain('--font-sans: "Space Grotesk", "IBM Plex Sans Thai Looped", -apple-system, BlinkMacSystemFont, sans-serif;');
  expect(css).toContain('--font-th: "IBM Plex Sans Thai Looped", "Space Grotesk", sans-serif;');
  expect(css).toContain('--font-mono: "JetBrains Mono", ui-monospace, monospace;');
  expect(css).toContain("font-family: var(--font-sans);");
  expect(css).toContain("background: var(--accent-grad);");
});

test("loads the brand fonts and records the new visual direction", () => {
  expect(layout).toContain('<link rel="preconnect" href="https://fonts.googleapis.com" />');
  expect(layout).toContain("family=JetBrains+Mono");
  expect(layout).toContain("family=IBM+Plex+Sans+Thai+Looped");
  expect(layout).toContain("family=Space+Grotesk");
  expect(layout).toContain("Versotis Trust Workspace");
  expect(layout).toContain("light + dark brand system");
  expect(layout).not.toContain("Clean Editorial Casefile");
});

test("provides a Versotis dark theme for the reading workspace", () => {
  expect(css).toContain("color-scheme: light dark;");
  expect(css).toContain("@media (prefers-color-scheme: dark)");
  expect(css).toContain("--canvas: #060e1d;");
  expect(css).toContain("--surface: #0b1d3a;");
  expect(css).toContain("--surface-subtle: #10294b;");
  expect(css).toContain("--ink: #f5f7fb;");
  expect(css).toContain("--muted: #a7b3c8;");
  expect(css).toContain("--accent: #ff5aad;");
  expect(css).toContain("--accent-dark: #d1a5ff;");
  expect(css).toContain("--line: rgba(255, 255, 255, 0.12);");
});
