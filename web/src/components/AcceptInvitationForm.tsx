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
        <span className="section-index">INVITED ACCESS</span>
        <h1 id="activate-title">Activate account</h1>
        <p className="muted">Set your display name and password to join the workspace.</p>
      </div>
      {error && <div className="error" role="alert">{error}</div>}
      <form className="auth-form" onSubmit={(event) => {
        event.preventDefault();
        onAccept({ name: name.trim(), password });
      }}>
        <label className="field"><span>Name</span><input required value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label className="field"><span>Password</span><input type="password" minLength={12} required value={password} onChange={(event) => setPassword(event.target.value)} /></label>
        <button className="primary" disabled={loading}>{loading ? "Activating…" : "Activate account"}</button>
      </form>
    </section>
  );
}
