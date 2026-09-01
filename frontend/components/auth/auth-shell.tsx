import type { ReactNode } from "react";

type AuthShellProps = {
  titleId: string;
  title: string;
  description: string;
  children: ReactNode;
  footer: ReactNode;
};

const benefits = [
  { label: "Precificação orientada", icon: <PricingIcon /> },
  { label: "Custos sob controle", icon: <SecurityIcon /> },
  { label: "Acesso em qualquer dispositivo", icon: <AccessIcon /> },
];

export function AuthShell({
  titleId,
  title,
  description,
  children,
  footer,
}: AuthShellProps) {
  return (
    <main className="auth-shell">
      <aside className="brand-panel" aria-labelledby="brand-title">
        <div>
          <BrandIcon />
          <h1 id="brand-title">JudCalc</h1>
          <p className="brand-copy">
            Honorários mais seguros. Decisões mais claras.
          </p>
        </div>
        <ul className="benefit-list" aria-label="Benefícios">
          {benefits.map((benefit) => (
            <li key={benefit.label}>
              {benefit.icon}
              <span>{benefit.label}</span>
            </li>
          ))}
        </ul>
      </aside>

      <section className="form-panel" aria-labelledby={titleId}>
        <div className="auth-card">
          <h2 id={titleId}>{title}</h2>
          <p className="auth-description">{description}</p>
          {children}
          <p className="auth-footer">{footer}</p>
        </div>
      </section>
    </main>
  );
}

function BrandIcon() {
  return (
    <svg
      className="brand-icon"
      width="48"
      height="48"
      viewBox="0 0 48 48"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      aria-hidden="true"
    >
      <path
        d="M24 6v36M12 15h24M15 15l-7 18h14l-7-18ZM33 15l-7 18h14l-7-18ZM16 42h16"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function PricingIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="m4 14 4-4 3 3 7-7M14 6h4v4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SecurityIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M5 11h14v9H5zM8 11V8a4 4 0 0 1 8 0v3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function AccessIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <circle cx="12" cy="12" r="8" />
      <path d="M4 12h16M12 4a12 12 0 0 1 0 16M12 4a12 12 0 0 0 0 16" strokeLinecap="round" />
    </svg>
  );
}
