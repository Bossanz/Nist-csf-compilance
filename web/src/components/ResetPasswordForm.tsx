"use client";

import { useState } from "react";

type Props = {
  token: string;
  loading: boolean;
  error: string;
  onSubmit: (token: string, password: string) => void;
};

export function ResetPasswordForm({ token, loading, error, onSubmit }: Props) {
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [validationError, setValidationError] = useState("");

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (password !== confirmation) {
      setValidationError("Passwords do not match.");
      return;
    }
    setValidationError("");
    onSubmit(token, password);
  }

  return (
    <section className="auth-panel" aria-labelledby="reset-password-title">
      <h1 id="reset-password-title">Set a new password</h1>
      <p className="muted">Choose a password with at least 8 characters. This reset link can only be used once.</p>
      {(error || validationError) && <div className="error" role="alert">{validationError || error}</div>}
      <form className="auth-form" onSubmit={submit}>
        <label className="field"><span>Password</span><input type="password" autoComplete="new-password" required minLength={8} value={password} onChange={(event) => setPassword(event.target.value)} /></label>
        <label className="field"><span>Confirm password</span><input type="password" autoComplete="new-password" required minLength={8} value={confirmation} onChange={(event) => setConfirmation(event.target.value)} /></label>
        <button className="primary" type="submit" disabled={loading}>{loading ? "Resetting…" : "Reset password"}</button>
      </form>
      <a className="auth-link" href="/login">Back to sign in</a>
    </section>
  );
}
