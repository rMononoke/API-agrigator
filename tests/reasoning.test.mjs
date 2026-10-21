import { readFileSync } from 'node:fs';
import test from 'node:test';
import assert from 'node:assert/strict';
import Ajv2020 from 'ajv/dist/2020.js';

const ajv = new Ajv2020({ allErrors: true, strict: false });
const reqSchema = JSON.parse(
  readFileSync(new URL('../schema/llm-request.schema.json', import.meta.url)),
);
const respSchema = JSON.parse(
  readFileSync(new URL('../schema/llm-response.schema.json', import.meta.url)),
);

const validateReq = ajv.compile(reqSchema);
const validateResp = ajv.compile(respSchema);

test('reasoning: request schema accepts valid reasoning configurations', () => {
  const payloads = [
    { effort: 'low' },
    { effort: 'medium', max_tokens: 4096 },
    { effort: 'high', max_tokens: 16384, exclude: true },
    { effort: 'minimal' },
    { effort: 'none' }
  ];

  for (const r of payloads) {
    const body = {
      model: 'o3-mini',
      messages: [{ role: 'user', content: 'Solve this equation: 2x + 5 = 15' }],
      reasoning: r
    };
    const valid = validateReq(body);
    assert.equal(valid, true, `Payload failed: ${JSON.stringify(r)} errors: ${JSON.stringify(validateReq.errors)}`);
  }
});

test('reasoning: request schema rejects invalid reasoning effort', () => {
  const body = {
    model: 'o3-mini',
    messages: [{ role: 'user', content: 'hello' }],
    reasoning: { effort: 'super-high' }
  };
  const valid = validateReq(body);
  assert.equal(valid, false);
});

test('reasoning: request schema rejects negative max_tokens in reasoning', () => {
  const body = {
    model: 'o3-mini',
    messages: [{ role: 'user', content: 'hello' }],
    reasoning: { effort: 'medium', max_tokens: -100 }
  };
  const valid = validateReq(body);
  assert.equal(valid, false);
});

test('reasoning: response schema accepts reasoning and reasoning_content in assistant message', () => {
  const resp = {
    id: 'chatcmpl-reason-123',
    object: 'chat.completion',
    created: 1729540000,
    model: 'deepseek-reasoner',
    choices: [
      {
        index: 0,
        message: {
          role: 'assistant',
          content: 'The solution is x = 5.',
          reasoning: 'First subtract 5 from both sides: 2x = 10, then divide by 2: x = 5.',
          reasoning_content: 'First subtract 5 from both sides...'
        },
        finish_reason: 'stop'
      }
    ],
    usage: {
      prompt_tokens: 25,
      completion_tokens: 40,
      total_tokens: 65,
      reasoning_tokens: 30
    }
  };

  const valid = validateResp(resp);
  assert.equal(valid, true, `Validation failed: ${JSON.stringify(validateResp.errors)}`);
});

test('reasoning: response usage preserves reasoning_tokens and rejects negative reasoning_tokens', () => {
  const validResp = {
    id: 'chatcmpl-o1-999',
    object: 'chat.completion',
    created: 1729540000,
    model: 'o1',
    choices: [
      {
        index: 0,
        message: {
          role: 'assistant',
          content: 'Verified.'
        },
        finish_reason: 'stop'
      }
    ],
    usage: {
      prompt_tokens: 100,
      completion_tokens: 500,
      total_tokens: 600,
      reasoning_tokens: 450
    }
  };
  assert.equal(validateResp(validResp), true);

  const invalidResp = {
    ...validResp,
    usage: {
      ...validResp.usage,
      reasoning_tokens: -10
    }
  };
  assert.equal(validateResp(invalidResp), false);
});

// Simulation of provider mapping logic
function mapReasoningToProvider(provider, cfg) {
  if (!cfg || cfg.effort === 'none') return {};

  switch (provider) {
    case 'openai': {
      let effort = cfg.effort;
      if (effort === 'minimal') effort = 'low';
      if (!['low', 'medium', 'high'].includes(effort)) effort = 'medium';
      return { reasoning_effort: effort };
    }
    case 'anthropic': {
      let budget = cfg.max_tokens;
      if (!budget) {
        if (cfg.effort === 'minimal') budget = 1024;
        else if (cfg.effort === 'low') budget = 2048;
        else if (cfg.effort === 'high') budget = 16384;
        else budget = 4096;
      }
      return { thinking: { type: 'enabled', budget_tokens: budget } };
    }
    case 'gemini': {
      let budget = cfg.max_tokens || (cfg.effort === 'high' ? 16384 : 4096);
      return { thinkingConfig: { thinkingBudget: budget } };
    }
    case 'deepseek': {
      return { reasoning_content_enabled: true };
    }
    default:
      return { reasoning_effort: cfg.effort };
  }
}

test('reasoning: provider mapping generates correct upstream parameters', () => {
  const oai = mapReasoningToProvider('openai', { effort: 'minimal' });
  assert.deepEqual(oai, { reasoning_effort: 'low' });

  const ant = mapReasoningToProvider('anthropic', { effort: 'high' });
  assert.deepEqual(ant, { thinking: { type: 'enabled', budget_tokens: 16384 } });

  const gem = mapReasoningToProvider('gemini', { effort: 'medium', max_tokens: 8192 });
  assert.deepEqual(gem, { thinkingConfig: { thinkingBudget: 8192 } });

  const ds = mapReasoningToProvider('deepseek', { effort: 'medium' });
  assert.deepEqual(ds, { reasoning_content_enabled: true });
});
