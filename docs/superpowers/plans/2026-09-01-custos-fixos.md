# Cadastro de custos fixos Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Entregar a rota autenticada `/custos-fixos`, fiel ao protótipo, com carregamento, total mensal, salvamento e restauração dos custos fixos.

**Architecture:** A funcionalidade usa um cliente HTTP comum para normalizar respostas e erros, um cliente tipado de custos fixos, uma moldura reutilizável de páginas internas e um componente cliente que coordena sessão e formulário. Valores trafegam em centavos; os inputs mantêm strings monetárias formatadas e são convertidos apenas nas bordas.

**Tech Stack:** Next.js 16 App Router, React 19, TypeScript 5.9, CSS global, Fetch API.

**Spec:** `docs/superpowers/specs/2026-09-01-custos-fixos-design.md`

## Global Constraints

- A página deve seguir `frontend/prototypes/custos-fixos.html` em desktop e mobile.
- A rota pública da funcionalidade é `/custos-fixos`.
- GET e PATCH usam `/api/v1/fixed-costs` com o token Bearer da sessão.
- Valores da API são inteiros em centavos; negativos não são aceitos.
- “Restaurar” repõe a última resposta válida da API sem realizar nova requisição.
- Itens destinados a páginas inexistentes permanecem visíveis, mas não navegam.
- Uma resposta `401` limpa a sessão e redireciona para `/login`.
- Por decisão do usuário, não adicionar testes unitários nem dependências de teste nesta etapa.
- Não modificar backend, banco de dados ou a seção de banco do `AGENTS.md`.

---

### Task 1: Extrair o cliente HTTP e generalizar feedback da API

**Files:**
- Create: `frontend/lib/api/client.ts`
- Modify: `frontend/lib/auth/client.ts`
- Modify: `frontend/lib/auth/error-toast.ts`
- Modify: `frontend/components/auth/login-form.tsx`
- Modify: `frontend/components/auth/signup-form.tsx`
- Modify: `frontend/components/ui/toast-provider.tsx`
- Modify: `frontend/app/globals.css`

**Interfaces:**
- Produces: `apiRequest(path: string, init?: ApiRequestInit): Promise<unknown>`
- Produces: `ApiError`, `unexpectedResponse(status: number)`, `isRecord(value: unknown)`
- Produces: `apiErrorToast(error: unknown): ToastInput`
- Produces: `ToastInput` com `tone: "error" | "success"`

- [ ] **Step 1: Criar o cliente HTTP compartilhado**

Criar `frontend/lib/api/client.ts` com URL base, envelope de erro e tratamento uniforme de rede:

```ts
export type ErrorEnvelope = {
  error: { code: string; message: string };
};

export type ApiRequestInit = {
  method?: "GET" | "POST" | "PATCH";
  token?: string;
  body?: unknown;
};

const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL
).replace(/\/$/, "");

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export async function apiRequest(
  path: string,
  init: ApiRequestInit = {},
): Promise<unknown> {
  const headers = new Headers();
  if (init.body !== undefined) headers.set("Content-Type", "application/json");
  if (init.token) headers.set("Authorization", `Bearer ${init.token}`);

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: init.method ?? "GET",
      headers,
      body: init.body === undefined ? undefined : JSON.stringify(init.body),
    });
  } catch {
    throw new ApiError(
      0,
      "network_error",
      "Não foi possível conectar ao servidor. Tente novamente.",
    );
  }

  const payload: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    if (isErrorEnvelope(payload)) {
      throw new ApiError(response.status, payload.error.code, payload.error.message);
    }
    throw unexpectedResponse(response.status);
  }
  return payload;
}

export function unexpectedResponse(status: number): ApiError {
  return new ApiError(
    status,
    "unexpected_response",
    "O servidor retornou uma resposta inesperada.",
  );
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    isRecord(value) &&
    isRecord(value.error) &&
    typeof value.error.code === "string" &&
    typeof value.error.message === "string"
  );
}
```

- [ ] **Step 2: Refatorar autenticação para consumir o cliente comum**

Em `frontend/lib/auth/client.ts`, remover URL base, Fetch e definições duplicadas. Importar `apiRequest`, `ApiError`, `unexpectedResponse` e `isRecord`; manter os tipos públicos e o type guard da sessão. `requestSession` passa a ser:

