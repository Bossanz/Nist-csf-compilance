"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { APIError, api } from "../../lib/api";
import type { Organization, User } from "../../lib/types";
import { OrganizationDashboard } from "../../components/OrganizationDashboard";
import { organizationPath } from "../../lib/routes";

export default function OrganizationsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [authChecked, setAuthChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void initialize(active);
    return () => { active = false; };
  }, []);

  async function initialize(active: boolean) {
    try {
      const [currentUser, organizationRows] = await Promise.all([api.me(), api.getOrganizations()]);
      if (!active) return;
      setUser(currentUser);
      setOrganizations(organizationRows);
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

  async function logout() {
    try {
      await api.logout();
    } finally {
      router.replace("/login");
    }
  }

  if (!authChecked || !user) return <main className="screen-center">Loading...</main>;

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
    />
  );
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
