const { http }=require('./http-client')
const togetherService={
 listAnniversaries:()=>http.get('/v1/anniversaries'),createAnniversary:input=>http.post('/v1/anniversaries',input),deleteAnniversary:id=>http.delete(`/v1/anniversaries/${id}`),
 listTodos:()=>http.get('/v1/todos'),createTodo:input=>http.post('/v1/todos',input),setTodoCompleted:(id,completed)=>http.patch(`/v1/todos/${id}`,{completed}),deleteTodo:id=>http.delete(`/v1/todos/${id}`),
 listWishes:()=>http.get('/v1/wishes'),createWish:input=>http.post('/v1/wishes',input),setWishCompleted:(id,completed,linkedMomentId)=>http.patch(`/v1/wishes/${id}`,{completed,linkedMomentId}),deleteWish:id=>http.delete(`/v1/wishes/${id}`),
}
function splitCompleted(items=[]){return{active:items.filter(x=>x.status==='open'||x.status==='planned'),completed:items.filter(x=>x.status==='completed')}}
module.exports={togetherService,splitCompleted}
