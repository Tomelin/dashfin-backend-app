# API de Planos de Gastos

## Visão Geral

Esta API permite aos usuários criar, visualizar e atualizar seus planos de gastos mensais. Um plano de gastos consiste na renda mensal do usuário e uma lista de orçamentos por categoria.

## Caminho Base

Todas as rotas para esta API estão sob o caminho base: `/api/finance/spending-plan`

## Autenticação

O `UserID` para identificar o usuário é derivado do token de autenticação fornecido no cabeçalho da requisição. Todas as requisições para esta API devem ser autenticadas.

---

## Endpoints

### Obter Plano de Gastos

*   **Método**: `GET`
*   **Endpoint**: `/api/finance/spending-plan`
*   **Descrição**: Obtém o plano de gastos salvo pelo usuário.

*   **Resposta de Sucesso (200 OK)**:
    *   **Content-Type**: `application/json`
    *   **Corpo**: Um objeto JSON representando o plano de gastos do usuário.
        ```json
        {
            "monthlyIncome": 5000.00,
            "categoryBudgets": [
                {
                    "category": "moradia",
                    "amount": 1500.00,
                    "percentage": 30.00
                },
                {
                    "category": "alimentacao",
                    "amount": 500.00,
                    "percentage": 10.00
                }
            ],
            "userID": "user-id-example",
            "createdAt": "2023-10-27T10:00:00Z",
            "updatedAt": "2023-10-27T10:00:00Z"
        }
        ```

*   **Respostas de Erro**:
    *   `401 Unauthorized`: Falha na autenticação (e.g., token inválido ou ausente).
        ```json
        {
            "error": "Unauthorized"
        }
        ```
    *   `404 Not Found`: Nenhum plano de gastos encontrado para o usuário.
        ```json
        {
            "error": "Spending plan not found for this user"
        }
        ```
    *   `500 Internal Server Error`: Erro inesperado no servidor.
        ```json
        {
            "error": "Failed to retrieve spending plan",
            "details": "descrição do erro interno"
        }
        ```

---

### Salvar/Atualizar Plano de Gastos

*   **Método**: `PUT`
*   **Endpoint**: `/api/finance/spending-plan`
*   **Descrição**: Salva (cria um novo) ou atualiza o plano de gastos existente do usuário. O sistema determina se é uma criação ou atualização baseado na existência de um plano para o `UserID`.

*   **Corpo da Requisição**:
    *   **Content-Type**: `application/json`
    *   **Esquema**: Um objeto JSON contendo os campos `monthlyIncome` e `categoryBudgets`.
        ```json
        {
            "monthlyIncome": 5500.00,
            "categoryBudgets": [
                {
                    "category": "moradia",
                    "amount": 1600.00,
                    "percentage": 29.09
                },
                {
                    "category": "transporte",
                    "amount": 300.00,
                    "percentage": 5.45
                }
            ]
        }
        ```

*   **Resposta de Sucesso (200 OK)**:
    *   **Content-Type**: `application/json`
    *   **Corpo**: Um objeto JSON representando o plano de gastos salvo ou atualizado, incluindo os campos gerenciados pelo backend (`userID`, `createdAt`, `updatedAt`).
        ```json
        {
            "monthlyIncome": 5500.00,
            "categoryBudgets": [
                {
                    "category": "moradia",
                    "amount": 1600.00,
                    "percentage": 29.09
                },
                {
                    "category": "transporte",
                    "amount": 300.00,
                    "percentage": 5.45
                }
            ],
            "userID": "user-id-example",
            "createdAt": "2023-10-27T10:00:00Z",
            "updatedAt": "2023-10-28T14:30:00Z"
        }
        ```

*   **Respostas de Erro**:
    *   `400 Bad Request`: Se o corpo da requisição for inválido ou malformado.
        ```json
        {
            "error": "Invalid request body",
            "details": "descrição do erro de validação"
        }
        ```
    *   `401 Unauthorized`: Falha na autenticação.
        ```json
        {
            "error": "Unauthorized"
        }
        ```
    *   `500 Internal Server Error`: Erro inesperado no servidor ao tentar salvar o plano.
        ```json
        {
            "error": "Failed to save spending plan",
            "details": "descrição do erro interno"
        }
        ```

---

## Esquemas de Dados

### `SpendingPlan`
Representa o plano de gastos mensal de um usuário.

| Campo             | Tipo                                      | Descrição                                                                 | Restrições                     | Atributos   |
|-------------------|-------------------------------------------|---------------------------------------------------------------------------|--------------------------------|-------------|
| `monthlyIncome`   | número (float)                            | A renda mensal total do usuário.                                          | Mínimo: 0                      | Obrigatório |
| `categoryBudgets` | array de [`CategoryBudget`](#categorybudget) | Uma lista dos orçamentos definidos para cada categoria de gasto.        |                                | Obrigatório |
| `userID`          | string                                    | O identificador único do usuário.                                         |                                | Read-Only   |
| `createdAt`       | string (datetime ISO8601)                 | O timestamp de quando o plano de gastos foi criado.                       | Formato: `YYYY-MM-DDTHH:MM:SSZ` | Read-Only   |
| `updatedAt`       | string (datetime ISO8601)                 | O timestamp da última atualização do plano de gastos.                     | Formato: `YYYY-MM-DDTHH:MM:SSZ` | Read-Only   |

### `CategoryBudget`
Representa o orçamento para uma categoria específica dentro do plano de gastos.

| Campo        | Tipo           | Descrição                                                              | Restrições              | Atributos   |
|--------------|----------------|------------------------------------------------------------------------|-------------------------|-------------|
| `category`   | string         | Um identificador único para a categoria de gasto (e.g., 'moradia', 'alimentacao', 'transporte'). |                         | Obrigatório |
| `amount`     | número (float) | O valor monetário orçado para esta categoria.                          | Mínimo: 0               | Obrigatório |
| `percentage` | número (float) | A porcentagem da renda mensal que este orçamento representa.             | Mínimo: 0, Máximo: 100  | Obrigatório |

---

## Comportamento do Cache

Para melhorar o desempenho e reduzir a carga no banco de dados, os dados do plano de gastos do usuário são armazenados em cache:

*   **Leitura (GET):** Ao solicitar o plano de gastos, o sistema primeiro tenta recuperá-lo do cache. Se encontrado, os dados em cache são retornados imediatamente. Caso contrário, os dados são buscados no banco de dados principal e armazenados em cache para solicitações futuras.
*   **Vida útil do Cache (TTL):** Os planos de gastos dos usuários são mantidos em cache por 10 minutos.
*   **Escrita (PUT):** Quando um plano de gastos é salvo ou atualizado (via PUT), o cache correspondente àquele usuário é invalidado (removido). Isso garante que a próxima solicitação GET para este usuário buscará os dados mais recentes do banco de dados e atualizará o cache.

---