```ts
async function requestSession(
  path: string,
  input: SignInInput | SignUpInput,
): Promise<AuthSession> {
  const payload = await apiRequest(path, { method: "POST", body: input });
  if (!isSessionEnvelope(payload)) {
    throw unexpectedResponse(200);
  }
  return payload.data;
}

export { ApiError } from "@/lib/api/client";
```

- [ ] **Step 3: Tornar o toast adequado a sucesso e erro**

Exportar `ToastInput` em `toast-provider.tsx`, incluindo `tone`, e usar a classe correspondente:

```ts
export type ToastInput = {
  title: string;
  message: string;
  tone: "error" | "success";
};

<div className={`toast toast-${toast.tone}`} role="status" aria-live="polite">
```

Renomear `authErrorToast` para `apiErrorToast` em `error-toast.ts`, importar
`ApiError` de `@/lib/api/client`, incluir `tone: "error"` nos retornos e atualizar
os imports de login e cadastro. Acrescentar `unauthorized: "Sessão expirada"` ao
mapa de títulos.

- [ ] **Step 4: Ajustar os estilos de toast sem regressão visual**

Manter o toast de erro vermelho e adicionar sucesso em teal:

```css
.toast-error {
  border-color: color-mix(in srgb, var(--danger) 35%, var(--border));
  border-left-color: var(--danger);
}

.toast-success {
  border-color: color-mix(in srgb, var(--teal) 35%, var(--border));
  border-left-color: var(--teal);
}
```

Mover do seletor `.toast` apenas as cores dependentes do tom.

- [ ] **Step 5: Verificar compilação e autenticação existente**

Run: `cd frontend && npm run build`

Expected: build concluído, rotas `/`, `/login` e `/cadastro` listadas, sem erros TypeScript.

- [ ] **Step 6: Commit**

```bash
git add frontend/lib/api/client.ts frontend/lib/auth/client.ts frontend/lib/auth/error-toast.ts frontend/components/auth/login-form.tsx frontend/components/auth/signup-form.tsx frontend/components/ui/toast-provider.tsx frontend/app/globals.css
git commit -m "refactor: share frontend api handling"
```

---

### Task 2: Implementar contrato e conversões de custos fixos

**Files:**
- Create: `frontend/lib/fixed-costs/client.ts`
- Create: `frontend/lib/fixed-costs/money.ts`

**Interfaces:**
- Consumes: `apiRequest`, `unexpectedResponse`, `isRecord` da Task 1
- Produces: `FixedCosts`, `FixedCostsValues`, `FixedCostKey`
- Produces: `getFixedCosts(token: string): Promise<FixedCosts>`
- Produces: `saveFixedCosts(token: string, values: FixedCostsValues): Promise<FixedCosts>`
- Produces: `formatCentsInput`, `parseMoneyInput`, `formatCurrency`, `monthlyTotal`

- [ ] **Step 1: Definir tipos e lista canônica de campos**

Criar em `frontend/lib/fixed-costs/client.ts`:

```ts
export const MONTHLY_COST_KEYS = [
  "digital_certificate",
  "accountant",
  "legal_software",
  "internet",
  "phone",
  "recurring_transport",
  "coworking_or_office_rent",
  "professional_domain_website_email",
  "marketing",
  "office_supplies",
  "equipment_and_depreciation",
  "other_costs",
] as const;

export type MonthlyCostKey = (typeof MONTHLY_COST_KEYS)[number];
export type FixedCostKey = "oab_annual_fee" | MonthlyCostKey;

export type FixedCostsValues = Record<FixedCostKey, number>;

export type FixedCosts = {
  userId: number;
  values: FixedCostsValues;
};
```

- [ ] **Step 2: Implementar validação da resposta e mapeamento para o domínio do formulário**

Validar integralmente `data.user_id`, os 13 objetos e seus inteiros não negativos.
Mapear `annual_amount_cents` da OAB e `monthly_amount_cents` dos demais campos
para `FixedCostsValues`. Qualquer ausência, número fracionário ou negativo deve
lançar `unexpectedResponse(status)`.

```ts
function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}
```

- [ ] **Step 3: Implementar GET e PATCH tipados**

