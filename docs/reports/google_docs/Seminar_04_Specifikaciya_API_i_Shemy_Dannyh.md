# Семинар 4. Спецификация API, структуры данных, OpenAPI и критерии приёмки

**Дисциплина:** Программная инженерия · **Группа:** ИД24 · **Проект:** NeuroGateway (`ID24_API`)  
**Дата оформления:** 29.10.2026 · **GitHub:** [rMononoke/API-agrigator](https://github.com/rMononoke/API-agrigator)

---

## 1. Контекст и цели семинара
* **Тема в базе знаний Сферы:** FR/NFR, требования к данным, спецификация API, схемы структур данных, проверка типов и форматов, канонические схемы
* **Цель работы:** Сформировать проверяемую спецификацию контрактов данных публичного и внутреннего API, разработать схемы валидации запросов и ответов.

---

## 2. Паспорт проектной команды
| Роль / Зона ответственности | Исполнитель | Email в Сфере |
|---|---|---|
| **Team Lead / Product Owner** | Улитин Максим | `16656@saas.sferaplatform.ru` |
| **Lead Architect / Core Gateway** | Ерофеев Илья | `16625@saas.sferaplatform.ru` |
| **Lead Integrations & Contracts** | Маркус Филипп | `16650@saas.sferaplatform.ru` |
| **Financial & Ledger Engineer** | Деревцов Андрей | `16582@saas.sferaplatform.ru` |

---

## 3. Выполненные задачи команды в трекере «Сфера SaaS»
| Код задачи | Ответственный | Название задачи | Статус в Сфере |
|---|---|---|---|
| **[ID24_API-24](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-24)** | Маркус Филипп | E02-04 · Каноническая схема LLM-запроса и ответа (Unified LLM Schema) | **Выполнено** |
| **[ID24_API-25](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-25)** | Ерофеев Илья | E02-05 · Спецификация публичного API (OpenAPI 3.1) | **Выполнено** |
| **[ID24_API-26](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-26)** | Ерофеев Илья | E02-06 · Модель данных PostgreSQL (ERD) и стратегия миграций | **Выполнено** |
| **[ID24_API-60](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-60)** | Маркус Филипп | E04-14 · Reasoning: единый параметр и учёт reasoning-токенов | **Выполнено** |
| **[ID24_API-62](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-62)** | Маркус Филипп | E04-16 · Prompt caching: cache_control, учёт cached и cache write токенов | **Выполнено** |
| **[ID24_API-93](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-93)** | Деревцов Андрей | E07-01 · Double-entry ledger в PostgreSQL | **Выполнено** |

---

## 4. Персональный вклад участников команды
* **Маркус Филипп (Lead Integrations):** Спроектировал и утвердил каноническую спецификацию Unified LLM Schema (schema/llm-request.schema.json и schema/llm-response.schema.json). Реализовал поддержку передовых возможностей 2026 года: единый параметр reasoning {effort: minimal|low|medium|high, max_tokens, exclude} (ID24_API-60) и директиву prompt caching cache_control: {type: 'ephemeral'} с алгоритмом sticky routing (ID24_API-62).
* **Ерофеев Илья (Lead Architect):** Разработал полную спецификацию публичного HTTP API шлюза в стандарте OpenAPI 3.1 (ID24_API-25), охватывающую /v1/chat/completions, /v1/models, /v1/responses. Спроектировал реляционную модель базы данных PostgreSQL (ERD) с версионированием миграций golang-migrate (ID24_API-26).
* **Деревцов Андрей (Financial & Data):** Спроектировал реляционную схему бухгалтерского учёта с двойной записью (ID24_API-93): таблицы accounts, ledger_entries, transactions, currency_rates с проверкой инварианта сбалансированности проводок.
* **Улитин Максим (Team Lead):** Провёл ревью контрактов данных на соответствие требованиям продуктовых сценариев и критериям совместимости с OpenAI Python/Node.js SDK.

---

## 5. Инженерные артефакты и результаты
* Канонические схемы: schema/llm-request.schema.json и schema/llm-response.schema.json.
* Спецификация OpenAPI 3.1 публичного шлюза: openapi.yaml.
* Схема базы данных PostgreSQL: ERD-диаграмма и SQL-миграции.
* Модули Go: packages/contracts/reasoning/ и packages/contracts/caching/.

---

## 6. Заключение и соответствие Definition of Done
Все задачи семинара выполнены, код и схемы зафиксированы в репозитории, статусы задач переведены в **done** («Выполнено»).
