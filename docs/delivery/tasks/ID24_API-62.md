# ID24_API-62 — Prompt caching: cache_control, учёт cached и cache write токенов (E04-16)

Дата: 21.10.2026 · Фаза F2 · Beta.  
Исполнитель: **Маркус Филипп Сергеевич**. Владелец / тех. приёмка: **Ерофеев Илья Владимирович**. Соисполнитель: **Андрей Деревцов** (биллинг/кэш).  
[Исходная карточка в Сфере](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-62).  
Зависит от: `E04-04` (Стриминг SSE), `E07-02` (Кэширование ответов), `ID24_API-24` (Каноническая схема).  
Связано с: `ID24_API-69` (OpenAI), `ID24_API-70` (Anthropic), `ID24_API-206` (Контракты usage).

---

## 1. Контекст и цели задачи

Кэширование промптов (Prompt Caching) позволяет разработчикам многократно переиспользовать статические системные инструкции, объёмные документы, кодовые базы и примеры с существенной экономией:
- **Anthropic Claude 3.5 Sonnet / Haiku:** скидка 90% на повторные чтения (`$0.30` за 1M токенов вместо `$3.00`), наценка 25% на первичное сохранение кэша (`$3.75` за 1M токенов, TTL 5 минут).
- **OpenAI (GPT-4o, GPT-4o mini, o1):** автоматическое кэширование префиксов $\ge 1024$ токенов со скидкой 50% на чтение (`$1.25` за 1M вместо `$2.50`), без доплаты за запись.
- **DeepSeek V3 / R1:** нативное кэширование со скидкой 90% (`$0.014` за 1M токенов).

Задача шлюза NeuroGateway — обеспечить прозрачное прокидывание `cache_control`, детерминированную маршрутизацию (*sticky routing*) для максимального попадания в кэш провайдера, корректный учёт токенов в биллинге и наглядное отражение экономии для клиента.

---

## 2. Реализованный результат

### 2.1. Поддержка `cache_control` в канонической схеме (`schema/llm-request.schema.json`)
В структуры сообщений диалога добавлена директива кэширования:
```json
"cache_control": {
  "type": "object",
  "properties": {
    "type": {
      "enum": ["ephemeral"],
      "type": "string"
    }
  },
  "required": ["type"],
  "additionalProperties": false
}
```
Адаптер Anthropic транслирует этот маркер в контрольную точку кэша Messages API, адаптер OpenAI передаёт контекст с оптимизацией префикса.

### 2.2. Sticky Routing по префиксу промпта (`packages/contracts/caching/caching.go`)
Для предотвращения промахов мимо кэша при балансировке между несколькими кластерами или учётными записями шлюз вычисляет 16-значный детерминированный хэш префикса (`ComputePrefixHash`):
- Учитываются системные промпты и первые сообщения диалога ($\ge 1024$ токенов).
- Запросы с одинаковым хэшем направляются на один и тот же целевой upstream-воркер (консистентное хэширование).

### 2.3. Биллинг и отражение экономии
В модуле `caching.CalculateSavings` реализована точная модель расчёта затрат с учётом дифференцированных тарифов:
- **Базовая стоимость** (без кэширования):
  $$\text{BaseCost} = \frac{\text{PromptTokens}}{10^6} \times P_{\text{base\_in}} + \frac{\text{CompletionTokens}}{10^6} \times P_{\text{base\_out}}$$
- **Фактическая стоимость** (с кэшированием):
  $$\text{ActualCost} = \frac{\text{Prompt} - \text{Cached}}{10^6} \times P_{\text{base\_in}} + \frac{\text{Cached}}{10^6} \times P_{\text{cache\_read}} + \frac{\text{CacheWrite}}{10^6} \times P_{\text{cache\_write}} + \text{Cost}_{\text{out}}$$
- **Экономия** передаётся в ответе через `usage.prompt_tokens_details.cached_tokens`, `usage.cache_write_tokens` и сохраняется в биллинговой транзакции ЛК.

---

## 3. Критерии приёмки (DoD) и верификация

| Критерий DoD | Факт проверки | Статус |
|---|---|---|
| Прокидывание `cache_control` (Anthropic, Gemini) | Валидируется схемой `schema/llm-request.schema.json`, проверено в `tests/caching.test.mjs` | **Выполнено** |
| Sticky routing по хэшу префикса | Детерминированный `ComputePrefixHash`, покрыт автотестом `tests/caching.test.mjs` | **Выполнено** |
| Извлечение `cached_tokens` и `cache_write_tokens` | Поддержано в `NormalizedUsage`, `ExtractUsage` адаптеров OpenAI и Anthropic | **Выполнено** |
| Повторный запрос с длинным системным промптом тарифицируется по цене cache read | Подтверждено на тарифах Claude 3.5 Sonnet (скидка 90% на чтение) и OpenAI GPT-4o (скидка 50%) | **Выполнено** |
| Защита от дрейфа (R06) | Запрет отрицательных значений `cached_tokens` и `cached_tokens > prompt_tokens` | **Выполнено** |

---

## 4. Ссылки на код и артефакты

- Контракт кэширования и биллинга: [`packages/contracts/caching/caching.go`](../../../packages/contracts/caching/caching.go)
- Схема запроса: [`schema/llm-request.schema.json`](../../../schema/llm-request.schema.json)
- Схема ответа: [`schema/llm-response.schema.json`](../../../schema/llm-response.schema.json)
- Автотесты: [`tests/caching.test.mjs`](../../../tests/caching.test.mjs)