```ts
export async function getFixedCosts(token: string): Promise<FixedCosts> {
  const payload = await apiRequest("/fixed-costs", { token });
  return parseFixedCostsEnvelope(payload, 200);
}

export async function saveFixedCosts(
  token: string,
  values: FixedCostsValues,
): Promise<FixedCosts> {
  const monthlyCosts = Object.fromEntries(
    MONTHLY_COST_KEYS.map((key) => [
      key,
      { monthly_amount_cents: values[key] },
    ]),
  );
  const payload = await apiRequest("/fixed-costs", {
    method: "PATCH",
    token,
    body: {
      costs: {
        oab_annual_fee: {
          annual_amount_cents: values.oab_annual_fee,
        },
        ...monthlyCosts,
      },
    },
  });
  return parseFixedCostsEnvelope(payload, 200);
}
```

- [ ] **Step 4: Implementar funções monetárias puras**

Em `money.ts`, usar centavos inteiros e locale fixo:

```ts
const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

export function formatCentsInput(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function parseMoneyInput(value: string): number {
  const digits = value.replace(/\D/g, "");
  return digits ? Number.parseInt(digits, 10) : 0;
}

export function normalizeMoneyInput(value: string): string {
  return formatCentsInput(parseMoneyInput(value));
}

export function formatCurrency(cents: number): string {
  return currencyFormatter.format(cents / 100);
}

export function monthlyTotal(values: FixedCostsValues): number {
  const oabMonthly = Math.floor(values.oab_annual_fee / 12) +
    (values.oab_annual_fee % 12 >= 6 ? 1 : 0);
  return MONTHLY_COST_KEYS.reduce(
    (total, key) => total + values[key],
    oabMonthly,
  );
}
```

- [ ] **Step 5: Verificar tipos**

Run: `cd frontend && npx tsc --noEmit`

Expected: saída vazia e exit code 0.

- [ ] **Step 6: Commit**

```bash
git add frontend/lib/fixed-costs/client.ts frontend/lib/fixed-costs/money.ts
git commit -m "feat: add fixed costs api client"
```

---

### Task 3: Criar a moldura reutilizável das páginas internas

**Files:**
- Create: `frontend/components/app/app-shell.tsx`
- Create: `frontend/components/app/app-page-header.tsx`

**Interfaces:**
- Consumes: `AuthUser` de `frontend/lib/auth/client.ts`
- Produces: `AppShell({ activeItem, children })`
- Produces: `AppPageHeader({ eyebrow, title, description, user })`

- [ ] **Step 1: Criar `AppShell` com navegação sem destinos falsos**

Definir `activeItem` como `"home" | "fees" | "fixed-costs" | "profile"`.
Renderizar marca, dica e as duas navegações a partir de uma lista interna. Cada
item usa `<button type="button">`; o ativo recebe `is-active`,
`aria-current="page"` e `disabled`. Os demais continuam acionáveis visualmente,
mas não recebem `onClick` e não alteram a URL.

```tsx
type AppItem = "home" | "fees" | "fixed-costs" | "profile";

type AppShellProps = {
  activeItem: AppItem;
  children: ReactNode;
};

const navigationItems: Array<{
  id: AppItem;
  desktopLabel: string;
  mobileLabel: string;
}> = [
  { id: "home", desktopLabel: "Início", mobileLabel: "Início" },
  { id: "fees", desktopLabel: "Honorários", mobileLabel: "Honorários" },
  { id: "fixed-costs", desktopLabel: "Custos fixos", mobileLabel: "Custos" },
  { id: "profile", desktopLabel: "Perfil", mobileLabel: "Perfil" },
];
```

O JSX raiz contém, nesta ordem, `<aside className="app-sidebar">`,
`<main className="app-main">{children}</main>` e
`<nav className="mobile-nav">`. As duas navegações mapeiam `navigationItems` e
usam `desktopLabel` ou `mobileLabel`; cada botão compara `item.id === activeItem`
para definir classe, `disabled` e `aria-current`.

Usar os paths SVG do protótipo diretamente como componentes internos sem
dependência externa de ícones.

- [ ] **Step 2: Criar cabeçalho de página e iniciais robustas**

`AppPageHeader` renderiza eyebrow, `h1`, descrição e o botão de perfil sem ação.
Gerar até duas iniciais a partir das primeiras letras de palavras não vazias; se
o nome estiver vazio, usar `?`.

```ts
function userInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  return parts.slice(0, 2).map((part) => part[0].toLocaleUpperCase("pt-BR")).join("") || "?";
}
```

- [ ] **Step 3: Verificar tipos**

Run: `cd frontend && npx tsc --noEmit`

Expected: saída vazia e exit code 0.

- [ ] **Step 4: Commit**

```bash
git add frontend/components/app/app-shell.tsx frontend/components/app/app-page-header.tsx
git commit -m "feat: add reusable application shell"
```

