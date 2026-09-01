# Authentication Prototypes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Criar protótipos estáticos, navegáveis e responsivos para login e cadastro do JudCalc, seguindo a referência visual fornecida.

**Architecture:** Dois documentos HTML independentes compartilham uma única folha de estilos em `frontend/prototypes`. Cada documento contém apenas o JavaScript progressivo necessário às interações locais, sem chamadas de rede ou dependência do Next.js.

**Tech Stack:** HTML5 semântico, CSS3 responsivo, JavaScript nativo.

**Spec:** `docs/superpowers/specs/2026-09-01-prototipos-autenticacao-design.md`

## Global Constraints

- Criar apenas arquivos dentro de `frontend/prototypes` durante a implementação.
- Não modificar `frontend/app`, dependências ou configuração do Next.js.
- Não integrar API, persistência, autenticação real ou validação no servidor.
- Usar fundo branco ou cinza muito claro, azul-marinho, azul-cobalto e verde-petróleo; não usar painel escuro, dourado ou tipografia ornamental.
- As páginas devem funcionar abertas diretamente no navegador e não apresentar rolagem horizontal.
- Manter rótulos associados aos campos, foco visível, alvos confortáveis e respeito a `prefers-reduced-motion`.

---

### Task 1: Sistema visual compartilhado

**Files:**
- Create: `frontend/prototypes/styles.css`

**Interfaces:**
- Consumes: nenhuma dependência local.
- Produces: tokens `--navy`, `--blue`, `--teal`, `--surface`, `--border`; classes `.auth-shell`, `.brand-panel`, `.form-panel`, `.auth-card`, `.field`, `.button`, `.benefit-list`, `.password-wrap`, `.password-meter` e utilitário `.sr-only`.

- [ ] **Step 1: Executar a verificação inicial ausente**

Run:

```bash
test -f frontend/prototypes/styles.css
```

Expected: exit code `1`, pois o arquivo ainda não existe.

- [ ] **Step 2: Criar os tokens e a estrutura desktop**

Adicionar reset local, tokens de cor, tipografia de sistema e um layout de duas colunas. A base deve seguir esta interface:

```css
:root {
  --navy: #07132f;
  --blue: #075fd8;
  --blue-soft: #eaf2ff;
  --teal: #168f8b;
  --teal-soft: #e8f7f5;
  --surface: #ffffff;
  --canvas: #f6f8fb;
  --text-muted: #5d6679;
  --border: #dce2eb;
  --danger: #c63d4f;
  --shadow: 0 18px 50px rgba(7, 19, 47, 0.10);
}

*, *::before, *::after { box-sizing: border-box; }
html { color-scheme: light; }
body {
  margin: 0;
  min-width: 20rem;
  color: var(--navy);
  background: var(--canvas);
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.auth-shell {
  display: grid;
  grid-template-columns: minmax(20rem, 0.9fr) minmax(28rem, 1.1fr);
  min-height: 100svh;
}

.auth-card {
  padding: clamp(1.5rem, 3vw, 2.5rem);
  width: min(100%, 31rem);
  border: 1px solid var(--border);
  border-radius: 1.25rem;
  background: var(--surface);
  box-shadow: var(--shadow);
}

.brand-panel, .form-panel { padding: clamp(2rem, 5vw, 5rem); }
.brand-panel { display: flex; flex-direction: column; justify-content: space-between; background: linear-gradient(145deg, #fff 0%, var(--blue-soft) 100%); }
.form-panel { display: flex; align-items: center; justify-content: center; }
.benefit-list { display: grid; gap: 0.75rem; padding: 0; list-style: none; }
.benefit-list li { padding: 1rem; border: 1px solid var(--border); border-radius: 1rem; background: rgba(255,255,255,.82); }
.field { display: grid; gap: 0.5rem; margin-top: 1rem; }
.field input { width: 100%; min-height: 3rem; padding: 0.75rem 0.875rem; border: 1px solid var(--border); border-radius: 0.75rem; background: #fff; color: var(--navy); font: inherit; }
.password-wrap { position: relative; }
.password-wrap input { padding-right: 5rem; }
.password-wrap button { position: absolute; inset: 0 0 0 auto; border: 0; padding: 0 0.875rem; color: var(--blue); background: transparent; font: inherit; font-weight: 700; }
.button { width: 100%; min-height: 3rem; margin-top: 1.25rem; border: 0; border-radius: 0.75rem; color: #fff; background: var(--blue); font: inherit; font-weight: 700; cursor: pointer; }
.button:hover { background: #034fad; }
a { color: var(--blue); font-weight: 650; text-underline-offset: 0.2em; }
input:focus-visible, button:focus-visible, a:focus-visible { outline: 3px solid rgba(7,95,216,.25); outline-offset: 2px; }
button:disabled { cursor: not-allowed; opacity: 0.6; }
.password-meter { height: 0.3rem; overflow: hidden; border-radius: 999px; background: var(--border); }
.password-meter span { display: block; width: 0; height: 100%; background: var(--danger); transition: width .2s ease, background-color .2s ease; }
.password-meter.is-weak span { width: 33%; }
.password-meter.is-medium span { width: 66%; background: #dc8b16; }
.password-meter.is-strong span { width: 100%; background: var(--teal); }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }
```

