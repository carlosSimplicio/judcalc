# Design do cálculo de honorários por serviço

## Objetivo

Adicionar um endpoint autenticado que calcule o custo mínimo sustentável e a estimativa técnica de um serviço específico. O cálculo combinará os custos fixos do usuário, sua capacidade mensal de horas faturáveis, as horas estimadas do serviço e fatores predefinidos de complexidade e risco.

A referência da OAB permanecerá separada dos valores calculados. O endpoint apoiará a decisão do advogado sem apresentar a estimativa técnica como validação automática de regularidade ética.

## Escopo

- Criar `POST /api/v1/services/:service_id/fee-calculation` sob o middleware de autenticação existente.
- Identificar o usuário exclusivamente pela sessão autenticada.
- Buscar o serviço pelo identificador da rota e os custos fixos do usuário autenticado.
- Calcular e devolver o custo operacional por hora, o custo mínimo sustentável e a estimativa técnica quando houver custos fixos utilizáveis.
- Devolver a referência da OAB mesmo quando o cálculo econômico não puder ser realizado.
- Não persistir cálculos nem alterar o esquema SQLite.
- Não calcular em dinheiro referências da OAB expressas somente em percentual, pois o MVP ainda não recebe a base econômica sobre a qual o percentual incide.
- Não incluir remuneração desejada, tributos, custos variáveis, margem de segurança ou faixa sugerida nesta entrega.

## Contrato HTTP

### Requisição

```http
POST /api/v1/services/42/fee-calculation
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "estimated_hours": 10,
  "billable_hours_per_month": 80,
  "complexity": "medium",
  "risk": "high"
}
```

`estimated_hours` e `billable_hours_per_month` aceitarão números inteiros ou fracionários maiores que zero. `complexity` aceitará `low`, `medium` ou `high`; `risk` aceitará os mesmos três níveis. Campos desconhecidos serão rejeitados, seguindo o comportamento estrito dos demais corpos JSON da API.

### Resposta com cálculo disponível

```json
{
  "data": {
    "service": {
      "id": 42,
      "area_id": 3,
      "name": "Serviço de exemplo"
    },
    "oab_reference": {
      "amount_cents": 150000,
      "percentage_min": null,
      "percentage_max": null
    },
    "inputs": {
      "estimated_hours": 10,
      "billable_hours_per_month": 80,
      "complexity": "medium",
      "complexity_factor": 1.25,
      "risk": "high",
      "risk_factor": 1.2
    },
    "calculation": {
      "monthly_fixed_costs_cents": 480000,
      "operational_hour_cost_cents": 6000,
      "minimum_sustainable_cost_cents": 60000,
      "technical_estimate_cents": 90000
    },
    "warnings": []
  }
}
```

Os campos monetários calculados serão inteiros em centavos. A referência da OAB preservará os campos nulos existentes no catálogo. Quando o serviço possuir somente percentuais, esses percentuais serão apresentados sem conversão para valor monetário.

### Resposta sem cálculo econômico

Quando o usuário não possuir registro de custos fixos ou quando o total mensal for zero, a requisição continuará bem-sucedida. A API devolverá `200 OK`, preservará a referência da OAB, informará `monthly_fixed_costs_cents` como zero e devolverá como `null` os três resultados econômicos:

```json
{
  "calculation": {
    "monthly_fixed_costs_cents": 0,
    "operational_hour_cost_cents": null,
    "minimum_sustainable_cost_cents": null,
    "technical_estimate_cents": null
  },
  "warnings": [
    {
      "code": "fixed_costs_unavailable",
      "message": "Cadastre custos fixos para calcular o custo sustentável e a estimativa técnica."
    }
  ]
}
```

O objeto acima representa os campos relevantes dentro de `data`; os objetos `service`, `oab_reference` e `inputs` continuarão presentes na resposta completa. O uso de `null` evita que a ausência de dados seja confundida com um honorário calculado de valor zero.

