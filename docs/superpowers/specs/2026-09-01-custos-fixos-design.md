# Cadastro de custos fixos — Design

## Objetivo

Criar a página autenticada `/custos-fixos` conforme o protótipo
`frontend/prototypes/custos-fixos.html`, permitindo consultar, editar, salvar e
restaurar os custos fixos do usuário por meio dos endpoints existentes
`GET /api/v1/fixed-costs` e `PATCH /api/v1/fixed-costs`.

## Escopo

- Reproduzir a estrutura visual responsiva do protótipo para desktop e mobile.
- Introduzir uma moldura reutilizável para páginas internas do aplicativo.
- Exibir os dados do usuário autenticado no cabeçalho.
- Carregar e salvar os 13 custos definidos pela API.
- Calcular e atualizar o total mensal estimado durante a edição.
- Restaurar o formulário para os últimos dados retornados com sucesso pela API.
- Manter visíveis, mas sem navegação, os itens destinados a páginas ainda
  inexistentes.

Não fazem parte deste escopo a implementação das páginas Início, Honorários ou
Perfil, menu suspenso de perfil, logout, alteração dos endpoints do backend ou
testes unitários do frontend.

## Arquitetura de componentes

### Página e moldura geral

A rota `frontend/app/custos-fixos/page.tsx` define os metadados da página e
renderiza a funcionalidade de custos fixos.

Um componente geral `AppShell` concentra a marca, sidebar, dica, navegação
responsiva, perfil resumido e a área de conteúdo recebida por `children`. Ele
recebe a identificação do item ativo para destacar “Custos fixos” na sidebar e
“Custos” no mobile. Itens sem página não usam links navegáveis nem alteram a
URL ao serem acionados.

O cabeçalho de conteúdo permanece separado da navegação para permitir que cada
página interna forneça eyebrow, título e descrição próprios. O nome e as
iniciais exibidos no perfil vêm do usuário armazenado na sessão autenticada.

### Formulário

`FixedCostsForm` é um componente cliente responsável pelo ciclo da sessão,
carregamento, estado editável, total mensal, restauração, envio e mensagens de
erro ou sucesso.

As três seções visuais são descritas por dados e renderizadas por componentes
reutilizáveis. `MoneyField` encapsula label, prefixo `R$`, ajuda, atributos de
acessibilidade e entrada monetária, evitando repetição dos 13 campos. O botão
geral já existente é reutilizado e ampliado apenas quando necessário por meio
de classes e conteúdo, sem duplicar sua lógica de loading e disabled.

## Modelo de dados no frontend

O cliente de custos fixos define tipos que espelham a resposta da API:

- `oab_annual_fee` contém `annual_amount_cents` e
  `monthly_amount_cents`.
- Os outros 12 custos contêm `monthly_amount_cents`.
- O envelope contém `user_id` e `costs`.

O estado do formulário usa valores monetários em centavos inteiros. A entrada
e a apresentação usam reais em formato brasileiro. A conversão não aceita
valores negativos nem precisão maior que centavos.

## Fluxo de autenticação e dados

1. Ao montar, o formulário lê e valida a sessão em `localStorage`.
2. Sem sessão válida, limpa qualquer dado inválido e substitui a rota por
   `/login`.
3. Com sessão válida, chama `GET /fixed-costs` usando o token Bearer e mantém os
   controles desabilitados durante o carregamento.
4. A resposta preenche o estado editável e uma cópia de referência chamada
   “últimos dados salvos”.
5. Durante a edição, o resumo soma os 12 custos mensais à média mensal da OAB,
   calculada a partir do valor anual com a mesma regra de arredondamento do
   backend: divisão inteira por 12 e acréscimo de um centavo quando o resto é
   pelo menos 6.
6. “Salvar custos” envia todos os 13 valores em um `PATCH /fixed-costs`.
7. A resposta bem-sucedida atualiza tanto o formulário quanto a referência
   restaurável e apresenta confirmação no toast.
8. “Restaurar” substitui apenas o estado editável pelos últimos dados retornados
   com sucesso, sem nova requisição.

## Contrato do PATCH

O corpo enviado possui a chave `costs`. A anuidade usa
`annual_amount_cents`; os demais campos usam `monthly_amount_cents`. Todos os
campos são enviados para que o formulário represente um cadastro completo,
embora o endpoint também aceite atualizações parciais.

## Estados e tratamento de erros

- Durante o GET, a página preserva sua estrutura e comunica o carregamento,
  mantendo campos e ações indisponíveis.
- Durante o PATCH, os campos e ações ficam indisponíveis e o botão principal
  mostra um rótulo de salvamento.
- Erros válidos da API e falhas de rede são convertidos para mensagens do toast
  já existente.
- Uma resposta `401` limpa a sessão e redireciona para `/login`.
- Uma resposta fora do contrato esperado é tratada como erro inesperado, sem
  aplicar dados parciais ao formulário.
- O botão “Restaurar” fica indisponível enquanto não houver uma resposta válida
  carregada ou enquanto outra operação estiver em andamento.

## Responsividade e acessibilidade

O desktop usa sidebar fixa e conteúdo em duas colunas nos grupos de campos. Em
telas menores, a sidebar é substituída pela navegação fixa inferior, os campos
passam para uma coluna e as ações permanecem acessíveis próximas ao rodapé.

Inputs mantêm labels associados, ajuda ligada por `aria-describedby`, foco
visível e `inputMode="decimal"`. Estados de carregamento são comunicados de
forma textual, botões usam tipos corretos e o item ativo tem `aria-current`.

## Verificação

Por decisão do usuário, esta etapa não adiciona testes unitários nem uma
infraestrutura de testes ao frontend. A implementação será verificada por:

- compilação e checagem de tipos durante `npm run build`;
- inspeção do diff e do contrato JSON contra o handler existente;
- validação funcional do carregamento, edição, total mensal, salvamento,
  restauração, responsividade e redirecionamento por autenticação, na medida em
  que o ambiente local permitir.

## Arquivos previstos

- Criar `frontend/app/custos-fixos/page.tsx`.
- Criar componentes gerais da moldura interna em `frontend/components/app/`.
- Criar componentes e dados do formulário em `frontend/components/fixed-costs/`.
- Criar o cliente e utilitários da funcionalidade em
  `frontend/lib/fixed-costs/`.
- Ajustar `frontend/components/ui/button.tsx` somente se a composição visual
  exigir uma extensão genérica.
- Ajustar `frontend/app/globals.css` com os estilos do protótipo adaptados aos
  componentes React.

O esquema, os campos e a carga do banco não serão alterados; portanto, a seção
de banco de dados do `AGENTS.md` não requer atualização.
