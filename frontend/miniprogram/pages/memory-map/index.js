const { momentService } = require('../../services/moments');

// 上海默认兜底：只在没有任何坐标时用，避免页面停在广州造成"地图坏了"的错觉
const DEFAULT_CENTER = { latitude: 31.2304, longitude: 121.4737 };

Page({
  data: {
    markers: [],
    latitude: DEFAULT_CENTER.latitude,
    longitude: DEFAULT_CENTER.longitude,
    // includePoints 由 <map> 自动按 markers 缩放；有坐标时无需手写 scale
    includePoints: [],
    empty: true,
  },
  onLoad() { this.load(); },
  async load() {
    const to = new Date();
    const from = new Date(to);
    from.setFullYear(from.getFullYear() - 5);
    try {
      const result = await momentService.getMap({ from: from.toISOString(), to: to.toISOString() });
      const points = result.points || [];
      const markers = points.map((x) => ({
        id: Number(x.momentId),
        latitude: x.latitude,
        longitude: x.longitude,
        // 没有地名时用日期兜底，否则图上只剩一个孤零零的默认图标
        title: x.placeName || `${String(x.occurredAt || '').slice(0, 10) || '这一天'}的回忆`,
        width: 28,
        height: 28,
        callout: {
          content: x.placeName || `${String(x.occurredAt || '').slice(0, 10)}的回忆`,
          color: '#403946',
          fontSize: 12,
          borderRadius: 6,
          bgColor: '#FFFCFD',
          padding: 6,
          display: 'BYCLICK',
        },
      }));
      this.setData({
        markers,
        includePoints: markers.map((m) => ({ latitude: m.latitude, longitude: m.longitude })),
        empty: markers.length === 0,
        ...(markers.length
          ? { latitude: markers[0].latitude, longitude: markers[0].longitude }
          : DEFAULT_CENTER),
      });
    } catch (_) {
      this.setData({ markers: [], empty: true });
    }
  },
});