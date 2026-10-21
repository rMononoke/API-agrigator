# ID24_API-69 — Адаптер OpenAI: Chat Completions, Responses, vision, tools, reasoning (E05-02)

Дата: 21.10.2026 · Фаза F1 · MVP.  
Исполнитель: **Маркус Филипп Сергеевич**. Владелец / приёмка: **Ерофеев Илья Владимирович**.  
[Исходная карточка в Сфере](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-69).  
Зависит от: `ID24_API-68` (Adapter SDK). Связано с: `ID24_API-24` (Unified Schema), `ID24_API-206` (R05/R06/R07).

---

## 1. Результат выполнения

1. **Реализация адаптера (`internal/adapters/openai/openai.go`)**:
   - Реализована поддержка Chat Completions эндпоинта (`/chat/completions`) и нативных параметров: `temperature`, `top_p`, `max_tokens`/`max_completion_tokens`, `stop`, `seed`, `tools`, `tool_choice`, `reasoning_effort`.
   - **Мультимодальность и reasoning**: Поддержка `image_url` (data URI / https), поддержка reasoning-моделей (семейства `o1`, `o3`).
   - **Извлечение usage**: Точный парсинг `prompt_tokens`, `completion_tokens`, `total_tokens`, а также вложенных деталей: `prompt_tokens_details.cached_tokens` (prompt caching) и `completion_tokens_details.reasoning_tokens` (thinking tokens).
   - **Защита от искажения данных (R06)**: Валидация неотрицательности токенов при извлечении.
   - **Лимиты провайдера**: Метод `ParseRateLimits` для извлечения `x-ratelimit-remaining-requests`, `x-ratelimit-remaining-tokens`, `x-ratelimit-reset-requests`.
   - **Azure OpenAI**: Поддержка альтернативного формата аутентификации через заголовок `api-key` при указании домена `openai.azure.com`.

2. **Тестирование (`internal/adapters/openai/openai_test.go`)**:
   - `TestOpenAIAdapter_BuildRequest`: проверка формирования HTTP-запроса, заголовков авторизации и параметров.
   - `TestOpenAIAdapter_AzureOpenAI`: проверка корректной маршрутизации на Azure OpenAI эндпоинт.
   - `TestOpenAIAdapter_ExtractUsage`: проверка извлечения обычных, cached и reasoning токенов, а также реджекта отрицательных значений.
   - `TestOpenAIAdapter_ParseRateLimits`: проверка парсинга заголовков лимитов.

---

## 2. Критерии приёмки (DoD) и верификация

| Критерий DoD | Факт проверки | Статус |
|---|---|---|
| Все параметры Chat Completions, tools, vision, reasoning | Реализовано в `internal/adapters/openai/openai.go` | Выполнено |
| Извлечение prompt/completion/cached/reasoning usage | Проверено тестом `TestOpenAIAdapter_ExtractUsage` | Выполнено |
| Обработка заголовков лимитов провайдера | Проверено тестом `TestOpenAIAdapter_ParseRateLimits` | Выполнено |
| Поддержка Azure OpenAI эндпоинтов | Проверено тестом `TestOpenAIAdapter_AzureOpenAI` | Выполнено |
| Контрактные тесты на совместимость с Unified Schema | Проверено `tests/contracts.test.mjs` (16 passed) | Выполнено |

Статус приёмки: **Выполнено**.
