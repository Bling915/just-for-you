function dailyQuestionView(value){if(!value)return{state:'empty',question:'',myAnswer:'',partnerAnswer:''};return{...value,state:value.unlocked?'unlocked':value.answered?'waiting':'answer',myAnswer:value.myAnswer&&value.myAnswer.text||'',partnerAnswer:value.unlocked&&value.partnerAnswer?value.partnerAnswer.text:''}}
module.exports={dailyQuestionView}
