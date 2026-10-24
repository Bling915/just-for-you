const test = require('node:test');
const assert = require('node:assert/strict');
const { hasContent, decideOnExit, isDraftEmpty, buildDraftPayload, restoreDraft, restoreStatusEditor, selectQuickStatus, inputCustomStatus } = require('../utils/compose-guard');

/**
 * 新增记录页的离开决策。
 *
 * 原来 onUnload 里静默 save()，用户不知道草稿被存了；
 * 现在改成有内容时弹窗询问，空白时安静离开。
 */

test('空白页直接走，不打扰用户', () => {
  assert.equal(decideOnExit({}, {}).action, 'leave');
  assert.equal(decideOnExit({ text: '   ' }, {}).action, 'leave');
  assert.equal(decideOnExit({ text: '', status: '', files: [] }, {}).action, 'leave');
});

test('有文字内容时弹窗询问', () => {
  assert.equal(decideOnExit({ text: '今天天气不错' }, {}).action, 'confirm');
});

test('只有状态标签也算有内容', () => {
  assert.equal(decideOnExit({ status: '想你了' }, {}).action, 'confirm');
});

test('只有地点也算有内容', () => {
  assert.equal(decideOnExit({ placeName: '外滩' }, {}).action, 'confirm');
});

test('只有坐标也算有内容（选了位置）', () => {
  assert.equal(decideOnExit({ latitude: 31.23, longitude: 121.47 }, {}).action, 'confirm');
});

test('只有照片也算有内容', () => {
  assert.equal(decideOnExit({ files: [{ path: '/tmp/a.jpg' }] }, {}).action, 'confirm');
});

test('lat 明确为 null 不算有内容', () => {
  assert.equal(decideOnExit({ latitude: null, longitude: null }, {}).action, 'leave');
});

test('发布中禁止离开，避免上传中断', () => {
  const d = decideOnExit({ text: '正在写的东西' }, { publishing: true });
  assert.equal(d.action, 'blocked');
  assert.ok(d.reason);
});

test('发布中即使空白也拦住', () => {
  assert.equal(decideOnExit({}, { publishing: true }).action, 'blocked');
});

test('hasContent 判定与 decideOnExit 一致', () => {
  const cases = [
    [{}, false],
    [{ text: ' ' }, false],
    [{ text: 'x' }, true],
    [{ status: '有点累' }, true],
    [{ placeName: '家' }, true],
    [{ latitude: 0, longitude: 0 }, true], // 0 是有效坐标，不能被当成空
    [{ latitude: null }, false],
    [{ files: [] }, false],
    [{ files: [{}] }, true],
  ];
  for (const [form, expected] of cases) {
    assert.equal(hasContent(form), expected, JSON.stringify(form));
  }
});

test('isDraftEmpty 与 hasContent 判断一致', () => {
  assert.equal(isDraftEmpty({}), true);
  assert.equal(isDraftEmpty({ text: '' }), true);
  assert.equal(isDraftEmpty({ text: '写了一半' }), false);
  assert.equal(isDraftEmpty({ files: [{ path: '/a.jpg' }] }), false);
});

test('确认保存与拒绝保存两条路径的草稿结果不同', () => {
  const { createDraftStore } = require('../store/draft');
  const memory = {};
  const storage = { get: (k) => memory[k], set: (k, v) => (memory[k] = v), remove: (k) => delete memory[k] };

  // 拒绝保存
  const storeA = createDraftStore(storage);
  storeA.patch({ text: '写了又删' });
  storeA.clear();
  assert.equal(isDraftEmpty(storeA.snapshot()), true);
  assert.equal(storage.get('momentDraft'), undefined);

  // 确认保存
  const storeB = createDraftStore(storage);
  storeB.patch({ text: '留着下次写' });
  const restored = createDraftStore(storage);
  assert.equal(restored.load().text, '留着下次写');
});

test('发布成功后清空即回到全新一页', () => {
  // reset() 的语义：发布成功后页面 data 与空草稿等价
  const fresh = { text: '', status: '', placeName: '', latitude: null, longitude: null, files: [], textCount: 0, placeCount: 0, restored: false };
  assert.equal(hasContent(fresh), false);
  assert.equal(isDraftEmpty(fresh), true);
});

test('草稿不保存照片：临时文件路径跨会话必然失效', () => {
  const payload = buildDraftPayload({
    text: '今天去了海边',
    status: '想你了',
    placeName: '外滩',
    latitude: 31.23,
    longitude: 121.47,
    occurredAt: '2026-10-22T10:00:00.000Z',
    files: [{ path: 'wxfile://tmp_abc.jpg' }],
  });
  assert.deepEqual(payload.files, []);
  assert.equal(payload.text, '今天去了海边');
  assert.equal(payload.placeName, '外滩');
  assert.equal(payload.latitude, 31.23);
});

