"use client";

import Link from "next/link";
import { useEffect, useState, type ReactNode } from "react";
import { AppPageHeader } from "@/components/app/app-page-header";
import { AppShell } from "@/components/app/app-shell";
import type { AuthSession } from "@/lib/auth/client";
import { getSession } from "@/lib/auth/session";

export function HomeScreen() {
  const [session, setSession] = useState<AuthSession | null>();

  useEffect(() => {
    setSession(getSession());
  }, []);

  if (session === undefined) {
    return (
      <main className="session-loading" role="status">
        Carregando...
      </main>
    );
  }

  if (!session) {
    return <PublicHome />;
  }

  return <AuthenticatedHome session={session} />;
}

function PublicHome() {
  return (
    <div className="landing-shell">
      <header className="landing-header">
        <Link className="landing-brand" href="/" aria-label="JudCalc — início">
          <BrandIcon />
          <span>JudCalc</span>
        </Link>
        <nav className="landing-nav" aria-label="Acesso à conta">
          <Link className="landing-login" href="/login">
            Entrar
          </Link>
          <Link className="landing-button landing-button-small" href="/cadastro">
            Criar conta
          </Link>
        </nav>
      </header>

      <main className="landing-main">
        <section className="landing-hero" aria-labelledby="landing-title">
          <div className="landing-hero-copy">
            <p className="eyebrow">Gestão para a advocacia</p>
            <h1 id="landing-title">
              Honorários mais seguros. Decisões mais claras.
            </h1>
            <p>
              Organize os custos do escritório e encontre uma referência mais
              confiável para precificar seus serviços jurídicos.
            </p>
            <div className="landing-actions">
              <Link className="landing-button" href="/cadastro">
                Começar agora
              </Link>
              <Link className="landing-button landing-button-secondary" href="/login">
                Já tenho uma conta
              </Link>
            </div>
          </div>

          <div className="landing-highlight" aria-label="Recursos do JudCalc">
            <span className="landing-highlight-icon">
              <CalculatorIcon />
            </span>
            <p>Uma base mais clara para sua precificação</p>
            <strong>Custos + serviço + características do caso</strong>
            <small>Reúna os principais fatores em um único cálculo.</small>
          </div>
        </section>

        <section className="landing-benefits" aria-label="Benefícios">
          <BenefitCard
            icon={<PricingIcon />}
            title="Precificação orientada"
            description="Considere o serviço e a complexidade de cada caso."
          />
          <BenefitCard
            icon={<CostsIcon />}
            title="Custos sob controle"
            description="Mantenha as despesas recorrentes organizadas."
          />
          <BenefitCard
            icon={<AccessIcon />}
            title="Acesso responsivo"
            description="Use o JudCalc no computador, tablet ou celular."
          />
        </section>
      </main>

      <footer className="landing-footer">
        <span>JudCalc</span>
        <small>Mais clareza para decisões financeiras na advocacia.</small>
      </footer>
    </div>
  );
}

function AuthenticatedHome({ session }: { session: AuthSession }) {
  const firstName = session.user.name.trim().split(/\s+/)[0] || session.user.name;

  return (
    <AppShell
      activeItem="home"
      tip="Mantenha seus custos atualizados antes de calcular novos honorários."
    >
      <AppPageHeader
        eyebrow="Visão geral"
        title={`Olá, ${firstName}`}
        description="Escolha uma opção para começar a organizar sua precificação."
        user={session.user}
      />

      <section className="home-welcome" aria-labelledby="home-welcome-title">
        <div>
          <p>Seu espaço de trabalho</p>
          <h2 id="home-welcome-title">O que você deseja fazer?</h2>
          <span>
            Acesse as principais ferramentas do JudCalc pelos atalhos abaixo.
          </span>
        </div>
        <span className="home-welcome-icon">
          <CalculatorIcon />
        </span>
      </section>

      <section className="home-actions" aria-label="Ações rápidas">
        <ActionCard
          href="/honorarios"
          icon={<PricingIcon />}
          tone="blue"
          title="Calcular honorários"
          description="Crie uma estimativa considerando serviço, tempo e características do caso."
          linkLabel="Iniciar cálculo"
        />
        <ActionCard
          href="/custos-fixos"
          icon={<CostsIcon />}
          tone="teal"
          title="Cadastrar custos fixos"
          description="Informe suas despesas recorrentes para tornar os cálculos mais confiáveis."
          linkLabel="Atualizar custos"
        />
      </section>
    </AppShell>
  );
}

function BenefitCard({
  icon,
  title,
  description,
}: {
  icon: ReactNode;
  title: string;
  description: string;
}) {
  return (
    <article className="landing-benefit-card">
      <span>{icon}</span>
      <h2>{title}</h2>
      <p>{description}</p>
    </article>
  );
}

function ActionCard({
  href,
  icon,
  tone,
  title,
  description,
  linkLabel,
}: {
  href: string;
  icon: ReactNode;
  tone: "blue" | "teal";
  title: string;
  description: string;
  linkLabel: string;
}) {
  return (
    <Link className="home-action-card" href={href}>
      <span className={`home-action-icon home-action-icon-${tone}`}>{icon}</span>
      <h2>{title}</h2>
      <p>{description}</p>
      <strong>
        {linkLabel}
        <ArrowIcon />
      </strong>
    </Link>
  );
}

function BrandIcon() {
  return (
    <svg width="38" height="38" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="2.5" aria-hidden="true">
      <path d="M24 6v36M12 15h24M15 15l-7 18h14l-7-18ZM33 15l-7 18h14l-7-18ZM16 42h16" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function PricingIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="m4 14 4-4 3 3 7-7M14 6h4v4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function CostsIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M6 2v3M18 2v3M3 9h18M5 4h14a2 2 0 0 1 2 2v14H3V6a2 2 0 0 1 2-2Z" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M8 14h8M8 17h5" strokeLinecap="round" />
    </svg>
  );
}

function AccessIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <circle cx="12" cy="12" r="8" />
      <path d="M4 12h16M12 4a12 12 0 0 1 0 16M12 4a12 12 0 0 0 0 16" strokeLinecap="round" />
    </svg>
  );
}

function CalculatorIcon() {
  return (
    <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <rect x="5" y="2" width="14" height="20" rx="2" />
      <path d="M8 6h8v4H8zM8 14h.01M12 14h.01M16 14h.01M8 18h.01M12 18h.01M16 18h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function ArrowIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M5 12h14m-5-5 5 5-5 5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
