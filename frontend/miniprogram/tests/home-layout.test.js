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

test('moment pages use inline compose actions instead of a floating button', () => {
  for (const page of ['home', 'timeline', 'personal-moments', 'moment-spaces']) {
    const view = read(`pages/${page}/index.wxml`)
    assert.doesNotMatch(view, /floating-compose-button/, `${page} still renders a floating compose button`)
    assert.match(view, /class="[^"]*inline-compose/, `${page} has no inline compose action`)
  }
})

test('moment owner actions use a compact card popover instead of the system action sheet', () => {
  const view = read('components/moment-card/index.wxml')
  assert.match(view, /class="owner-menu /)
  assert.match(view, /bindtap="openActions"/)
  assert.match(view, /class="owner-popover"/)
  assert.match(view, /class="action-icon edit-icon"/)
  assert.match(view, /class="action-icon delete-icon"/)
  assert.doesNotMatch(read('components/moment-card/index.js'), /showActionSheet/)
})

test('moment spaces is a bottom tab alongside home timeline and me', () => {
  const app = JSON.parse(read('app.json'))
  const tabs = app.tabBar.list.map(item => item.pagePath)
  assert.deepEqual(tabs, ['pages/home/index', 'pages/moment-spaces/index', 'pages/timeline/index', 'pages/me/index'])
})
