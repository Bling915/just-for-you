const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')

const root = path.join(__dirname, '..')
const read = relative => fs.readFileSync(path.join(root, relative), 'utf8')

test('home keeps daily question and resurfaced memory inside one collapsible inbox', () => {
  const view = read('pages/home/index.wxml')
  assert.match(view, /class="today-inbox/)
  assert.match(view, /bindtap="toggleInbox"/)
  assert.match(view, /wx:if="{{inboxExpanded}}"[\s\S]*daily-question-card/)
  assert.match(view, /wx:if="{{inboxExpanded}}"[\s\S]*memory-resurface-card/)
})

test('compose button is outside the animated page so fixed positioning uses the viewport', () => {
  for (const page of ['home', 'timeline', 'personal-moments', 'moment-spaces']) {
    const view = read(`pages/${page}/index.wxml`)
    assert.match(view, /<\/view>\s*<floating-compose-button/, `${page} compose button is still inside the page root`)
  }
})

test('moment owner actions use one polished menu instead of bare edit and delete buttons', () => {
  const view = read('components/moment-card/index.wxml')
  assert.match(view, /class="owner-menu"/)
  assert.match(view, /bindtap="openActions"/)
  assert.doesNotMatch(view, />编辑<\/button>/)
  assert.doesNotMatch(view, />删除<\/button>/)
})
