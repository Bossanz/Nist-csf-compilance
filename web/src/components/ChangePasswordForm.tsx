"use client";

import { useState } from "react";

type Props = {
  loading: boolean;
  error: string;
  onSubmit: (currentPassword: string, newPassword: string) => void;
};

export function ChangePasswordForm({ loading, error, onSubmit }: Props) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [validationError, setValidationError] = useState("");

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (newPassword !== confirmation) {
      setValidationError("Passwords do not match.");
      return;
    }
    setValidationError("");
    onSubmit(currentPassword, newPassword);
  }

  return (
    <section className="auth-panel" aria-labelledby="change-password-title">
      <h1 id="change-password-title">Change password</h1>
      <p className="muted">Changing your password signs out all active sessions for security.</p>
      {(error || validationError) && <div className="error" role="alert">{validationError || error}</div>}
      <form className="auth-form" onSubmit={submit}>
        <label className="field"><span>Current password</span><input type="password" autoComplete="current-password" required value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} /></label>
        <label className="field"><span>New password</span><input type="password" autoComplete="new-password" required minLength={8} value={newPassword} onChange={(event) => setNewPassword(event.target.value)} /></label>
        <label className="field"><span>Confirm new password</span><input type="password" autoComplete="new-password" required minLength={8} value={confirmation} onChange={(event) => setConfirmation(event.target.value)} /></label>
        <button className="primary" type="submit" disabled={loading}>{loading ? "Changing…" : "Change password"}</button>
      </form>
      <a className="auth-link" href="/organizations">Back to workspace</a>
    </section>
  );
}
