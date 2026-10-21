import { readFileSync } from 'node:fs';
import test from 'node:test';
import assert from 'node:assert/strict';
import Ajv2020 from 'ajv/dist/2020.js';
import crypto from 'node:crypto';

const ajv = new Ajv2020({ allErrors: true, strict: false });
const reqSchema = JSON.parse(
  readFileSync(new URL('../schema/llm-request.schema.json', import.meta.url)),
);
const respSchema = JSON.parse(
  readFileSync(new URL('../schema/llm-response.schema.json', import.meta.url)),
);

const validateReq = ajv.compile(reqSchema);
const validateResp = ajv.compile(respSchema);

test('caching: request schema accepts cache_control ephemeral on messages', () => {
  const body = {
    model: 'claude-3-5-sonnet',
    messages: [
      {
        role: 'system',
        content: 'You are an enterprise AI assistant with a large knowledge base.',
        cache_control: { type: 'ephemeral' }
      },
      {
        role: 'user',
        content: 'Summarize the documentation.'
      }
    ]
  };

  const valid = validateReq(body);
  assert.equal(valid, true, `Validation failed: ${JSON.stringify(validateReq.errors)}`);
});

test('caching: request schema rejects invalid cache_control type', () => {
  const body = {
    model: 'claude-3-5-sonnet',
    messages: [
      {
        role: 'system',
        content: 'You are an enterprise AI assistant.',
        cache_control: { type: 'permanent_disk' }
      },
      {
        role: 'user',
        content: 'ping'
      }
    ]
  };

  const valid = validateReq(body);
  assert.equal(valid, false);
});

test('caching: response schema accepts cached_tokens and cache_write_tokens', () => {
  const resp = {
    id: 'chatcmpl-cache-test-1',
    object: 'chat.completion',
    created: 1729540000,
    model: 'claude-3-5-sonnet',
    choices: [
      {
        index: 0,
        message: {
          role: 'assistant',
          content: 'Cache hit verified.'
        },
        finish_reason: 'stop'
      }
    ],
    usage: {
      prompt_tokens: 12000,
      completion_tokens: 150,
      total_tokens: 12150,
      cached_tokens: 10500,
      cache_write_tokens: 0
    }
  };

  const valid = validateResp(resp);
  assert.equal(valid, true, `Validation failed: ${JSON.stringify(validateResp.errors)}`);
});

// Deterministic prefix hash computation for sticky routing
function computePrefixHash(messages) {
  const h = crypto.createHash('sha256');
  const parts = [];
  for (let i = 0; i < messages.length; i++) {
    const m = messages[i];
    if (m.role === 'system' || m.role === 'developer' || i < 2) {
      parts.push(`${m.role}:${m.content}`);
    }
  }
  h.update(parts.join('||'));
  return h.digest('hex').slice(0, 16);
}

test('caching: prefix hash is deterministic for identical prompt prefixes (sticky routing)', () => {
  const messagesA = [
    { role: 'system', content: 'Enterprise Knowledge Base V4' },
    { role: 'user', content: 'First query' }
  ];
  const messagesB = [
    { role: 'system', content: 'Enterprise Knowledge Base V4' },
    { role: 'user', content: 'First query' }
  ];
  const messagesDiff = [
    { role: 'system', content: 'Enterprise Knowledge Base V5' },
    { role: 'user', content: 'First query' }
  ];

  const hashA = computePrefixHash(messagesA);
  const hashB = computePrefixHash(messagesB);
  const hashDiff = computePrefixHash(messagesDiff);

  assert.equal(hashA, hashB, 'Hashes for identical prefixes must match');
  assert.notEqual(hashA, hashDiff, 'Hashes for different prefixes must differ');
  assert.equal(hashA.length, 16);
});

// Differential billing calculation test
function calculateSavings(pricing, usage) {
  const nonCached = Math.max(0, usage.prompt_tokens - (usage.cached_tokens || 0));
  const baseCost = (usage.prompt_tokens / 1_000_000) * pricing.baseInput + (usage.completion_tokens / 1_000_000) * pricing.baseOutput;
  const actualCost = (nonCached / 1_000_000) * pricing.baseInput +
                     ((usage.cached_tokens || 0) / 1_000_000) * pricing.cacheRead +
                     ((usage.cache_write_tokens || 0) / 1_000_000) * pricing.cacheWrite +
                     (usage.completion_tokens / 1_000_000) * pricing.baseOutput;

  const saved = Math.max(0, baseCost - actualCost);
  const discountPct = baseCost > 0 ? (saved / baseCost) * 100 : 0;
  return { baseCost, actualCost, saved, discountPct };
}

test('caching: differential pricing rewards 90% discount on Claude 3.5 Sonnet cache read', () => {
  const claudePricing = {
    baseInput: 3.00,
    cacheRead: 0.30,   // 90% discount
    cacheWrite: 3.75,  // 25% surcharge
    baseOutput: 15.00
  };

  const usage = {
    prompt_tokens: 10000,
    completion_tokens: 1000,
    cached_tokens: 10000,
    cache_write_tokens: 0
  };

  const res = calculateSavings(claudePricing, usage);
  // Base cost = (10000/1e6)*3.00 + (1000/1e6)*15 = 0.03 + 0.015 = 0.045
  // Actual cost = (10000/1e6)*0.30 + (1000/1e6)*15 = 0.003 + 0.015 = 0.018
  // Saved = 0.045 - 0.018 = 0.027 (60% total request discount, 90% prompt discount)
  assert.equal(res.baseCost.toFixed(4), '0.0450');
  assert.equal(res.actualCost.toFixed(4), '0.0180');
  assert.equal(res.saved.toFixed(4), '0.0270');
  assert.equal(res.discountPct.toFixed(1), '60.0');
});
