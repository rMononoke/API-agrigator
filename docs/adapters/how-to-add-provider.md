# Руководство разработчика: как добавить нового LLM-провайдера

Данный регламент входит в артефакты межкомандного взаимодействия **Семинара 7** и задачи **`ID24_API-68` (E05-01)**. Он гарантирует, что подключение нового провайдера (OpenAI-совместимого или проприетарного) занимает менее 1 рабочего дня.

---

## 1. Чек-лист подключения провайдера

1. **Юридический и безопасный контур:**
   - [ ] Проверен регламент обработки ПДн (152-ФЗ) и геолокация серверов провайдера.
   - [ ] Ключ доступа размещён в Vault / Secret Manager (не в репозитории).
2. **Реализация интерфейса `ProviderAdapter`:**
   - [ ] Создан пакет `internal/adapters/<provider_name>`.
   - [ ] Реализованы методы: `BuildRequest`, `ParseResponse`, `ParseStreamEvent`, `MapError`, `ExtractUsage`, `Capabilities`.
3. **Контрактные тесты и фикстуры:**
   - [ ] Добавлены golden fixtures (запрос, ответ, SSE-стриминг, ошибки 429/500).
   - [ ] Добавлен тест в `tests/contracts.test.mjs` на корректность маппинга usage и finish_reason.
4. **Регистрация в ядре шлюза:**
   - [ ] Вызов `adapter.Register(NewAdapter(...))` в инициализации шлюза.
   - [ ] Модели провайдера добавлены в каталог `schema/models.json` с указанием capabilities.

---

## 2. Шаблон реализации адаптера

```go
package myprovider

import (
	"context"
	"net/http"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

type MyAdapter struct {}

func (a *MyAdapter) ID() string { return "myprovider" }

func (a *MyAdapter) Capabilities(model string) adapter.Capabilities {
	return adapter.Capabilities{
		SupportsStreaming: true,
		SupportsTools:     true,
		SupportsVision:    false,
		MaxInputTokens:    128000,
		MaxOutputTokens:   4096,
	}
}

func (a *MyAdapter) BuildRequest(ctx context.Context, cfg *adapter.EndpointConfig, req *publicapi.ChatCompletionRequest) (*http.Request, error) {
	// 1. Трансляция канонического DTO в формат API провайдера
	// 2. Создание HTTP-запроса с URL cfg.BaseURL
	// 3. Установка заголовков авторизации из cfg.APIKeySecretRef
	return nil, nil
}

func (a *MyAdapter) ParseResponse(resp *http.Response) (*publicapi.ChatCompletionResponse, error) {
	// Десериализация ответа провайдера в каноническую структуру
	return nil, nil
}

func (a *MyAdapter) ExtractUsage(rawBody []byte) (*adapter.NormalizedUsage, error) {
	// Валидация отсутствия отрицательных токенов и расчет total
	return nil, nil
}

func (a *MyAdapter) MapError(code int, raw []byte) (int, publicapi.Error) {
	// Нормализация кодов 429/500/timeout
	return code, publicapi.Error{}
}
```

---

## 3. Критерии готовности (DoD)

- Любой новый провайдер обязан проходить набор контрактных тестов на совместимость с канонической схемой (`llm-request.schema.json` и `llm-response.schema.json`).
- Проверена изоляция финансовых транзакций: при сбое провайдера не происходит списание баланса.
