# Design dos protótipos de autenticação

## Objetivo

Criar dois protótipos estáticos e navegáveis para validar o design das telas de autenticação do JudCalc: login e cadastro. Os arquivos serão independentes da aplicação Next.js e funcionarão ao serem abertos diretamente no navegador.

## Escopo

- Adicionar `frontend/prototypes/login.html`.
- Adicionar `frontend/prototypes/cadastro.html`.
- Adicionar `frontend/prototypes/styles.css`, compartilhado pelas duas páginas.
- Não modificar arquivos em `frontend/app`, dependências ou configuração do Next.js.
- Não integrar API, persistência, autenticação real ou validação no servidor.

## Direção visual

A interface será premium e minimalista, com paleta de grafite, marfim e dourado discreto. A composição usará bastante espaço livre, contrastes contidos, bordas finas e tipografia elegante sem comprometer a legibilidade.

Em desktop, cada página terá um painel dividido: uma área editorial com a marca e uma mensagem curta de valor, e uma área funcional com o formulário. Em telas menores, a área editorial será reduzida ao cabeçalho da marca e o formulário ocupará o fluxo principal, sem rolagem horizontal.

## Conteúdo e componentes

As duas páginas terão marca JudCalc, título, texto de apoio, campos com rótulos visíveis, chamada principal e link para alternar entre login e cadastro.

### Login

- E-mail.
- Senha, com controle visual para exibir ou ocultar o conteúdo.
- Opção “Lembrar-me”.
- Link “Esqueci minha senha”.
- Botão “Entrar”.
- Link para `cadastro.html`.

### Cadastro

- Nome completo.
- E-mail.
- Senha, com orientação compacta e indicador visual de segurança.
- Aceite dos termos e da política de privacidade.
- Botão “Criar conta”.
- Link para `login.html`.

## Comportamento do protótipo

Os formulários serão semânticos, mas não enviarão dados. O JavaScript embutido será limitado a interações demonstrativas locais, como exibir ou ocultar senha e atualizar o indicador visual da senha. Submissões serão neutralizadas para evitar navegação acidental.

Os estados de foco, hover e preenchimento serão visíveis. Campos usarão tipos HTML adequados, autocomplete e atributos básicos de obrigatoriedade para permitir avaliação da experiência nativa do navegador.

## Responsividade e acessibilidade

- Layout fluido com ponto de adaptação para telas pequenas, sem depender de uma largura fixa de dispositivo.
- Alvos interativos confortáveis em dispositivos móveis.
- Contraste legível, foco perceptível e rótulos associados aos controles.
- Uso correto de títulos, regiões principais e textos auxiliares.
- Respeito à preferência por movimento reduzido.

## Validação

- Confirmar que somente arquivos de `frontend/prototypes` foram adicionados durante a implementação.
- Verificar a existência e a estrutura semântica dos dois documentos.
- Conferir que os links entre login e cadastro apontam para os arquivos corretos.
- Validar que não há rolagem horizontal nas larguras representativas de mobile e desktop.
- Fazer inspeção visual das duas páginas nas duas faixas de resolução.
