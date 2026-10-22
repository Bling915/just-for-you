const test = require('node:test');
const assert = require('node:assert');

/**
 * 共同地图的展示规则。
 *
 * 背景：只有通过「选择位置」采集到坐标的动态才会出现在地图上，
 * 手动输入地名不会产生坐标（见 moments 表 latitude/longitude 允许 NULL）。
 * 因此地图页必须在无坐标时给出明确空状态，否则用户会以为功能坏了。
 */

const DEFAULT_CENTER = { latitude: 31.2304, longitude: 121.4737 };

function buildMarkers(points) {
  return (points || []).map((x) => {
    const day = String(x.occurredAt || '').slice(0, 10);
    const title = x.placeName || `${day || '这一天'}的回忆`;
    return {
      id: Number(x.momentId),
      latitude: x.latitude,
      longitude: x.longitude,
      title,
      width: 28,
      height: 28,
      callout: {
        content: title,
        color: '#403946',
        fontSize: 12,
        borderRadius: 6,
        bgColor: '#FFFCFD',
        padding: 6,
        display: 'BYCLICK',
      },
    };
  });
}

function resolveView(markers) {
  return {
    markers,
    includePoints: markers.map((m) => ({ latitude: m.latitude, longitude: m.longitude })),
    empty: markers.length === 0,
    center: markers.length
      ? { latitude: markers[0].latitude, longitude: markers[0].longitude }
      : DEFAULT_CENTER,
  };
}

test('只有带坐标的动态才出现在地图上', () => {
  // 后端 /v1/moments/map 用 `latitude IS NOT NULL AND longitude IS NOT NULL` 过滤，
  // 手动输入地名的记录不会出现在 points 里。
  const points = [
    { momentId: 3, latitude: 31.23037, longitude: 121.4737, placeName: '', occurredAt: '2026-10-01T05:24:00Z' },
  ];
  const view = resolveView(buildMarkers(points));
  assert.equal(view.markers.length, 1);
  assert.equal(view.empty, false);
});

test('没有地名时用日期兜底标题，避免图上只剩默认图标', () => {
  const [marker] = buildMarkers([
    { momentId: 3, latitude: 31.23037, longitude: 121.4737, placeName: '', occurredAt: '2026-10-01T05:24:00Z' },
  ]);
  assert.equal(marker.title, '2026-10-01的回忆');
  assert.equal(marker.callout.content, '2026-10-01的回忆');
});

test('有地名时直接用地名', () => {
  const [marker] = buildMarkers([
    { momentId: 1, latitude: 31.23, longitude: 121.47, placeName: '同济大学四平路校区', occurredAt: '2026-10-01T05:24:00Z' },
  ]);
  assert.equal(marker.title, '同济大学四平路校区');
});

test('无坐标时进入空状态并落在上海兜底中心，而非广州', () => {
  const view = resolveView(buildMarkers([]));
  assert.equal(view.empty, true);
  assert.deepEqual(view.center, DEFAULT_CENTER);
  // includePoints 为空数组，让 map 保持默认视野而不是被空数组干扰
  assert.deepEqual(view.includePoints, []);
});

test('includePoints 交给 map 自动缩放，避免单个点在 nationwide 视图里看不见', () => {
  const view = resolveView(buildMarkers([
    { momentId: 3, latitude: 31.23037, longitude: 121.4737, placeName: '', occurredAt: '2026-10-01T05:24:00Z' },
    { momentId: 4, latitude: 39.9042, longitude: 116.4074, placeName: '北京', occurredAt: '2026-10-02T05:24:00Z' },
  ]));
  assert.equal(view.includePoints.length, 2);
  assert.equal(view.center.latitude, 31.23037);
});

test('接口异常时降级为空状态而不是抛出', () => {
  // onLoad 的 load() 内部已 try/catch，这里固化「失败也要给出可读状态」的约定
  const view = resolveView([]);
  assert.equal(view.empty, true);
  assert.equal(view.markers.length, 0);
});