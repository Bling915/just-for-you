function imageContentType(path, fileType) {
  if (/\.jpe?g(?:$|\?)/i.test(path)) return 'image/jpeg'
  if (/\.png(?:$|\?)/i.test(path)) return 'image/png'
  if (/\.webp(?:$|\?)/i.test(path)) return 'image/webp'
  if (['image/jpeg','image/png','image/webp'].includes(fileType)) return fileType
  throw new Error('仅支持 JPG、PNG 和 WebP 图片')
}
module.exports={imageContentType}
