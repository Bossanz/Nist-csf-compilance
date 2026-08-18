"use client";

import { useState } from "react";
import { APIError, api } from "../../lib/api";
import { ForgotPasswordForm } from "../../components/ForgotPasswordForm";
import { ThemeToggle } from "../../components/ThemeToggle";

export default function ForgotPasswordPage() {
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState("");

  async function requestReset(email: string) {
    setLoading(true);
    setError("");
    try {
      await api.requestPasswordReset(email);
      setSubmitted(true);
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="auth-layout">
      <section className="auth-intro">
        <div className="auth-theme-control"><ThemeToggle /></div>
        <h1>Account recovery.</h1>
        <p>We’ll help you get back to your compliance workspace securely.</p>
      </section>
      <ForgotPasswordForm loading={loading} error={error} submitted={submitted} onSubmit={requestReset} />
    </main>
  );
}

function messageOf(cause: unknown) {
  return cause instanceof APIError ? cause.message : cause instanceof Error ? cause.message : "Something went wrong";
}
