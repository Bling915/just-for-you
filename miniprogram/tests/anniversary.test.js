const test=require('node:test');const assert=require('node:assert/strict');
const {anniversaryView,sortAnniversaries,splitAnniversaries}=require('../utils/anniversary')

test('same Shanghai natural day is day one for elapsed and zero days for countdown',()=>{
  const now=new Date('2026-10-21T16:30:00.000Z') // 2026-10-22 00:30 +08
  assert.equal(anniversaryView({date:'2026-10-22',countMode:'elapsed'},now).days,1)
  assert.equal(anniversaryView({date:'2026-10-22',countMode:'countdown'},now).days,0)
})
test('countdown crosses year and repeats annually',()=>{
  const got=anniversaryView({date:'2020-01-02',countMode:'countdown',repeatsYearly:true},new Date('2026-12-31T04:00:00Z'))
  assert.equal(got.days,2);assert.equal(got.targetDate,'2027-01-02')
})
test('Feb 29 repeats on Feb 28 in a non-leap year',()=>{
  const got=anniversaryView({date:'2024-02-29',countMode:'countdown',repeatsYearly:true},new Date('2025-02-27T04:00:00Z'))
  assert.equal(got.days,1);assert.equal(got.targetDate,'2025-02-28')
})
test('pinned items sort first then countdown soonest and elapsed newest',()=>{
  const now=new Date('2026-10-22T04:00:00Z');const items=[
    {id:4,date:'2024-01-01',countMode:'elapsed',pinned:false},
    {id:2,date:'2026-10-24',countMode:'countdown',pinned:true},
    {id:3,date:'2026-10-23',countMode:'countdown',pinned:true},
    {id:1,date:'2025-01-01',countMode:'elapsed',pinned:false},
  ]
  assert.deepEqual(sortAnniversaries(items,now).map(x=>x.id),[3,2,1,4])
  const split=splitAnniversaries(items,now);assert.deepEqual(split.pinned.map(x=>x.id),[3,2]);assert.deepEqual(split.all.map(x=>x.id),[1,4])
})
test('invalid or missing dates return a safe invalid view',()=>{
  for(const date of ['',null,'not-a-date','2026-02-30']){const got=anniversaryView({date,countMode:'elapsed'},new Date());assert.equal(got.valid,false);assert.equal(got.days,null)}
})
