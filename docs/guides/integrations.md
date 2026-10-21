# Руководство по интеграциям с NeuroGateway (16 инструментов)

В данном документе собраны проверенные конфигурации и примеры кода для подключения популярных SDK, фреймворков, сред разработки (IDE) и No-Code платформ к шлюзу **NeuroGateway**.

**Базовые параметры платформы:**
- **Base URL (OpenAI-совместимый):** `https://api.neurogateway.io/v1` (локально: `http://localhost:8000/v1`)
- **Base URL (Anthropic-совместимый):** `https://api.neurogateway.io/v1` (локально: `http://localhost:8000/v1`)
- **API Key:** `ng-live-...` (создается в Личном кабинете)

---

## 1. Официальные SDK и библиотеки

### 1.1. Python: OpenAI SDK
```python
from openai import OpenAI

client = OpenAI(
    base_url="https://api.neurogateway.io/v1",
    api_key="ng-live-your-token-here"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Привет, NeuroGateway!"}]
)
print(response.choices[0].message.content)
```

### 1.2. TypeScript / Node.js: OpenAI SDK
```typescript
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "https://api.neurogateway.io/v1",
  apiKey: process.env.NEUROGATEWAY_API_KEY || "ng-live-your-token-here",
});

async function main() {
  const completion = await client.chat.completions.create({
    model: "claude-3-5-sonnet",
    messages: [{ role: "user", content: "Привет из TypeScript!" }],
  });
  console.log(completion.choices[0].message.content);
}
main();
```

### 1.3. Python: Anthropic SDK
```python
import anthropic

client = anthropic.Anthropic(
    base_url="https://api.neurogateway.io/v1",
    api_key="ng-live-your-token-here"
)

message = client.messages.create(
    model="claude-3-5-sonnet",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Привет, Claude через шлюз!"}]
)
print(message.content[0].text)
```

### 1.4. LangChain (Python)
```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gpt-4o-mini",
    openai_api_base="https://api.neurogateway.io/v1",
    openai_api_key="ng-live-your-token-here",
    temperature=0.7
)

res = llm.invoke("Объясни понятие Data Mesh в двух предложениях.")
print(res.content)
```

### 1.5. LlamaIndex (Python)
```python
from llama_index.llms.openai import OpenAI
from llama_index.core.llms import ChatMessage

llm = OpenAI(
    model="gpt-4o",
    api_base="https://api.neurogateway.io/v1",
    api_key="ng-live-your-token-here"
)

response = llm.chat([ChatMessage(role="user", content="Привет, LlamaIndex!")])
print(response.message.content)
```

### 1.6. Vercel AI SDK (Next.js / Node.js)
```typescript
import { createOpenAI } from "@ai-sdk/openai";
import { generateText } from "ai";

const gateway = createOpenAI({
  baseURL: "https://api.neurogateway.io/v1",
  apiKey: process.env.NEUROGATEWAY_API_KEY,
});

export async function run() {
  const { text } = await generateText({
    model: gateway("gpt-4o"),
    prompt: "Напиши SQL-запрос для подсчёта MAU.",
  });
  return text;
}
```

### 1.7. LiteLLM Proxy / Library
```python
import litellm

response = litellm.completion(
    model="openai/gpt-4o",
    api_base="https://api.neurogateway.io/v1",
    api_key="ng-live-your-token-here",
    messages=[{"role": "user", "content": "Hello LiteLLM!"}]
)
print(response.choices[0].message.content)
```

---

## 2. Инструменты разработчика (Developer IDE / CLI)

### 2.1. Cursor IDE
1. Перейдите в `Settings` $\rightarrow$ `Models` $\rightarrow$ `OpenAI API Key`.
2. Включите галочку `Override OpenAI Base URL`.
3. Укажите: `https://api.neurogateway.io/v1`.
4. Введите ваш API-ключ `ng-live-...`.
5. Добавьте любые доступные модели: `gpt-4o`, `claude-3-5-sonnet`, `o1`.

### 2.2. Claude Code CLI
```bash
export ANTHROPIC_BASE_URL="https://api.neurogateway.io/v1"
export ANTHROPIC_API_KEY="ng-live-your-token-here"
claude
```

### 2.3. Cline (VS Code Extension)
1. Откройте панель **Cline** в VS Code $\rightarrow$ `Settings` (иконка шестерёнки).
2. Выберите провайдера: `OpenAI Compatible`.
3. **Base URL:** `https://api.neurogateway.io/v1`.
4. **API Key:** `ng-live-your-token-here`.
5. **Model ID:** `claude-3-5-sonnet` или `gpt-4o`.

### 2.4. Continue.dev (VS Code & JetBrains)
В конфигурационном файле `~/.continue/config.json`:
```json
{
  "models": [
    {
      "title": "NeuroGateway GPT-4o",
      "provider": "openai",
      "model": "gpt-4o",
      "apiKey": "ng-live-your-token-here",
      "apiBase": "https://api.neurogateway.io/v1"
    }
  ]
}
```

### 2.5. Roo Code (VS Code Extension)
1. Выберите режим `OpenAI Compatible` в настройках провайдера.
2. В поле `Base URL` введите `https://api.neurogateway.io/v1`.
3. Укажите API-ключ `ng-live-...`.

---

## 3. Чат-интерфейсы и No-Code / Automation платформы

### 3.1. OpenWebUI
В `docker-compose.yaml` или `.env`:
```yaml
OPENAI_API_BASE_URL: "https://api.neurogateway.io/v1"
OPENAI_API_KEY: "ng-live-your-token-here"
```

### 3.2. LibreChat
В файле `librechat.yaml`:
```yaml
endpoints:
  custom:
    - name: "NeuroGateway"
      apiKey: "ng-live-your-token-here"
      baseURL: "https://api.neurogateway.io/v1"
      models:
        default: ["gpt-4o", "claude-3-5-sonnet", "deepseek-chat"]
```

### 3.3. n8n Workflow Automation
1. Добавьте ноду **OpenAI**.
2. В разделе Credentials создайте новое подключение.
3. Установите:
   - **Base URL:** `https://api.neurogateway.io/v1`
   - **API Key:** `ng-live-your-token-here`

### 3.4. Dify.ai Platform
1. Перейдите в `Settings` $\rightarrow$ `Model Provider` $\rightarrow$ `OpenAI-API-compatible`.
2. **Server URL:** `https://api.neurogateway.io/v1`.
3. **API Key:** `ng-live-your-token-here`.
