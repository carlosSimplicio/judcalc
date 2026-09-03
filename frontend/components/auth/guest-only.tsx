"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { getSession } from "@/lib/auth/session";

export function GuestOnly({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [canRender, setCanRender] = useState(false);

  useEffect(() => {
    if (getSession()) {
      router.replace("/");
      return;
    }

    setCanRender(true);
  }, [router]);

  if (!canRender) {
    return (
      <main className="session-loading" role="status">
        Carregando...
      </main>
    );
  }

  return children;
}
