Component({ properties: { title: { type:String,value:'还没有小事' }, description:{type:String,value:'第一件值得记住的事，正在等你写下。'}, action:{type:String,value:''} }, methods:{act(){this.triggerEvent('action')}} })
