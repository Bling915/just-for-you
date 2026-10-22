function compressImage(wxAPI,filePath,quality=78){return new Promise(resolve=>{if(!wxAPI||typeof wxAPI.compressImage!=='function')return resolve(filePath);wxAPI.compressImage({src:filePath,quality,success:result=>resolve(result.tempFilePath||filePath),fail:()=>resolve(filePath)})})}
const compressSelectedImage=(filePath)=>compressImage(wx,filePath)
module.exports={compressImage,compressSelectedImage}
