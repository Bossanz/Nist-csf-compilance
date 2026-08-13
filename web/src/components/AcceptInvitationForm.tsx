"use client";

import { useState } from "react";

export function AcceptInvitationForm({
  loading,
  error,
  onAccept,
}: {
  loading: boolean;
  error: string;
  onAccept: (input: { name: string; password: string }) => void;
}) {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");

  return (
    <section className="auth-panel" aria-labelledby="activate-title">
      <div>
        <h1 id="activate-title">Activate account</h1>
        <p className="muted">Set your display name and password to join the workspace.</p>
      </div>
      {error && <div className="error" role="alert">{error}</div>}
      <form className="auth-form" onSubmit={(event) => {
        event.preventDefault();
        onAccept({ name: name.trim(), password });
      }}>
        <label className="field"><span>Name</span><input required value={name} onChange={(event) => setName(event.target.value)} /></label>
        <div className="field">
          <label className="field-label" htmlFor="invite-password">Password</label>
          <small id="invite-password-help" className="field-help">Use at least 12 characters.</small>
          <input id="invite-password" type="password" minLength={12} required aria-describedby="invite-password-help" value={password} onChange={(event) => setPassword(event.target.value)} />
        </div>
        <button className="primary" disabled={loading}>{loading ? "Activating…" : "Activate account"}</button>
      </form>
    </section>
  );
}