---

### Task 4: Construir os componentes do formulário de custos

**Files:**
- Create: `frontend/components/fixed-costs/fields.tsx`
- Create: `frontend/components/fixed-costs/money-field.tsx`
- Create: `frontend/components/fixed-costs/fixed-costs-form.tsx`

**Interfaces:**
- Consumes: tipos e funções de `frontend/lib/fixed-costs/`
- Produces: `COST_SECTIONS` com os 13 campos em três grupos
- Produces: `MoneyField({ field, value, disabled, onChange })`
- Produces: `FixedCostsForm({ token, initialCosts, onUnauthorized })`

- [ ] **Step 1: Descrever as três seções e os 13 campos**

Em `fields.tsx`, criar `CostFieldDefinition` e `CostSectionDefinition`. Preencher:

```ts
export const COST_SECTIONS: CostSectionDefinition[] = [
  {
    id: "professional-costs-title",
    title: "Estrutura profissional",
    description: "Obrigações e ferramentas essenciais para o escritório.",
    tone: "blue",
    fields: [
      { key: "oab_annual_fee", label: "Anuidade da OAB", help: "Informe o valor anual. Consideraremos a média mensal." },
      { key: "digital_certificate", label: "Certificado digital", help: "Valor mensal ou média do período contratado." },
      { key: "accountant", label: "Contador" },
      { key: "legal_software", label: "Software jurídico" },
    ],
  },
  {
    id: "office-costs-title",
    title: "Escritório e operação",
    description: "Despesas para manter sua rotina de trabalho.",
    tone: "teal",
    fields: [
      { key: "internet", label: "Internet" },
      { key: "phone", label: "Telefone" },
      { key: "recurring_transport", label: "Transporte recorrente" },
      { key: "coworking_or_office_rent", label: "Coworking ou aluguel" },
      { key: "professional_domain_website_email", label: "Domínio, site e e-mail profissional" },
      { key: "marketing", label: "Marketing" },
    ],
  },
  {
    id: "other-costs-title",
    title: "Materiais e outros custos",
    description: "Complete o cadastro com as demais despesas mensais.",
    tone: "orange",
    fields: [
      { key: "office_supplies", label: "Materiais de escritório" },
      { key: "equipment_and_depreciation", label: "Equipamentos e depreciação" },
      { key: "other_costs", label: "Outros custos", help: "Use este campo para despesas recorrentes não listadas acima.", wide: true },
    ],
  },
];
```

Cada seção inclui o SVG correspondente do protótipo.

- [ ] **Step 2: Criar o input monetário reutilizável**

O input é controlado, recebe string formatada e normaliza no `onChange`:

```tsx
export function MoneyField({ field, value, disabled, onChange }: MoneyFieldProps) {
  const id = `fixed-cost-${field.key.replaceAll("_", "-")}`;
  const helpId = field.help ? `${id}-help` : undefined;
  return (
    <div className={`cost-field${field.wide ? " cost-field-wide" : ""}`}>
      <label htmlFor={id}>{field.label}</label>
      <div className="money-input">
        <span aria-hidden="true">R$</span>
        <input
          id={id}
          name={field.key}
          type="text"
          inputMode="decimal"
          value={value}
          disabled={disabled}
          aria-describedby={helpId}
          onChange={(event) => onChange(normalizeMoneyInput(event.target.value))}
        />
      </div>
      {field.help ? <small id={helpId}>{field.help}</small> : null}
    </div>
  );
}
```

- [ ] **Step 3: Implementar estado, total, salvar e restaurar**

`FixedCostsForm` recebe os custos carregados. Armazena `savedValues` em centavos
e `inputs` como `Record<FixedCostKey, string>`. O valor derivado atual converte
cada input por `parseMoneyInput`; o total usa `monthlyTotal`.

No submit:

```ts
setSaving(true);
try {
  const result = await saveFixedCosts(token, currentValues);
  setSavedValues(result.values);
  setInputs(valuesToInputs(result.values));
  showToast({
    tone: "success",
    title: "Custos salvos",
    message: "Seus custos fixos foram atualizados.",
  });
} catch (error) {
  if (error instanceof ApiError && error.status === 401) {
    onUnauthorized();
    return;
  }
  showToast(apiErrorToast(error));
} finally {
  setSaving(false);
}
```

