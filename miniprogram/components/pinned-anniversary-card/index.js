Component({properties:{items:{type:Array,value:[]}},methods:{open(e){this.triggerEvent('open',{id:e.currentTarget.dataset.id})}}})
