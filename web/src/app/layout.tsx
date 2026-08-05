import "./globals.css";
import type { Metadata } from "next";
export const metadata: Metadata = { title: "CSF Compliance", description: "NIST CSF 2.0 compliance workspace" };
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) { return <html lang="en"><body>{children}</body></html>; }
