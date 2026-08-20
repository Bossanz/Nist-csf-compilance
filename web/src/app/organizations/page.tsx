"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { APIError, api } from "../../lib/api";
import type { Invitation, Organization, User } from "../../lib/types";
import { OrganizationDashboard } from "../../components/OrganizationDashboard";
import { organizationPath } from "../../lib/routes";

export default function OrganizationsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [counselors, setCounselors] = useState<User[]>([]);
  const [counselorInvitationURL, setCounselorInvitationURL] = useState("");
  const [authChecked, setAuthChecked] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, [retryCount]);

  async function initialize(active: boolean) {
    try {
      const currentUser = await api.me();
      const [organizationRows, counselorRows] = await Promise.all([
        api.getOrganizations(),
        currentUser.role === "counselor_admin" ? api.getCounselors() : Promise.resolve<User[]>([]),
      ]);
      if (!active) return;
      setUser(currentUser);
      setOrganizations(organizationRows);
      setCounselors(counselorRows);
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 401) {
        router.replace("/login");
      } else if (active) {
        setError(messageOf(cause));
      }
    } finally {
      if (active) setAuthChecked(true);
    }
  }

  async function createOrganization(input: { name: string }) {
    setLoading(true);
    setError("");
    try {
      const created = await api.createOrganization(input);
      setOrganizations((rows) => [...rows, created]);
    } catch (cause) {
      setError(messageOf(cause));
      throw cause;
    } finally {
      setLoading(false);
    }
  }

  async function deleteOrganization(item: Organization) {
    setLoading(true);
    setError("");
    try {
      await api.deleteOrganization(item.id);
      setOrganizations((rows) => rows.filter((row) => row.id !== item.id));
    } finally {
      setLoading(false);
    }
  }

  async function inviteCounselor(input: { email: string; role: "counselor" | "counselor_admin" }) {
    setError("");
    try {
      const invitation: Invitation = await api.createCounselorInvitation(input);
      setCounselorInvitationURL(invitation.invitationURL || "");
    } catch (cause) {
      setError(messageOf(cause));
    }
  }

  async function updateCounselor(userID: string, input: { role: "counselor" | "counselor_admin"; status: "active" | "disabled" }) {
    setError("");
    try {
      const updated = await api.updateCounselor(userID, input);
      setCounselors((rows) => rows.map((row) => row.id === updated.id ? updated : row));
    } catch (cause) {
      setError(messageOf(cause));
      throw cause;
    }
  }

  async function logout() {
    try {
      await api.logout();
    } finally {
      router.replace("/login");
    }
  }

  function retryLoad() {
    setUser(null);
    setOrganizations([]);
    setCounselors([]);
    setError("");
    setAuthChecked(false);
    setRetryCount((value) => value + 1);
  }

  if (!authChecked) return <main className="screen-center" role="status" aria-live="polite" aria-busy="true">Loading organizations…</main>;
  if (!user && error) {
    return (
      <main className="screen-center" aria-labelledby="organizations-load-error">
        <section className="empty-state" role="alert">
          <h1 id="organizations-load-error">Could not load organizations</h1>
          <p>{error}</p>
          <button className="primary" type="button" onClick={retryLoad}>Try again</button>
        </section>
      </main>
    );
  }
  if (!user) return <main className="screen-center" role="status" aria-live="polite">Redirecting to sign in…</main>;

  return (
    <OrganizationDashboard
      user={user}
      organizations={organizations}
      loading={loading}
      error={error}
      onSelect={(organization) => router.push(organizationPath(organization))}
      onCreate={createOrganization}
      onDelete={deleteOrganization}
      onLogout={() => void logout()}
      counselors={counselors}
      counselorInvitationURL={counselorInvitationURL}
      onInviteCounselor={inviteCounselor}
      onUpdateCounselor={updateCounselor}
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
