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
        <span className="section-index">NIST CSF 2.0 / INVITATION</span>
        <h2>Join the assessment workspace.</h2>
        <p>Your access and permissions were selected by the organization administrator.</p>
      </section>
      {accepted ? (
        <section className="auth-panel success-panel">
          <span className="section-index">ACCOUNT READY</span>
          <h1>Access activated</h1>
          <p className="muted">Your account is ready. Sign in with the password you just created.</p>
          <Link className="primary button-link" href="/">Continue to sign in</Link>
        </section>
      ) : <AcceptInvitationForm loading={loading} error={error} onAccept={accept} />}
    </main>
  );
}
