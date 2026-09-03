import type { ReactNode } from "react";
import Link from "next/link";

export type AppItem = "home" | "fees" | "fixed-costs" | "profile";

type AppShellProps = {
  activeItem: AppItem;
  children: ReactNode;
  tip?: string;
};

const navigationItems: Array<{
  id: AppItem;
  desktopLabel: string;
  mobileLabel: string;
  href?: string;
}> = [
  { id: "home", desktopLabel: "Início", mobileLabel: "Início", href: "/" },
  { id: "fees", desktopLabel: "Honorários", mobileLabel: "Honorários", href: "/honorarios" },
  { id: "fixed-costs", desktopLabel: "Custos fixos", mobileLabel: "Custos", href: "/custos-fixos" },
  { id: "profile", desktopLabel: "Perfil", mobileLabel: "Perfil" },
];

export function AppShell({ activeItem, children, tip = "Revise seus custos sempre que uma despesa recorrente mudar." }: AppShellProps) {
  return (
    <div className="app-shell">
      <aside className="app-sidebar" aria-label="Navegação principal">
        <div className="app-brand" aria-label="JudCalc">
          <LogoIcon />
          <span>JudCalc</span>
        </div>

        <nav className="sidebar-nav">
          {navigationItems.map((item) => (
            <NavigationButton
              activeItem={activeItem}
              item={item}
              key={item.id}
              label={item.desktopLabel}
              size={22}
            />
          ))}
        </nav>

        <div className="sidebar-tip">
          <TipIcon />
          <p>
            <strong>Dica</strong>
            {tip}
          </p>
        </div>
      </aside>

      <main className="app-main">{children}</main>

      <nav className="mobile-nav" aria-label="Navegação principal para dispositivos móveis">
        {navigationItems.map((item) => (
          <NavigationButton
            activeItem={activeItem}
            item={item}
            key={item.id}
            label={item.mobileLabel}
            size={21}
          />
        ))}
      </nav>
    </div>
  );
}

type NavigationButtonProps = {
  activeItem: AppItem;
  item: (typeof navigationItems)[number];
  label: string;
  size: number;
};

function NavigationButton({ activeItem, item, label, size }: NavigationButtonProps) {
  const isActive = item.id === activeItem;

  if (item.href) {
    return (
      <Link
        aria-current={isActive ? "page" : undefined}
        className={isActive ? "is-active" : undefined}
        href={item.href}
      >
        <NavigationIcon item={item.id} size={size} />
        {label}
      </Link>
    );
  }

  return (
    <button
      aria-current={isActive ? "page" : undefined}
      className={isActive ? "is-active" : undefined}
      disabled={isActive}
      type="button"
    >
      <NavigationIcon item={item.id} size={size} />
      {label}
    </button>
  );
}

function LogoIcon() {
  return (
    <svg width="36" height="36" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="2.5" aria-hidden="true">
      <path d="M24 6v36M12 15h24M15 15l-7 18h14l-7-18ZM33 15l-7 18h14l-7-18ZM16 42h16" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function TipIcon() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <circle cx="12" cy="12" r="9" />
      <path d="M9.5 9a2.6 2.6 0 1 1 4.1 2.1c-.9.6-1.6 1.1-1.6 2.4M12 17h.01" strokeLinecap="round" />
    </svg>
  );
}

function NavigationIcon({ item, size }: { item: AppItem; size: number }) {
  const commonProps = {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: "2",
    "aria-hidden": true,
  };

  if (item === "home") {
    return <svg {...commonProps}><path d="m3 11 9-8 9 8v9a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1z" strokeLinejoin="round" /></svg>;
  }

  if (item === "fees") {
    return <svg {...commonProps}><path d="M4 4h16v16H4zM8 8h8M8 12h3M8 16h8" strokeLinecap="round" /></svg>;
  }

  if (item === "fixed-costs") {
    return <svg {...commonProps}><path d="M6 2v3M18 2v3M3 9h18M5 4h14a2 2 0 0 1 2 2v14H3V6a2 2 0 0 1 2-2Z" strokeLinecap="round" strokeLinejoin="round" /><path d="M8 14h8M8 17h5" strokeLinecap="round" /></svg>;
  }

  return <svg {...commonProps}><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" strokeLinecap="round" /></svg>;
}
