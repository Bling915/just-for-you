const test = require('node:test');
const assert = require('node:assert/strict');

const { countRunes, truncateRunes } = require('../utils/runes');

test('countRunes counts Chinese and emoji as visible characters', () => {
  assert.equal(countRunes('我想你🌙'), 4);
});

test('truncateRunes never splits a surrogate pair', () => {
  assert.equal(truncateRunes('你🌙好', 2), '你🌙');
});

test('rune helpers handle nullish values', () => {
  assert.equal(countRunes(null), 0);
  assert.equal(truncateRunes(undefined, 3), '');
});