- [ ] **Step 3: Adicionar adaptação mobile e movimento reduzido**

Usar o seguinte contrato responsivo:

```css
@media (max-width: 760px) {
  .auth-shell { grid-template-columns: 1fr; }
  .brand-panel { min-height: auto; padding: 1.25rem; }
  .brand-copy, .benefit-list { display: none; }
  .form-panel { padding: 1rem; align-items: start; }
  .auth-card { border-radius: 1rem; box-shadow: none; }
}

@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    scroll-behavior: auto !important;
    transition-duration: 0.01ms !important;
  }
}
```

- [ ] **Step 4: Verificar tokens e breakpoints**

Run:

```bash
python -c "from pathlib import Path; s=Path('frontend/prototypes/styles.css').read_text(); assert all(x in s for x in ('--navy:', '--blue:', '--teal:', '@media (max-width: 760px)', 'prefers-reduced-motion', ':focus-visible'))"
```

Expected: exit code `0`.

- [ ] **Step 5: Commit**

```bash
git add frontend/prototypes/styles.css
git commit -m "feat: add prototype visual system"
```

### Task 2: Protótipo de login

**Files:**
- Create: `frontend/prototypes/login.html`

**Interfaces:**
- Consumes: `styles.css` e as classes públicas da Task 1.
- Produces: página com `form#login-form`, `input#email`, `input#password`, `input#remember`, `button#toggle-password` e link `cadastro.html`.

- [ ] **Step 1: Executar a verificação estrutural antes da implementação**

Run:

```bash
python -c "from pathlib import Path; s=Path('frontend/prototypes/login.html').read_text(); assert 'id=\"login-form\"' in s"
```

Expected: falha com `FileNotFoundError`.

- [ ] **Step 2: Criar o documento semântico**

Criar HTML em `pt-BR` com `meta viewport`, referência relativa para `styles.css`, `main.auth-shell`, `aside.brand-panel` e `section.form-panel`. O painel editorial deve exibir a marca JudCalc, a mensagem “Honorários mais seguros. Decisões mais claras.” e três benefícios curtos: precificação orientada, custos sob controle e acesso em qualquer dispositivo.

O formulário deve usar esta interface:

```html
<form id="login-form" novalidate>
  <label for="email">E-mail</label>
  <input id="email" name="email" type="email" autocomplete="email" required>
  <label for="password">Senha</label>
  <div class="password-wrap">
    <input id="password" name="password" type="password" autocomplete="current-password" required>
    <button id="toggle-password" type="button" aria-controls="password" aria-pressed="false">Mostrar</button>
  </div>
  <label><input id="remember" name="remember" type="checkbox"> Lembrar-me</label>
  <a href="#">Esqueci minha senha</a>
  <button class="button" type="submit">Entrar</button>
</form>
```

