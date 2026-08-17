import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/app/globals.css"), "utf8");

test("adapts organization controls for tablet widths", () => {
  expect(css).toContain(".invite-form { grid-template-columns: minmax(0, 1fr) minmax(180px, 0.8fr); }");
  expect(css).toContain(".invite-form .primary { grid-column: 1 / -1; justify-self: start; }");
});

test("lets repeated workspaces use available width without forcing empty columns", () => {
  expect(css).toMatch(/\.organization-grid, \.project-grid \{[^}]*grid-template-columns: repeat\(auto-fit, minmax\(min\(100%, 420px\), 1fr\)\);/);
});

test("gives a single workspace card a focused reading width", () => {
  expect(css).toContain(".organization-grid:has(> :only-child), .project-grid:has(> :only-child) { grid-template-columns: minmax(0, 760px); }");
});

test("gives mobile organization and project actions full-width touch targets", () => {
  expect(css).toContain(".organization-actions, .project-actions { width: 100%; justify-content: stretch; }");
  expect(css).toContain(".organization-actions button, .project-actions button { flex: 1; }");
  expect(css).toContain(".organization-card { grid-column: 1 / -1; }");
});

test("gives the role-first project context and outcome summary readable structure", () => {
  expect(css).toContain(".project-context-panel {");
  expect(css).toContain(".project-context-overview { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr));");
  expect(css).toContain(".project-metadata { display: grid;");
  expect(css).toContain(".project-progress { display: flex;");
  expect(css).toContain(".coverage-value { display: grid;");
  expect(css).toContain(".evidence-count {");
});

test("stacks project context and evidence actions at narrow widths", () => {
  expect(css).toContain(".project-context-overview { grid-template-columns: 1fr; gap: 12px; }");
  expect(css).toContain(".project-metadata { grid-template-columns: 1fr; gap: 10px; }");
  expect(css).toContain(".project-progress { align-items: flex-start; flex-direction: column; gap: 4px; }");
  expect(css).toContain(".evidence-list li .button-link { max-width: 100%; }");
  expect(css).toContain(".response-buttons button { flex: 1; }");
  expect(css).toContain("@media (prefers-reduced-motion: reduce)");
});
test("keeps mobile assessment content readable and stacked", () => {
  expect(css).toContain(".project-card { min-height: auto; gap: 18px; padding: 16px; }");
  expect(css).toContain(".project-card-top { align-items: flex-start; flex-direction: column; gap: 4px; }");
  expect(css).toMatch(/\.assessment-summary \{[^}]*grid-template-columns: 88px minmax\(0, 1fr\) 30px;/);
  expect(css).toMatch(/\.coverage-route \{[^}]*grid-column: 1 \/ -1;[^}]*flex-wrap: wrap;/);
});
