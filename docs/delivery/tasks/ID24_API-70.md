# ID24_API-70 — Адаптер Anthropic: Messages API, thinking, cache_control, tools (E05-03)

Дата: 21.10.2026 · Фаза F1 · MVP.  
Исполнитель: **Маркус Филипп Сергеевич**. Владелец / приёмка: **Ерофеев Илья Владимирович**.  
[Исходная карточка в Сфере](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-70).  
Зависит от: `ID24_API-68` (Adapter SDK). Связано с: `ID24_API-24` (Unified Schema), `ID24_API-206` (R05/R06/R07).

---

## 1. Результат выполнения

1. **Реализация адаптера (`internal/adapters/anthropic/anthropic.go`)**:
   - **Трансляция структуры диалога**: Выделение `system` и `developer` сообщений из общего массива messages в top-level поле `system` Messages API. Нормализация чередования ролей `user` и `assistant`.
   - **Поддержка инструментов (Tools)**: Трансляция канонического `tools` в `input_schema` Anthropic, обработка `tool_use` и ответов `tool_result`.
   - **Prompt Caching и Extended Thinking**: Поддержка передачи `thinking` (budget_tokens) и извлечения `cache_read_input_tokens`, `cache_creation_input_tokens`.
   - **Нормализация ответа и стриминга**: Трансляция `stop_reason` (`end_turn` -> `stop`, `max_tokens` -> `length`, `tool_use` -> `tool_calls`). Парсинг потоковых SSE-событий (`message_start`, `content_block_delta`, `message_stop`) в канонические чанки `chat.completion.chunk`.
   - **Извлечение usage и защита от искажения (R06)**: Точное сведение `input_tokens` $\rightarrow$ `prompt_tokens`, `output_tokens` $\rightarrow$ `completion_tokens` с валидацией неотрицательности.
   - **Аутентификация**: Использование обязательных заголовков `x-api-key` и `anthropic-version: 2023-06-01`.

2. **Тестирование (`internal/adapters/anthropic/anthropic_test.go`)**:
   - `TestAnthropicAdapter_BuildRequest`: проверка трансляции system-промпта, URL `/v1/messages` и заголовков.
   - `TestAnthropicAdapter_ExtractUsage`: проверка извлечения входных, выходных и кэшированных токенов.

---

## 2. Критерии приёмки (DoD) и верификация

| Критерий DoD | Факт проверки | Статус |
|---|---|---|
| Трансляция формата OpenAI $\rightarrow$ Anthropic Messages | Реализовано в `internal/adapters/anthropic/anthropic.go` | Выполнено |
| Извлечение точного usage (input, output, cache_read) | Проверено тестом `TestAnthropicAdapter_ExtractUsage` | Выполнено |
| Трансляция SSE-событий в канонические chunks | Реализовано в методе `ParseStreamEvent` | Выполнено |
| Контрактные тесты на совместимость с Unified Schema | Проверено `tests/contracts.test.mjs` (16 passed) | Выполнено |

Статус приёмки: **Выполнено**.
