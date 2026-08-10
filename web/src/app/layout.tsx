import "./globals.css";
import type { Metadata } from "next";
export const metadata: Metadata = { title: "CSF Compliance", description: "NIST CSF 2.0 compliance workspace" };

const designContract = "Clean Editorial Casefile | direction 4/7 | seed 313a6a7a";

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body data-design-contract={designContract}>
        {/*
        THESIS: Readability is the product surface: help counselors scan, interpret, and decide.
        OWN-WORLD: Clean Editorial Casefile — white paper, soft neutral structure, restrained teal marks.
        STORY: Counselor reads and reviews; Stakeholder fills and supports with evidence.
        FIRST VIEWPORT: Project context, Function index, current/target status, and one clear next action.
        FORM: Clean Editorial Casefile; grounded direction 4/7; source key 313a6a7a. FINISH: document the built world.
        */}
        {children}
      </body>
    </html>
  );
}
