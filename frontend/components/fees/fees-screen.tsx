"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AppPageHeader } from "@/components/app/app-page-header";
import { AppShell } from "@/components/app/app-shell";
import type { AuthSession } from "@/lib/auth/client";
import { clearSession, getSession } from "@/lib/auth/session";
import { FeesCalculator } from "./fees-calculator";

export function FeesScreen() {
  const router = useRouter();
  const [session, setSession] = useState<AuthSession | null>(null);

  useEffect(() => {
    const currentSession = getSession();
    if (!currentSession) {
      router.replace("/login");
      return;
    }
    setSession(currentSession);
  }, [router]);

  const handleUnauthorized = useCallback(() => {
    clearSession();
    router.replace("/login");
  }, [router]);

  if (!session) {
    return <main className="session-loading" role="status">Carregando...</main>;
  }

  return (
    <AppShell
      activeItem="fees"
      tip="Quanto mais precisas as informações, mais confiável será a estimativa."
    >
      <AppPageHeader
        eyebrow="Precificação jurídica"
        title="Calcular honorários"
        description="Estime um valor justo considerando o serviço, o tempo de trabalho e as características do caso."
        user={session.user}
      />
      <FeesCalculator token={session.access_token} onUnauthorized={handleUnauthorized} />
    </AppShell>
  );
}