test('草稿缺字段时补成安全的空值', () => {
  const payload = buildDraftPayload({});
  assert.equal(payload.text, '');
  assert.equal(payload.status, '');
  assert.equal(payload.placeName, '');
  assert.equal(payload.latitude, null);
  assert.equal(payload.longitude, null);
  assert.equal(payload.occurredAt, '');
});

test('恢复草稿时丢弃照片，其余字段原样带回', () => {
  const form = restoreDraft({
    text: '上次写到一半',
    status: '有点累',
    placeName: '家',
    latitude: 31.2,
    longitude: 121.4,
    occurredAt: '2026-10-22T10:00:00.000Z',
	spaceType: 'personal',
	visibility: 'partner',
    files: [{ path: 'wxfile://tmp_gone.jpg' }],
  });
  assert.deepEqual(form.files, []);
  assert.equal(form.text, '上次写到一半');
  assert.equal(form.status, '有点累');
  assert.equal(form.placeName, '家');
  assert.equal(form.latitude, 31.2);
  assert.equal(form.longitude, 121.4);
	assert.equal(form.spaceType, 'personal');
	assert.equal(form.visibility, 'partner');
});

test('快捷状态和自定义状态共用一个 status 字段且互斥', () => {
  assert.deepEqual(selectQuickStatus('', '想你了'), { status: '想你了', selectedQuickStatus: '想你了', customStatus: '' });
  assert.deepEqual(inputCustomStatus('今天很开心'), { status: '今天很开心', selectedQuickStatus: '', customStatus: '今天很开心' });
  assert.deepEqual(selectQuickStatus('想你了', '想你了'), { status: '', selectedQuickStatus: '', customStatus: '' });
});

test('恢复草稿时快捷值恢复为选中态，自定义值恢复到输入框', () => {
  assert.deepEqual(restoreStatusEditor('在上课'), { status: '在上课', selectedQuickStatus: '在上课', customStatus: '' });
  assert.deepEqual(restoreStatusEditor('等你下班'), { status: '等你下班', selectedQuickStatus: '', customStatus: '等你下班' });
});

test('草稿保留显式空间，新草稿仍由发布页默认到共同空间', () => {
	const payload = buildDraftPayload({ text: '写了一半', spaceType: 'personal', visibility: 'private' });
	assert.equal(payload.spaceType, 'personal');
	assert.equal(payload.visibility, 'private');
	const restored = restoreDraft(payload);
	assert.equal(restored.spaceType, 'personal');
	assert.equal(restored.visibility, 'private');
});

test('空草稿恢复出来是干净表单', () => {
  const form = restoreDraft({});
  assert.equal(form.text, '');
  assert.equal(form.latitude, null);
  assert.deepEqual(form.files, []);
  assert.equal(hasContent(form), false);
});

test('草稿里残留的旧照片不会被恢复出来', () => {
  // 模拟用户上一次选了照片没发布，草稿里存了临时路径
  const store = require('../store/draft');
  const memory = {};
  const storage = { get: (k) => memory[k], set: (k, v) => (memory[k] = v), remove: (k) => delete memory[k] };
  const { createDraftStore } = store;
  const draft = createDraftStore(storage);
  draft.patch(buildDraftPayload({ text: '带照片的草稿', files: [{ path: 'wxfile://tmp_x.jpg' }] }));
  const restored = restoreDraft(draft.load());
  assert.deepEqual(restored.files, []);
  assert.equal(restored.text, '带照片的草稿');
});

test('防抖保存：连续输入只落盘一次', async () => {
  const { createDraftStore } = require('../store/draft');
  const memory = {};
  const storage = { get: (k) => memory[k], set: (k, v) => (memory[k] = v), remove: (k) => delete memory[k] };
  const draft = createDraftStore(storage);

  let writes = 0;
  const originalSet = storage.set;
  storage.set = (k, v) => { writes += 1; originalSet(k, v); };

  let pending = null;
  const save = () => { if (pending) clearTimeout(pending); pending = setTimeout(() => { pending = null; draft.patch(buildDraftPayload({ text: '长长长长长' })); }, 400); };
  for (let i = 0; i < 20; i += 1) save();
  assert.equal(writes, 0, '防抖期间不该写盘');

  await new Promise((resolve) => setTimeout(resolve, 500));
  assert.equal(writes, 1, '20 次输入只应该落盘 1 次');
  assert.equal(draft.load().text, '长长长长长');
});
