const LABELS = { in_class: '在上课', miss_you: '想你了', tired: '有点累' }
Component({
  properties: { moment: { type: Object, value: null } },
  data: { labels: LABELS },
})
