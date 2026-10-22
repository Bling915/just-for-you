const DAY=86400000
function parseDay(value){
  const match=String(value||'').match(/^(\d{4})-(\d{2})-(\d{2})/);if(!match)return null
  const y=Number(match[1]),m=Number(match[2]),d=Number(match[3]);const stamp=Date.UTC(y,m-1,d);const date=new Date(stamp)
  return date.getUTCFullYear()===y&&date.getUTCMonth()===m-1&&date.getUTCDate()===d?{y,m,d,stamp}:null
}
function shanghaiToday(now){const shifted=new Date((now instanceof Date?now:new Date(now)).getTime()+8*60*60*1000);return{y:shifted.getUTCFullYear(),m:shifted.getUTCMonth()+1,d:shifted.getUTCDate(),stamp:Date.UTC(shifted.getUTCFullYear(),shifted.getUTCMonth(),shifted.getUTCDate())}}
function daysInMonth(y,m){return new Date(Date.UTC(y,m,0)).getUTCDate()}
function annualTarget(source,today){let y=today.y;let d=Math.min(source.d,daysInMonth(y,source.m));let stamp=Date.UTC(y,source.m-1,d);if(stamp<today.stamp){y++;d=Math.min(source.d,daysInMonth(y,source.m));stamp=Date.UTC(y,source.m-1,d)}return{y,m:source.m,d,stamp}}
function iso(day){return `${day.y}-${String(day.m).padStart(2,'0')}-${String(day.d).padStart(2,'0')}`}
function anniversaryView(item,now=new Date()){
  const source=parseDay(item&&item.date);const today=shanghaiToday(now);if(!source||!Number.isFinite(today.stamp))return{...(item||{}),valid:false,days:null,headline:'日期待完善'}
  const mode=item.countMode==='elapsed'?'elapsed':'countdown'
  if(mode==='elapsed'){const days=Math.max(1,Math.floor((today.stamp-source.stamp)/DAY)+1);return{...item,valid:true,mode,days,targetDate:iso(source),headline:`第 ${days} 天`,sortStamp:source.stamp}}
  const target=item.repeatsYearly?annualTarget(source,today):source;const days=Math.max(0,Math.floor((target.stamp-today.stamp)/DAY));return{...item,valid:true,mode,days,targetDate:iso(target),headline:days===0?'就是今天':`还有 ${days} 天`,sortStamp:target.stamp}
}
function sortAnniversaries(items,now=new Date()){return (items||[]).map(item=>anniversaryView(item,now)).sort((a,b)=>{
  if(Boolean(a.pinned)!==Boolean(b.pinned))return a.pinned?-1:1
  if(a.mode!==b.mode)return a.mode==='countdown'?-1:1
  if(a.sortStamp!==b.sortStamp)return a.mode==='countdown'?a.sortStamp-b.sortStamp:b.sortStamp-a.sortStamp
  return Number(a.id)-Number(b.id)
})}
function splitAnniversaries(items,now=new Date()){const sorted=sortAnniversaries(items,now);return{pinned:sorted.filter(x=>x.pinned),all:sorted.filter(x=>!x.pinned)}}
module.exports={anniversaryView,sortAnniversaries,splitAnniversaries}