Adicionar texto para novos usuários com link relativo para `cadastro.html`. Usar SVGs lineares embutidos para o símbolo da marca e os benefícios, sem carregar imagens ou fontes remotas.

- [ ] **Step 3: Adicionar interações progressivas**

Adicionar ao fim do `body` uma mensagem vazia `<p id="form-status" role="status"></p>` e o script:

```html
<script>
  const password = document.querySelector('#password');
  const toggle = document.querySelector('#toggle-password');
  const form = document.querySelector('#login-form');
  const status = document.querySelector('#form-status');

  toggle.addEventListener('click', () => {
    const visible = password.type === 'text';
    password.type = visible ? 'password' : 'text';
    toggle.textContent = visible ? 'Mostrar' : 'Ocultar';
    toggle.setAttribute('aria-pressed', String(!visible));
  });

  form.addEventListener('submit', (event) => {
    event.preventDefault();
    status.textContent = 'Protótipo: nenhuma informação foi enviada.';
  });
</script>
```

- [ ] **Step 4: Validar estrutura, acessibilidade básica e links**

Run:

```bash
python -c "from pathlib import Path; s=Path('frontend/prototypes/login.html').read_text(); required=('lang=\"pt-BR\"','name=\"viewport\"','href=\"styles.css\"','id=\"login-form\"','for=\"email\"','for=\"password\"','autocomplete=\"current-password\"','aria-pressed=\"false\"','href=\"cadastro.html\"','preventDefault()'); assert all(x in s for x in required)"
```

Expected: exit code `0`.

- [ ] **Step 5: Commit**

```bash
git add frontend/prototypes/login.html
git commit -m "feat: add login prototype"
```

### Task 3: Protótipo de cadastro

**Files:**
- Create: `frontend/prototypes/cadastro.html`

**Interfaces:**
- Consumes: `styles.css` e as classes públicas da Task 1.
- Produces: página com `form#signup-form`, `input#full-name`, `input#email`, `input#password`, `input#terms`, `button#toggle-password`, indicador `#password-meter` e link `login.html`.

- [ ] **Step 1: Executar a verificação estrutural antes da implementação**

Run:

```bash
python -c "from pathlib import Path; s=Path('frontend/prototypes/cadastro.html').read_text(); assert 'id=\"signup-form\"' in s"
```

Expected: falha com `FileNotFoundError`.

- [ ] **Step 2: Criar o documento semântico**

Repetir a estrutura visual do login, ajustando o título e o texto editorial para criação da conta. O formulário deve conter nome completo, e-mail, senha e aceite dos termos:

```html
<form id="signup-form" novalidate>
  <label for="full-name">Nome completo</label>
  <input id="full-name" name="full-name" type="text" autocomplete="name" required>
  <label for="email">E-mail</label>
  <input id="email" name="email" type="email" autocomplete="email" required>
  <label for="password">Senha</label>
  <div class="password-wrap">
    <input id="password" name="password" type="password" autocomplete="new-password" minlength="8" required aria-describedby="password-help password-meter-label">
    <button id="toggle-password" type="button" aria-controls="password" aria-pressed="false">Mostrar</button>
  </div>
  <div id="password-meter" class="password-meter" aria-hidden="true"><span></span></div>
  <p id="password-meter-label">Segurança da senha: não informada</p>
  <label><input id="terms" name="terms" type="checkbox" required> Concordo com os Termos de Uso e a Política de Privacidade</label>
  <button class="button" type="submit">Criar conta</button>
</form>
```

Adicionar link relativo para `login.html` e reutilizar SVGs lineares embutidos.

- [ ] **Step 3: Adicionar interações progressivas**

Adicionar `<p id="form-status" role="status"></p>` e este script, que inclui visibilidade, indicador e submissão demonstrativa:

