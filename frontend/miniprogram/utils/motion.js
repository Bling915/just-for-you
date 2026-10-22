function motionModel(reduceMotion=false,index=0){if(reduceMotion)return{className:'reduce-motion',delayMs:0,durationMs:160};return{className:'motion scrapbook-motion',delayMs:Math.min(Math.max(0,Number(index)||0),8)*60}}
module.exports={motionModel}
