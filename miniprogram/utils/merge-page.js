function mergeMomentPage(current=[],incoming=[]){const byId=new Map();current.concat(incoming).forEach(item=>byId.set(String(item.id),item));return Array.from(byId.values()).sort((a,b)=>{const time=new Date(b.occurredAt||b.createdAt).getTime()-new Date(a.occurredAt||a.createdAt).getTime();if(time)return time;return Number(b.id)-Number(a.id)})}
module.exports={mergeMomentPage}
