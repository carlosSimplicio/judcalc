"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AppPageHeader } from "@/components/app/app-page-header";
import { AppShell } from "@/components/app/app-shell";
import { useToast } from "@/components/ui/toast-provider";
import { ApiError } from "@/lib/api/client";
import type { AuthSession } from "@/lib/auth/client";
import { apiErrorToast } from "@/lib/auth/error-toast";
import { clearSession, getSession } from "@/lib/auth/session";
import { getFixedCosts, type FixedCosts } from "@/lib/fixed-costs/client";
import { FixedCostsForm } from "./fixed-costs-form";

export function FixedCostsScreen() {
  const router = useRouter();
  const { showToast } = useToast();
  const [session, setSession] = useState<AuthSession | null>(null);
  const [costs, setCosts] = useState<FixedCosts | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadFailed, setLoadFailed] = useState(false);

  const handleUnauthorized = useCallback(() => {
    clearSession();
    router.replace("/login");
  }, [router]);

  useEffect(() => {
    const currentSession = getSession();
    if (!currentSession) {
      router.replace("/login");
      return;
    }

    setSession(currentSession);
    getFixedCosts(currentSession.access_token)
      .then((result) => {
        setCosts(result);
        setLoadFailed(false);
      })
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 401) {
          handleUnauthorized();
          return;
        }

        setLoadFailed(true);
        showToast(apiErrorToast(error));
      })
      .finally(() => setLoading(false));
  }, [handleUnauthorized, router, showToast]);

  if (!session) {
    return (
      <main className="session-loading" role="status">
        Carregando...
      </main>
    );
  }

  return (
    <AppShell activeItem="fixed-costs">
      <AppPageHeader
        eyebrow="Configurações financeiras"
        title="Custos fixos"
        description="Informe suas despesas recorrentes para calcular honorários com mais segurança."
        user={session.user}
      />
      {loading ? (
        <p className="app-loading" role="status">
          Carregando custos...
        </p>
      ) : null}
      {loadFailed ? (
        <div className="app-load-error" role="status">
          Não foi possível carregar os custos. Atualize a página para tentar
          novamente.
        </div>
      ) : null}
      <FixedCostsForm
        token={session.access_token}
        initialCosts={costs}
        loading={loading}
        onUnauthorized={handleUnauthorized}
      />
    </AppShell>
  );
}
