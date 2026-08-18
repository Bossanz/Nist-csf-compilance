"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { APIError, api } from "../../lib/api";
import { ResetPasswordForm } from "../../components/ResetPasswordForm";
import { ThemeToggle } from "../../components/ThemeToggle";

export default function ResetPasswordPage() {
  return <Suspense fallback={<main className="screen-center">Loading reset link…</main>}><ResetPasswordContent /></Suspense>;
}

function ResetPasswordContent() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") || "";
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function resetPassword(resetToken: string, password: string) {
    setLoading(true);
    setError("");
    try {
      await api.confirmPasswordReset(resetToken, password);
      router.replace("/login");
    } catch (cause) {
      setError(messageOf(cause));
      setLoading(false);
    }
  }

  return (
    <main className="auth-layout">
      <section className="auth-intro">
        <div className="auth-theme-control"><ThemeToggle /></div>
        <h1>Secure reset.</h1>
        <p>This one-time link lets you choose a new password for your compliance workspace.</p>
      </section>
      {token ? <ResetPasswordForm token={token} loading={loading} error={error} onSubmit={resetPassword} /> : (
        <section className="auth-panel" aria-labelledby="invalid-reset-title">
          <h1 id="invalid-reset-title">Reset link unavailable</h1>
          <p className="muted">This link is missing its token. Request a new one to continue.</p>
          <a className="anchor-primary" href="/forgot-password">Request a new link</a>
        </section>
      )}
    </main>
  );
}

function messageOf(cause: unknown) {
  return cause instanceof APIError ? cause.message : cause instanceof Error ? cause.message : "Something went wrong";
}