“Restaurar” executa `setInputs(valuesToInputs(savedValues))`. Desabilitar ambos
os botões durante save; usar `type="button"` para restaurar e `type="submit"`
para salvar. Incluir os SVGs do resumo, das seções e de salvar conforme o
protótipo.

- [ ] **Step 4: Verificar tipos**

Run: `cd frontend && npx tsc --noEmit`

Expected: saída vazia e exit code 0.

- [ ] **Step 5: Commit**

```bash
git add frontend/components/fixed-costs/fields.tsx frontend/components/fixed-costs/money-field.tsx frontend/components/fixed-costs/fixed-costs-form.tsx
git commit -m "feat: build fixed costs form"
```

---

### Task 5: Integrar sessão, página e estados de carregamento

**Files:**
- Create: `frontend/components/fixed-costs/fixed-costs-screen.tsx`
- Create: `frontend/app/custos-fixos/page.tsx`

**Interfaces:**
- Consumes: `getSession`, `clearSession`, `getFixedCosts`, `AppShell`, `AppPageHeader`, `FixedCostsForm`
- Produces: página completa em `/custos-fixos`

- [ ] **Step 1: Criar o coordenador autenticado da tela**

`FixedCostsScreen` usa `useRouter`, lê a sessão apenas em `useEffect` e evita
renderizar identidade ou formulário antes dessa validação. Com sessão válida,
executa GET e mostra a moldura com cabeçalho e uma região de status enquanto
carrega.

```tsx
const [session, setSession] = useState<AuthSession | null>(null);
const [costs, setCosts] = useState<FixedCosts | null>(null);
const [loading, setLoading] = useState(true);

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
    .then(setCosts)
    .catch((error) => {
      if (error instanceof ApiError && error.status === 401) {
        handleUnauthorized();
        return;
      }
      showToast(apiErrorToast(error));
    })
    .finally(() => setLoading(false));
}, [handleUnauthorized, router, showToast]);
```

Evitar repetição do GET causada por dependência instável: `showToast` já é
memoizado pelo provider e `handleUnauthorized` usa `useCallback`.

- [ ] **Step 2: Renderizar estados sem desmontar a moldura autenticada**

Quando há sessão:

```tsx
return (
  <AppShell activeItem="fixed-costs">
    <AppPageHeader
      eyebrow="Configurações financeiras"
      title="Custos fixos"
      description="Informe suas despesas recorrentes para calcular honorários com mais segurança."
      user={session.user}
    />
    {loading ? (
      <p className="app-loading" role="status">Carregando custos...</p>
    ) : costs ? (
      <FixedCostsForm
        token={session.access_token}
        initialCosts={costs}
        onUnauthorized={handleUnauthorized}
      />
    ) : (
      <div className="app-load-error" role="status">
        Não foi possível carregar os custos. Atualize a página para tentar novamente.
      </div>
    )}
  </AppShell>
);
```

Antes de validar a sessão, retornar `<main className="session-loading" role="status">Carregando...</main>`.

- [ ] **Step 3: Criar a rota e os metadados**

```tsx
import type { Metadata } from "next";
import { FixedCostsScreen } from "@/components/fixed-costs/fixed-costs-screen";

export const metadata: Metadata = {
  title: "Custos fixos | JudCalc",
  description: "Cadastre suas despesas recorrentes no JudCalc.",
};

export default function FixedCostsPage() {
  return <FixedCostsScreen />;
}
```

- [ ] **Step 4: Verificar tipos**

Run: `cd frontend && npx tsc --noEmit`

Expected: saída vazia e exit code 0.

- [ ] **Step 5: Commit**

```bash
git add frontend/components/fixed-costs/fixed-costs-screen.tsx frontend/app/custos-fixos/page.tsx
git commit -m "feat: add authenticated fixed costs page"
```

---

### Task 6: Adaptar o visual responsivo do protótipo

**Files:**
- Modify: `frontend/app/globals.css`

**Interfaces:**
- Consumes: classes produzidas pelas Tasks 3, 4 e 5
- Produces: layout desktop/mobile fiel ao protótipo sem alterar páginas de autenticação

- [ ] **Step 1: Adicionar variáveis e estilos estruturais**

Adicionar `--teal-soft: #e8f7f5`. Transportar do bloco “Fixed costs prototype”
de `frontend/prototypes/styles.css` os seletores `.app-shell`, `.app-sidebar`,
`.app-brand`, `.sidebar-nav`, `.sidebar-tip`, `.app-main`, `.app-header`,
`.profile-*`, `.cost-summary`, `.summary-*`, `.costs-form`, `.form-section`,
`.section-*`, `.cost-fields-grid`, `.cost-field`, `.money-input`,
`.form-actions` e `.mobile-nav`.

