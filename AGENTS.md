# Instruções para agentes

- Seja conciso e direto em suas respostas.

## Banco de dados

- Sempre que o esquema, os campos, as constraints, os relacionamentos, a carga
  ou a busca FTS forem modificados, atualize também esta seção do `AGENTS.md` na
  mesma alteração.
- O MVP usa SQLite com FTS5. O esquema está em `python/database/schema.sql`.
- `python/data/oab-sp.json` é a fonte de verdade. Execute
  `python -m python.scripts.init_database` para criar ou sincronizar
  `backend/data/app.db`; o arquivo do banco é regenerável e não
  deve ser versionado.
- A carga é transacional, substitui integralmente apenas o catálogo de áreas e
  serviços, preserva usuários, tokens e custos fixos válidos e habilita
  `PRAGMA foreign_keys = ON`.
- O projeto ainda não possui migrações. Após alterações incompatíveis no
  esquema, apague `backend/data/app.db` e execute novamente o inicializador.

### Tabela `areas`

- `id`: chave primária inteira.
- `name`: nome obrigatório e único da área jurídica.

### Tabela `services`

- `id`: chave primária inteira.
- `area_id`: chave estrangeira obrigatória para `areas.id`, com exclusão em
  cascata.
- `name`: descrição obrigatória do serviço.
- `amount_cents`: valor monetário mínimo em centavos; aceita `NULL` quando não
  informado e não aceita valores negativos.
- `percentage_min` e `percentage_max`: limites percentuais reais; ambos devem ser
  `NULL` ou ambos preenchidos, entre 0 e 100, com o mínimo menor ou igual ao
  máximo.
- `area_id` possui índice para relacionamentos e filtros.
- Uma variante é única pela combinação de área, nome, valor e percentuais.
  Isso preserva o serviço homônimo `Atuação somente a partir da fase recursal`,
  que aparece na mesma área previdenciária com dois valores distintos.

### Busca textual

- `areas_fts` indexa `areas.name` e `services_fts` indexa `services.name` como
  tabelas virtuais FTS5 de conteúdo externo.
- O tokenizer `unicode61 remove_diacritics 2` permite buscas sem acentos e por
  prefixos.
- Triggers de inserção, atualização e exclusão mantêm os índices FTS
  sincronizados; não os atualize diretamente.

### Tabela `users`

- `id`: chave primária inteira autoincrementável; identifica o usuário na API.
- `email`: email obrigatório, normalizado em minúsculas, único sem distinção de
  maiúsculas e minúsculas.
- `name`: nome obrigatório, armazenado sem espaços nas extremidades.
- `password_hash`: hash bcrypt obrigatório; senhas nunca são armazenadas em
  texto puro.
- `created_at`: instante de criação em segundos Unix, obrigatório e não
  negativo.

### Tabela `auth_tokens`

- `id`: chave primária inteira.
- `user_id`: chave estrangeira obrigatória para `users.id`, com exclusão em
  cascata e índice para relacionamento.
- `token_hash`: SHA-256 hexadecimal obrigatório e único do token opaco; o token
  utilizável nunca é armazenado.
- `created_at` e `expires_at`: instantes em segundos Unix; a expiração deve ser
  posterior à criação e possui índice para manutenção.
- Cada sign-in ou cadastro cria um token independente com validade de 30 dias.

### Tabela `user_fixed_costs`

- `user_id`: identificador inteiro e chave primária, relacionado a `users.id`
  com exclusão em cascata.
- `oab_annual_fee_cents`: anuidade da OAB em centavos. A média mensal é
  calculada pela API e não é armazenada.
- `digital_certificate_cents`, `accountant_cents`, `legal_software_cents`,
  `internet_cents`, `phone_cents`, `recurring_transport_cents`,
  `coworking_or_office_rent_cents`,
  `professional_domain_website_email_cents`, `marketing_cents`,
  `office_supplies_cents`, `equipment_and_depreciation_cents` e
  `other_costs_cents`: custos mensais em centavos.
- Todos os valores monetários são obrigatórios, têm padrão zero e não aceitam
  números negativos.
- Há um único registro por usuário. A API realiza upsert e preserva os campos
  omitidos em atualizações parciais.
