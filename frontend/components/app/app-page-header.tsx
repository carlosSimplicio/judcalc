import type { AuthUser } from "@/lib/auth/client";

type AppPageHeaderProps = {
  eyebrow: string;
  title: string;
  description: string;
  user: AuthUser;
};

export function AppPageHeader({ eyebrow, title, description, user }: AppPageHeaderProps) {
  return (
    <header className="app-header">
      <div>
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      <button className="profile-button" type="button" aria-label="Abrir menu do perfil">
        <span className="profile-avatar">{userInitials(user.name)}</span>
        <span className="profile-copy">
          <strong>{user.name}</strong>
          <small>Meu perfil</small>
        </span>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
          <path d="m7 10 5 5 5-5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>
    </header>
  );
}

function userInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  return parts.slice(0, 2).map((part) => part[0].toLocaleUpperCase("pt-BR")).join("") || "?";
}
