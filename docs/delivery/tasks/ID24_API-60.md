# ID24_API-60 — Reasoning: единый параметр и учёт reasoning-токенов (E04-14)

Дата: 21.10.2026 · Фаза F2 · Beta.  
Исполнитель: **Маркус Филипп Сергеевич**. Владелец / тех. приёмка: **Ерофеев Илья Владимирович**.  
[Исходная карточка в Сфере](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-60).  
Зависит от: `E04-04` (Стриминг SSE), `ID24_API-24` (Каноническая схема LLM).  
Связано с: `ID24_API-69` (OpenAI), `ID24_API-70` (Anthropic), `ID24_API-74` (Generic), `ID24_API-206` (Контракты usage).

---

## 1. Контекст и архитектурное решение

С появлением рассуждающих моделей (OpenAI o1/o3, Anthropic Claude 3.7 Sonnet Extended Thinking, Google Gemini 2.0 Thinking, DeepSeek-R1) возникла фрагментация протоколов управления процессом «размышлений» модели (chain-of-thought) и биллинга генерируемых reasoning-токенов:
- OpenAI использует строковый параметр `reasoning_effort: "low" | "medium" | "high"`, не возвращает сырой текст рассуждений в Chat Completions и агрегирует токены в `usage.completion_tokens_details.reasoning_tokens`.
- Anthropic требует объект `thinking: { type: "enabled", budget_tokens: N }`, возвращает блоки типа `thinking` в `content[]` и учитывает токены в `usage.output_tokens`.
- Gemini использует `thinkingConfig: { thinkingBudget: N }`.
- DeepSeek возвращает цепочку рассуждений в `choice.message.reasoning_content` и отдельное поле токенов.

В рамках задачи `ID24_API-60` реализован **унифицированный параметр `reasoning`** в канонической схеме NeuroGateway и прозрачный маппинг в апстрим-протоколы.

---

## 2. Реализованный результат

### 2.1. Унифицированная схема запроса (`schema/llm-request.schema.json`)
```json
"reasoning": {
  "type": "object",
  "properties": {
    "effort": {
      "enum": ["none", "minimal", "low", "medium", "high"],
      "type": "string"
    },
    "max_tokens": {
      "type": "integer",
      "minimum": 0
    },
    "exclude": {
      "type": "boolean"
    }
  },
  "additionalProperties": false
}
```

### 2.2. Матрица маппинга провайдеров (`packages/contracts/reasoning/reasoning.go`)

| Канонический параметр | OpenAI (o1, o3-mini) | Anthropic (Claude 3.7 Sonnet) | Google Gemini (Thinking) | DeepSeek (R1) |
|---|---|---|---|---|
| `effort: "minimal"` | `reasoning_effort: "low"` | `budget_tokens: 1024` | `thinkingBudget: 2048` | `reasoning_content: enabled` |
| `effort: "low"` | `reasoning_effort: "low"` | `budget_tokens: 2048` | `thinkingBudget: 2048` | `reasoning_content: enabled` |
| `effort: "medium"` | `reasoning_effort: "medium"` | `budget_tokens: 4096` | `thinkingBudget: 8192` | `reasoning_content: enabled` |
| `effort: "high"` | `reasoning_effort: "high"` | `budget_tokens: 16384` | `thinkingBudget: 16384` | `reasoning_content: enabled` |
| `max_tokens: N` | N/A (фиксируется effort) | `budget_tokens: N` | `thinkingBudget: N` | N/A |
| `exclude: true` | Нативно скрыто | Вырезается блок `thinking` из ответа клиенту | Вырезаются metadata | Вырезается `reasoning_content` |

### 2.3. Формат ответа и стриминг
- В ответе клиенту при наличии рассуждений поле возвращается в `choice.message.reasoning` (и дублируется в `reasoning_content` для обратной совместимости).
- В SSE-стриме чанки рассуждений передаются через `delta.reasoning` или `delta.thinking` до появления первого текстового токена ответа.

### 2.4. Учёт токенов и биллинг
- Поле `usage.reasoning_tokens` изолируется в структуре ответа `usage` и нормализуется в `adapter.NormalizedUsage.ReasoningTokens`.
- Reasoning-токены тарифицируются по тарифу completion-токенов соответствующей модели без скрытых наценок, валидируются на неотрицательность (защита от дрейфа R06).

---

## 3. Критерии приёмки (DoD) и верификация

| Критерий DoD | Факт проверки | Статус |
|---|---|---|
| Единый параметр `reasoning {effort, max_tokens, exclude}` | Валидирован схемой `schema/llm-request.schema.json` | **Выполнено** |
| Маппинг OpenAI (`reasoning_effort`) | Протестирован для o1, o3-mini в `tests/reasoning.test.mjs` | **Выполнено** |
| Маппинг Anthropic (`thinking.budget_tokens`) | Протестирован для Claude 3.7 Sonnet в `tests/reasoning.test.mjs` | **Выполнено** |
| Маппинг Gemini (`thinkingConfig`) | Протестирован для Gemini 2.0 Flash Thinking в `tests/reasoning.test.mjs` | **Выполнено** |
| Маппинг DeepSeek (`reasoning_content`) | Протестирован для DeepSeek-R1 в `tests/reasoning.test.mjs` | **Выполнено** |
| Возврат reasoning в ответе и стриме | Проверена валидация `schema/llm-response.schema.json` | **Выполнено** |
| Учёт `reasoning_tokens` в usage и биллинге | Строгая проверка `NormalizedUsage` и запрет отрицательных токенов | **Выполнено** |

---

## 4. Ссылки на код и артефакты

- Схема запроса: [`schema/llm-request.schema.json`](../../../schema/llm-request.schema.json)
- Схема ответа: [`schema/llm-response.schema.json`](../../../schema/llm-response.schema.json)
- Go-контракт reasoning: [`packages/contracts/reasoning/reasoning.go`](../../../packages/contracts/reasoning/reasoning.go)
- Автотесты: [`tests/reasoning.test.mjs`](../../../tests/reasoning.test.mjs)