Trocar `.costs-main` do protótipo por `.app-main`. Adaptar navegação de `a` para
`button`:

```css
.sidebar-nav button,
.mobile-nav button {
  border: 0;
  color: var(--text-muted);
  background: transparent;
  font: inherit;
  cursor: default;
}

.sidebar-nav button:disabled,
.mobile-nav button:disabled {
  opacity: 1;
}
```

- [ ] **Step 2: Preservar os botões secundário e principal**

Adicionar `.button-secondary` e `.button-save` do protótipo. Garantir que
`.button:hover:not(:disabled)` continue controlando o botão principal e usar:

```css
.button-secondary:hover:not(:disabled) {
  color: var(--navy);
  background: #eef2f7;
}
```

- [ ] **Step 3: Adicionar carregamento e erro no mesmo sistema visual**

```css
.app-loading,
.app-load-error {
  padding: 2rem;
  border: 1px solid var(--border);
  border-radius: 1.1rem;
  color: var(--text-muted);
  background: var(--surface);
  text-align: center;
}

.session-loading {
  display: grid;
  min-height: 100svh;
  margin: 0;
  place-items: center;
  color: var(--text-muted);
}
```

- [ ] **Step 4: Implementar os breakpoints do protótipo**

Em `@media (max-width: 50rem)`, esconder sidebar, zerar margem de `.app-main`,
mostrar `.mobile-nav`, usar uma coluna de campos e ações sticky. Em
`@media (max-width: 27rem)`, reduzir cópias auxiliares e tamanho dos botões.
Manter intactas as regras existentes das páginas de autenticação.

- [ ] **Step 5: Executar o build completo**

Run: `cd frontend && npm run build`

Expected: build concluído e rota estática `/custos-fixos` listada sem erros ou warnings.

- [ ] **Step 6: Commit**

```bash
git add frontend/app/globals.css
git commit -m "style: match fixed costs prototype"
```

---

### Task 7: Verificação funcional e revisão final

**Files:**
- Modify only if verification exposes a defect: files introduced or modified in Tasks 1–6

**Interfaces:**
- Consumes: aplicação completa e backend existente
- Produces: evidência de compilação e checklist funcional concluído

- [ ] **Step 1: Confirmar que o escopo não alterou banco ou backend**

Run: `git diff HEAD~6 -- backend AGENTS.md`

Expected: saída vazia.

- [ ] **Step 2: Confirmar qualidade do diff**

Run: `git diff --check HEAD~6..HEAD`

Expected: saída vazia e exit code 0.

- [ ] **Step 3: Executar verificação final do frontend**

Run: `cd frontend && npx tsc --noEmit && npm run build`

Expected: TypeScript exit code 0; build lista `/custos-fixos` e conclui sem erros.

- [ ] **Step 4: Validar manualmente com backend e frontend locais quando disponíveis**

Abrir `/custos-fixos` e confirmar, nesta ordem:

1. Sem sessão, a URL é substituída por `/login`.
2. Após autenticação, nome e até duas iniciais aparecem no cabeçalho.
3. GET preenche os 13 campos com moeda brasileira.
4. Alterar campos atualiza imediatamente o total mensal.
5. A OAB entra no total como média mensal arredondada igual ao backend.
6. “Restaurar” desfaz edições locais sem nova chamada de rede.
7. “Salvar custos” envia os 13 campos, mostra confirmação e redefine a base de restauração.
8. Uma falha de API mostra toast de erro sem apagar os valores editados.
9. Uma resposta 401 limpa a sessão e redireciona para login.
10. Em largura abaixo de 50rem, aparecem navegação inferior, uma coluna e ações sticky.
11. Início, Honorários, Perfil, marca e botão do perfil não alteram a URL.

- [ ] **Step 5: Corrigir somente defeitos reproduzidos e repetir Steps 2–4**

Para cada defeito encontrado, registrar o comportamento observado, ajustar o
menor arquivo responsável e repetir `git diff --check`, TypeScript, build e o
item manual afetado. Não ampliar o escopo para páginas inexistentes.

- [ ] **Step 6: Commit de correções, se houver**

```bash
git add frontend
git commit -m "fix: address fixed costs verification"
```

Se nenhuma correção for necessária, não criar commit vazio.