```html
<script>
  const password = document.querySelector('#password');
  const toggle = document.querySelector('#toggle-password');
  const meter = document.querySelector('#password-meter');
  const meterLabel = document.querySelector('#password-meter-label');
  const form = document.querySelector('#signup-form');
  const status = document.querySelector('#form-status');

  toggle.addEventListener('click', () => {
    const visible = password.type === 'text';
    password.type = visible ? 'password' : 'text';
    toggle.textContent = visible ? 'Mostrar' : 'Ocultar';
    toggle.setAttribute('aria-pressed', String(!visible));
  });

  password.addEventListener('input', () => {
    const value = password.value;
    const score = [value.length >= 8, /[A-Za-z]/.test(value), /\d/.test(value), /[^A-Za-z0-9]/.test(value)].filter(Boolean).length;
    const level = !value ? '' : score <= 2 ? 'weak' : score === 3 ? 'medium' : 'strong';
    const label = { '': 'não informada', weak: 'fraca', medium: 'média', strong: 'forte' }[level];
    meter.className = `password-meter${level ? ` is-${level}` : ''}`;
    meterLabel.textContent = `Segurança da senha: ${label}`;
  });

  form.addEventListener('submit', (event) => {
    event.preventDefault();
    status.textContent = 'Protótipo: nenhuma conta foi criada.';
  });
</script>
```

- [ ] **Step 4: Validar estrutura, acessibilidade básica e links**

Run:

```bash
python -c "from pathlib import Path; s=Path('frontend/prototypes/cadastro.html').read_text(); required=('lang=\"pt-BR\"','name=\"viewport\"','href=\"styles.css\"','id=\"signup-form\"','for=\"full-name\"','autocomplete=\"new-password\"','minlength=\"8\"','id=\"password-meter\"','id=\"terms\"','href=\"login.html\"','preventDefault()'); assert all(x in s for x in required)"
```

Expected: exit code `0`.

- [ ] **Step 5: Commit**

```bash
git add frontend/prototypes/cadastro.html
git commit -m "feat: add signup prototype"
```

### Task 4: Validação integrada e inspeção responsiva

**Files:**
- Verify: `frontend/prototypes/styles.css`
- Verify: `frontend/prototypes/login.html`
- Verify: `frontend/prototypes/cadastro.html`

**Interfaces:**
- Consumes: todas as páginas e classes produzidas nas Tasks 1–3.
- Produces: conjunto de protótipos verificado, sem novos arquivos obrigatórios.

- [ ] **Step 1: Confirmar isolamento em relação ao Next.js**

Run:

```bash
git diff --name-only HEAD~3..HEAD
```

Expected: somente `frontend/prototypes/styles.css`, `frontend/prototypes/login.html` e `frontend/prototypes/cadastro.html` entre arquivos de implementação.

- [ ] **Step 2: Validar referências locais e IDs únicos**

Run:

```bash
python -c "from pathlib import Path; from html.parser import HTMLParser; files=list(Path('frontend/prototypes').glob('*.html')); assert len(files)==2; assert all('href=\"styles.css\"' in p.read_text() for p in files); assert Path('frontend/prototypes/styles.css').is_file()"
```

Expected: exit code `0`.

- [ ] **Step 3: Servir os protótipos localmente**

Run:

```bash
python -m http.server 4173 --directory frontend/prototypes
```

Expected: servidor disponível em `http://localhost:4173/login.html` e `http://localhost:4173/cadastro.html`.

- [ ] **Step 4: Inspecionar desktop e mobile**

Abrir as duas páginas em `1440 × 900` e `390 × 844`. Em cada página, confirmar: nenhuma rolagem horizontal; painel dividido somente no desktop; formulário inteiramente utilizável no mobile; foco visível; alternância de senha; mensagem de submissão; links recíprocos; cores e cards coerentes com a referência.

- [ ] **Step 5: Executar a verificação final do repositório**

Run:

```bash
git diff --check && git status --short
```

Expected: nenhum erro de whitespace; nenhuma alteração em `frontend/app`; mudanças preexistentes fora de `frontend/prototypes` permanecem intactas.

- [ ] **Step 6: Commit de ajustes identificados na inspeção**

Se a inspeção exigir ajustes, registrar somente os três arquivos do protótipo:

```bash
git add frontend/prototypes/styles.css frontend/prototypes/login.html frontend/prototypes/cadastro.html
git commit -m "fix: refine responsive auth prototypes"
```
