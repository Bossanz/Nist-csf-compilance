"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { AcceptInvitationForm } from "../../../components/AcceptInvitationForm";
import { api } from "../../../lib/api";

export default function InvitationPage() {
  const params = useParams<{ token: string }>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [accepted, setAccepted] = useState(false);

  async function accept(input: { name: string; password: string }) {
    setLoading(true);
    setError("");
    try {
      await api.acceptInvitation(params.token, input);
      await api.logout();
      setAccepted(true);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not activate account");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="auth-layout invitation-layout">
      <section className="auth-intro">
        <div className="context-line"><span>NIST CSF 2.0</span><span aria-hidden="true">/</span><span>Invitation</span></div>
        <h2>Join the assessment workspace.</h2>
        <p>Your access and permissions were selected by the organization administrator.</p>
      </section>
      {accepted ? (
        <section className="auth-panel success-panel">
          <div className="context-line"><span>Invitation</span><span aria-hidden="true">/</span><span>Account ready</span></div>
          <h1>Access activated</h1>
          <p className="muted">Your account is ready. Sign in with the password you just created.</p>
          <Link className="anchor-primary" href="/login">Continue to sign in</Link>
        </section>
      ) : <AcceptInvitationForm loading={loading} error={error} onAccept={accept} />}
    </main>
  );
}
