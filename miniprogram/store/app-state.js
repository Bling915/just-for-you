const { sessionStore }=require('./session');const { draftStore }=require('./draft');const { timelineStore }=require('./timeline');const { storage }=require('../utils/storage')
function createAppState({session,draft,timeline,cache}){function clearRelationshipData(){timeline.reset();draft.clear();cache.remove('homeSnapshot');cache.remove('relationshipSnapshot')}function clearPrivateData(){clearRelationshipData();session.clear()}return{clearRelationshipData,clearPrivateData}}
const appState=createAppState({session:sessionStore,draft:draftStore,timeline:timelineStore,cache:storage})
module.exports={createAppState,appState}
