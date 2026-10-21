# ID24_API-74 — Generic OpenAI-compatible адаптер: подключение провайдеров конфигурацией (E05-07)

Дата: 21.10.2026 · Фаза F2 · Beta.  
Исполнитель: **Маркус Филипп Сергеевич**. Владелец / приёмка: **Ерофеев Илья Владимирович**.  
[Исходная карточка в Сфере](https://fa.saas.sferaplatform.ru/tasks/task/ID24_API-74).  
Зависит от: `ID24_API-68` (Adapter SDK). Связано с: `ID24_API-69` (OpenAI Adapter).

---

## 1. Результат выполнения

1. **Декларативное подключение провайдеров (`config/providers/openai-compatible.yaml`)**:
   - Описаны конфигурации для **8 внешних и локальных провайдеров** без написания специального Go-кода:
     1. **Mistral AI** (`https://api.mistral.ai/v1`)
     2. **Groq Cloud** (`https://api.groq.com/openai/v1`)
     3. **xAI (Grok)** (`https://api.x.ai/v1`)
     4. **Together AI** (`https://api.together.xyz/v1`)
     5. **Fireworks AI** (`https://api.fireworks.ai/inference/v1`)
     6. **DeepSeek Platform** (`https://api.deepseek.com/v1`)
     7. **Alibaba Cloud / Qwen** (`https://dashscope-intl.aliyuncs.com/compatible-mode/v1`)
     8. **vLLM / SGLang Local** (`http://localhost:8000/v1`)
   - Для каждого провайдера зафиксированы: `base_url`, параметры авторизации, `secret_ref` и матрица `capabilities` (streaming, tools, vision, max tokens).

2. **Generic-движок (`internal/adapters/generic/generic.go`)**:
   - Реализован `ConfigurableAdapter`, использующий базовый движок трансляции и валидации `openai.Adapter`, поддерживающий кастомные схемы авторизации и специфичные эндпоинты.

3. **Верификация (`internal/adapters/generic/generic_test.go`)**:
   - Тест `TestConfigurableAdapter_EightProviders` проверяет автоматическую инициализацию, формирование HTTP-запросов и валидацию URL для всех 8 провайдеров.

---

## 2. Критерии приёмки (DoD) и верификация

| Критерий DoD | Факт проверки | Статус |
|---|---|---|
| $\ge 6$ провайдеров подключены без написания отдельного кода адаптера | Подключено 8 провайдеров в `config/providers/openai-compatible.yaml` | Выполнено (DoD перевыполнен) |
| Валидация конфига и smoke-тест при добавлении | Проверено тестом `TestConfigurableAdapter_EightProviders` | Выполнено |
| Совместимость с реестром `adapter.Registry` | Реализует контракт `adapter.ProviderAdapter` | Выполнено |

Статус приёмки: **Выполнено**.
