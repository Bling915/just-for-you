function toText(value) {
  return value == null ? '' : String(value);
}

function countRunes(value) {
  return Array.from(toText(value)).length;
}

function truncateRunes(value, limit) {
  return Array.from(toText(value)).slice(0, Math.max(0, limit)).join('');
}

module.exports = { countRunes, truncateRunes };