## Regras de cálculo

O total mensal de custos fixos será a soma das categorias mensais cadastradas mais a média mensal da anuidade da OAB. A anuidade seguirá a regra já usada pela API: divisão por 12 e arredondamento para o centavo inteiro mais próximo.

Os fatores iniciais serão constantes no backend:

| Nível | Complexidade | Risco |
| --- | ---: | ---: |
| `low` | 1,00 | 1,00 |
| `medium` | 1,25 | 1,10 |
| `high` | 1,50 | 1,20 |

As fórmulas serão:

```text
custo operacional por hora = custos fixos mensais / horas faturáveis mensais
custo mínimo sustentável = custo operacional por hora * horas estimadas
estimativa técnica = custo mínimo sustentável * fator de complexidade * fator de risco
```

Os fatores nunca reduzirão o resultado abaixo do custo mínimo sustentável. Os cálculos usarão a precisão completa disponível; cada valor monetário exposto será arredondado independentemente para o centavo inteiro mais próximo, sem reutilizar um valor intermediário já arredondado.

## Componentes e fluxo

Um calculador puro no domínio receberá os custos fixos consolidados e as entradas validadas, aplicará os fatores e produzirá o resultado sem depender de HTTP ou SQLite. Essa separação manterá as fórmulas testáveis e permitirá calibrar os fatores posteriormente sem acoplar a regra ao handler.

O repositório de serviços ganhará uma operação de busca por ID. O handler executará o fluxo:

1. Obter o usuário autenticado.
2. Validar o ID do serviço e o corpo da requisição.
3. Buscar o serviço; interromper com `404` se ele não existir.
4. Buscar os custos fixos do usuário.
5. Invocar o calculador quando o total mensal for maior que zero.
6. Montar a resposta com as três referências separadas e eventuais avisos.

O endpoint reutilizará as dependências de serviços e custos fixos já entregues ao roteador; não será criado um novo repositório nem haverá persistência do resultado.

## Tratamento de erros

- Token ausente ou inválido: `401 unauthorized`, conforme o middleware existente.
- `service_id` inválido: `400 invalid_service_id`.
- Corpo malformado, campos desconhecidos, horas não positivas ou níveis desconhecidos: `400 invalid_body`.
- Serviço inexistente: `404 service_not_found`.
- Falha ao consultar serviços ou custos fixos: `500 internal_error`, sem expor detalhes internos.
- Custos fixos ausentes ou zerados: `200 OK` com `fixed_costs_unavailable`, referência da OAB e resultados econômicos nulos.

## Arquivos e documentação

A implementação deverá concentrar as mudanças em:

- novo modelo e calculador em `backend/internal/domain`;
- interface de repositório e busca por ID no repositório SQLite de serviços;
- novo handler HTTP para o cálculo;
- registro da rota e injeção das dependências existentes;
- testes de domínio, armazenamento e HTTP;
- documentação do endpoint no `README.md`.

Como o esquema, os campos persistidos, as constraints, os relacionamentos, a carga e a busca FTS não mudarão, a seção de banco de dados do `AGENTS.md` não precisará ser alterada.

## Validação

- Testar cada nível de complexidade e risco, inclusive as combinações cumulativas.
- Testar horas fracionárias e arredondamento dos resultados em centavos.
- Confirmar que o cálculo usa todos os custos mensais e a média mensal da anuidade da OAB.
- Confirmar que usuário sem custos recebe `200`, referência da OAB, resultados nulos e o aviso esperado.
- Confirmar que referências com valor, percentual ou ambos preservam seus campos sem conversão indevida.
- Testar `service_id` inválido, serviço inexistente, corpo malformado, horas não positivas e níveis inválidos.
- Testar falhas dos dois repositórios e garantir que detalhes internos não vazem.
- Executar `go test ./...` no backend após reconciliar os testes de autenticação que já estão em desenvolvimento no diretório de trabalho.
