const { http } = require('./http-client')

function readFile(filePath){return new Promise((resolve,reject)=>wx.getFileSystemManager().readFile({filePath,success:result=>resolve(result.data),fail:reject}))}
async function putFile({filePath,uploadURL,contentType,onStateChange=()=>{}}){const data=await readFile(filePath);onStateChange('uploading');return new Promise((resolve,reject)=>wx.request({url:uploadURL,method:'PUT',data,header:{'Content-Type':contentType},success:response=>response.statusCode>=200&&response.statusCode<300?resolve():reject(new Error('照片上传失败')),fail:reject}))}
const uploadService={requestGrant:({contentType,size})=>http.post('/v1/uploads',{contentType,size}),putFile}
module.exports={uploadService,putFile}
