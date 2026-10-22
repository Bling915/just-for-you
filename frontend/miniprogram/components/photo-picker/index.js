const { imageContentType }=require('../../utils/media-type.js')
Component({
  properties:{files:{type:Array,value:[]}},
  methods:{
    choose(){const count=9-this.data.files.length;if(count<=0)return wx.showToast({title:'最多选择 9 张',icon:'none'});wx.chooseMedia({count,mediaType:['image'],sourceType:['album','camera'],success:result=>{try{const added=result.tempFiles.map(file=>({path:file.tempFilePath,thumb:file.tempFilePath,size:file.size,contentType:imageContentType(file.tempFilePath,file.fileType)}));this.emit(this.data.files.concat(added))}catch(error){wx.showToast({title:error.message,icon:'none'})}}})},
    remove(event){const next=this.data.files.slice();next.splice(Number(event.currentTarget.dataset.index),1);this.emit(next)},
    preview(event){const urls=this.data.files.map(file=>file.thumb||file.path);wx.previewImage({current:urls[event.currentTarget.dataset.index],urls})},
    move(event){const index=Number(event.currentTarget.dataset.index),direction=Number(event.currentTarget.dataset.direction),target=index+direction;if(target<0||target>=this.data.files.length)return;const next=this.data.files.slice();[next[index],next[target]]=[next[target],next[index]];this.emit(next)},
    retry(event){this.triggerEvent('retry',{index:Number(event.currentTarget.dataset.index)})},emit(files){this.triggerEvent('change',{files})},
  },
})
