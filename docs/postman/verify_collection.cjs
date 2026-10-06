// Offline QA of collection structure, JavaScript, request data flow and assertion sensitivity.
// Synthetic response fixtures exercise tests; they do NOT validate the real Go backend.
const fs=require('fs'), path=require('path'), vm=require('vm'), crypto=require('crypto');
const modules=process.argv[2];
const chai=require(modules?path.join(modules,'chai'):'chai');
const Ajv=require(modules?path.join(modules,'ajv'):'ajv');
const newman=require(modules?path.join(modules,'newman'):'newman');
const ajv=new Ajv({schemaId:'auto',allErrors:true,unknownFormats:'ignore'});
ajv.addMetaSchema(require(modules?path.join(modules,'ajv/lib/refs/json-schema-draft-04.json'):'ajv/lib/refs/json-schema-draft-04.json'));
const read=f=>JSON.parse(fs.readFileSync(path.join(__dirname,f),'utf8'));
const collection=read('Sportik.postman_collection.json'), env=read('Local.postman_environment.json'), cases=read('cases.json');
const spec=JSON.parse(fs.readFileSync(path.join(__dirname,'../api/openapi.json'),'utf8'));
const definitions=JSON.parse(collection.variable.find(x=>x.key==='__definitions').value);
const contracts=JSON.parse(collection.variable.find(x=>x.key==='__contracts').value);
const flat=collection.item.flatMap(folder=>folder.item);
for(const event of collection.event)new vm.Script(event.script.exec.join('\n'));
for(const item of flat)for(const event of item.event)new vm.Script(event.script.exec.join('\n'));
const schemaFile=process.argv[3];
if(schemaFile && !ajv.validate(JSON.parse(fs.readFileSync(schemaFile,'utf8')),collection))throw Error(JSON.stringify(ajv.errors));
function assert(value,message){if(!value)throw Error(message);}
function convert(v){
 if(Array.isArray(v))return v.map(convert);if(!v||typeof v!=='object')return v;
 const o=Object.fromEntries(Object.entries(v).filter(([k])=>!['nullable','example','readOnly','writeOnly'].includes(k)).map(([k,x])=>[k,convert(x)]));
 if(o.$ref)o.$ref=o.$ref.replace('#/components/schemas/','#/definitions/');
 for(const [keyword,bound] of [['exclusiveMinimum','minimum'],['exclusiveMaximum','maximum']])if(typeof v[keyword]==='boolean'){delete o[keyword];if(v[keyword])o[keyword]=v[bound];}
 return v.nullable?{anyOf:[o,{type:'null'}]}:o;
}
function sample(s){
 if(s.$ref)return sample(definitions[s.$ref.split('/').pop()]);
 if(s.anyOf)return sample(s.anyOf[0]);if(s.enum)return s.enum[0];
 if(s.type==='object'||s.properties)return Object.fromEntries(Object.entries(s.properties||{}).map(([k,v])=>[k,sample(v)]));
 if(s.type==='array')return [];
 if(s.type==='boolean')return false;if(s.type==='null')return null;
 if(s.type==='integer'||s.type==='number')return s.minimum===undefined?1:s.minimum+(s.exclusiveMinimum?1:0);
 if(s.format==='uuid')return crypto.randomUUID();if(s.format==='date-time')return new Date().toISOString();
 if(s.format==='date')return new Date().toISOString().slice(0,10);if(s.format==='uri')return 'http://localhost/example';
 return 'fixture';
}
function store(initial={}){const m=new Map(Object.entries(initial));return {get:k=>m.get(k),set:(k,v)=>m.set(k,v),unset:k=>m.delete(k),toObject:()=>Object.fromEntries(m)};}
const state=store(Object.fromEntries(collection.variable.map(x=>[x.key,x.value])));
const environment=store(Object.fromEntries(env.values.map(x=>[x.key,x.value])));
environment.set('admin_token','fixture-admin');environment.set('verification_document_ref','fixture-document');
environment.set('max_stage','15');
const fixtures={plans:{},authors:{},published:{},profiles:{},users:{},workouts:[],favorites:new Set(),follow:false,order:null,report:null,review:null,commission:{commission_bps:2000,version:1}};
const R=n=>state.get('state_'+n), J=n=>JSON.parse(R(n));
const clone=x=>JSON.parse(JSON.stringify(x));
function page(items){return {items:clone(items),next_cursor:null};}
function fixture(c,request,schema){
 const op=contracts[c.operation], p=op.path, m=op.method, b=request.body.mode==='raw'?JSON.parse(request.body.raw):null;
 if(c.expected[0]>=400)return {error:{code:'FIXTURE_ERROR',message:'Synthetic negative response'},request_id:crypto.randomUUID()};
 let d=sample(schema);
 const currentUser=c.role==='stranger'?R('stranger_id'):c.role==='trainer'?R('trainer_id'):R('user_id');
 if(p==='/api/v1/auth/register'){const id=Object.keys(fixtures.users).length+1;fixtures.users[id]={id,roles:['user'],blocked:false};return {token:'fixture-token-'+id,user:{id,email:b.email,is_anonymous:false}};}
 if(p==='/api/v2/admin/users')return page(Object.values(fixtures.users));
 if(p==='/api/v2/admin/users/{user_id}'){const id=Number(R(c.ids.user_id));return Object.assign(fixtures.users[id],b);}
 if(p.includes('/exercises')){
  if(p==='/api/v2/admin/exercises'&&m==='POST'){fixtures.exercise={id:10,...b};return fixtures.exercise;}
  if(p==='/api/v2/admin/exercises/{exercise_id}')return Object.assign(fixtures.exercise,b);
  if(!p.startsWith('/api/v2/progress/'))return page([fixtures.exercise]);
 }
 if(p==='/api/v2/equipment')return page([]);
 if(p.includes('/categories')){
  if(m==='POST'){fixtures.category={id:crypto.randomUUID(),...b};return fixtures.category;}
  if(m==='PATCH')return Object.assign(fixtures.category,b);return page([fixtures.category]);
 }
 if(p==='/api/v2/me/profile'){
  fixtures.profiles[currentUser]??={version:1};
  if(m==='PATCH')fixtures.profiles[currentUser]={...fixtures.profiles[currentUser],...b,version:b.version+1};return fixtures.profiles[currentUser];
 }
 if(p==='/api/v2/me/plans'){
  if(m==='POST'){d={id:crypto.randomUUID(),version:1,source_type:'manual',is_active:false,...b};fixtures.plans[d.id]=d;return d;}
  const all=Object.values(fixtures.plans).filter(x=>!x.archived), url=new URL(request.url.raw);
  const start=Number(url.searchParams.get('cursor')||0),limit=Number(url.searchParams.get('limit')||20);
  return {items:clone(all.slice(start,start+limit)),next_cursor:start+limit<all.length?String(start+limit):null};
 }
 if(p.startsWith('/api/v2/me/plans/')){
  const id=R(c.ids.plan_id||'plan_id');d=fixtures.plans[id];assert(d,'Missing plan fixture '+c.name);
  if(m==='PATCH')fixtures.plans[id]=d={...d,...b,version:d.version+1};
  if(p.endsWith('/activate')){Object.values(fixtures.plans).forEach(x=>x.is_active=false);d.is_active=true;d.version++;}
  if(m==='DELETE'){d.archived=true;return null;}return d;
 }
 if(p==='/api/v2/plans/generate'){
  if(!fixtures.generated){const plan=clone(J('plan'));plan.id=crypto.randomUUID();plan.source_type='ai';fixtures.plans[plan.id]=plan;fixtures.generated={plan_id:plan.id,version:plan.version,model_version:'fixture-v1',warnings:[]};}return fixtures.generated;
 }
 if(p==='/api/v2/workouts/complete'){
  const existing=fixtures.workouts.find(x=>x.client_session_id===b.client_session_id);
  if(existing)return existing.result;
  const result={session_id:crypto.randomUUID(),metrics:{volume_kg:1080},highlights:[]};fixtures.workouts.push({...b,id:result.session_id,result});return result;
 }
 if(p==='/api/v2/workouts')return page(fixtures.workouts.map(x=>({...x,result:undefined})));
 if(p==='/api/v2/workouts/{session_id}')return {...fixtures.workouts.find(x=>x.id===R('session_id')),result:undefined};
 if(p==='/api/v2/progress')return {period:'all',time_zone:'Europe/Moscow',workout_count:fixtures.workouts.length,volume_kg:fixtures.workouts.length*1080};
 if(p==='/api/v2/progress/exercises/{exercise_id}')return {exercise_id:10,points:fixtures.workouts.map(x=>({session_id:x.id}))};
 if(p==='/api/v2/progress/summaries')return page([]);
 if(p==='/api/v2/trainer/profile'){
  fixtures.trainer??={id:Number(R('trainer_id')),display_name:'fixture',verified:false};if(m==='PATCH')Object.assign(fixtures.trainer,b);return fixtures.trainer;
 }
 if(p==='/api/v2/trainer/verification'){
  if(m==='POST')fixtures.verification={id:crypto.randomUUID(),status:'pending',...b};return fixtures.verification;
 }
 if(p==='/api/v2/admin/verifications')return page([fixtures.verification]);
 if(p==='/api/v2/admin/verifications/{verification_id}'){Object.assign(fixtures.verification,b);fixtures.trainer.verified=true;return fixtures.verification;}
 if(p==='/api/v2/admin/trainers'||p==='/api/v2/me/trainers')return page(p.includes('/admin/')||fixtures.follow?[fixtures.trainer]:[]);
 if(p==='/api/v2/trainers/{trainer_id}')return fixtures.trainer;
 if(p==='/api/v2/trainer/programs'){
  if(m==='POST'){d={id:crypto.randomUUID(),version:1,status:'draft',published_version_id:null,draft:b};fixtures.authors[d.id]=d;return d;}return page(Object.values(fixtures.authors));
 }
 if(p.includes('/trainer/programs/')||p==='/api/v2/admin/programs/{program_id}/moderation'){
  const id=R(c.ids.program_id||'program_id');const author=fixtures.authors[id];assert(author,'Missing author fixture '+c.name);
  if(p.endsWith('/versions')){author.draft=clone(author.draft);author.version++;author.status='draft';return author;}
  if(p.endsWith('/publish')){author.status='in_review';return {id,status:author.status,version:author.version};}
  if(p.endsWith('/moderation')){author.status='published';author.published_version_id=crypto.randomUUID();fixtures.published[id]={id,...clone(author.draft),trainer:clone(fixtures.trainer),published_version_id:author.published_version_id};return {id,status:'published',version:author.version};}
  if(m==='PATCH'){Object.assign(author.draft,b);author.version++;}return author;
 }
 if(p==='/api/v2/admin/programs')return page(Object.values(fixtures.authors));
 if(p==='/api/v2/catalog/programs')return page(Object.values(fixtures.published).map(x=>({...x,equipment:[]})));
 if(p==='/api/v2/catalog/programs/{program_id}'){const id=R(c.ids.program_id);return {...fixtures.published[id],equipment:[],rating:null,review_count:0,access:fixtures.access&&id===R('free_program_id')?'free':'preview'};}
 if(p==='/api/v2/me/favorites/{program_id}'){if(m==='PUT')fixtures.favorites.add(R(c.ids.program_id));else fixtures.favorites.delete(R(c.ids.program_id));return m==='DELETE'?null:{ok:true};}
 if(p==='/api/v2/me/favorites')return page([...fixtures.favorites].map(id=>({...fixtures.published[id],equipment:[]})));
 if(p==='/api/v2/programs/{program_id}/access'){fixtures.access=true;return {id:crypto.randomUUID(),program_id:R(c.ids.program_id),status:'active'};}
 if(p==='/api/v2/admin/settings/commission'){if(m==='PATCH')fixtures.commission={...b,version:b.version+1};return fixtures.commission;}
 if(p==='/api/v2/orders'){fixtures.order??={id:crypto.randomUUID(),status:'pending',amount_minor:100000,currency:'RUB',checkout_url:null};return fixtures.order;}
 if(p==='/api/v2/orders/{order_id}')return fixtures.order;
 if(p==='/api/v2/me/orders'||p==='/api/v2/admin/orders')return page([fixtures.order]);
 if(p.includes('/payouts'))return page([]);
 if(p==='/api/v2/programs/{program_id}/fork'){
  if(!fixtures.forkResult){const plan={id:crypto.randomUUID(),version:1,title:b.title,source_type:'fork',source_program_version_id:b.source_version_id,days:clone(fixtures.published[R(c.ids.program_id)].days),schedule:{time_zone:'Europe/Moscow',start_date:new Date().toISOString().slice(0,10),entries:[]},is_active:false};plan.days.forEach(x=>x.id=crypto.randomUUID());fixtures.plans[plan.id]=plan;fixtures.forkResult={plan_id:plan.id,source_program_version_id:b.source_version_id,version:1};}return fixtures.forkResult;
 }
 if(p==='/api/v2/shares'){fixtures.share={id:crypto.randomUUID(),url:'http://localhost/share/'+crypto.randomUUID(),expires_at:null,visibility:'public',plan:clone(fixtures.plans[b.target_id])};return fixtures.share;}
 if(p==='/api/v2/shares/{token}')return {id:fixtures.share.id,title:fixtures.share.plan.title,days:fixtures.share.plan.days,target_type:'plan',requires_access:false,can_fork:true};
 if(p==='/api/v2/shares/{token}/save')return {library_item_id:crypto.randomUUID()};
 if(p==='/api/v2/shares/{token}/fork'){const plan=clone(fixtures.share.plan);plan.id=crypto.randomUUID();fixtures.plans[plan.id]=plan;return {plan_id:plan.id,version:1};}
 if(p==='/api/v2/shares/{share_id}')return null;
 if(p==='/api/v2/programs/{program_id}/review-eligibility'){const count=fixtures.workouts.filter(x=>x.plan_id===R('fork_id')).length;return {eligible:count>=3,completed_workouts:count,required_workouts:3};}
 if(p==='/api/v2/programs/{program_id}/reviews'){
  if(m==='POST'){fixtures.review={id:crypto.randomUUID(),program_id:R(c.ids.program_id),status:'published',...b};return fixtures.review;}
  const items=fixtures.review&&!(c.role==='stranger'&&fixtures.review.status==='hidden')?[fixtures.review]:[];return {...page(items),count:items.length};
 }
 if(p==='/api/v2/reviews/{review_id}'||p==='/api/v2/admin/reviews/{review_id}/moderation'){if(m==='DELETE'){fixtures.review=null;return null;}return Object.assign(fixtures.review,b);}
 if(p==='/api/v2/admin/reviews')return {...page([fixtures.review]),count:1};
 if(p==='/api/v2/reports'){fixtures.report={id:crypto.randomUUID(),status:'open',...b};return fixtures.report;}
 if(p==='/api/v2/admin/reports')return page([fixtures.report]);
 if(p==='/api/v2/admin/reports/{report_id}')return Object.assign(fixtures.report,b);
 if(p==='/api/v2/trainers/{trainer_id}/follow'){fixtures.follow=m==='PUT';return m==='DELETE'?null:{trainer_id:Number(R('trainer_id')),following:true};}
 if(p==='/api/v2/notifications')return page([]);
 if(p==='/api/v2/recommendations/programs')return {items:[],model_version:'fixture-v1',fallback:true,next_cursor:null};
 if(p==='/api/v2/dashboard')return {state:'just_completed',blocks:[],generated_at:new Date().toISOString()};
 if(schema.$ref==='#/definitions/OK')return {ok:true};
 return d;
}

let assertions=0, negatives=0, checkedBodies=0, killedStatus=0, killedSchema=0, killedBusiness=0, skipped=0;
const common=collection.event[0].script.exec.join('\n');
for(let index=0;index<flat.length;index++){
 const item=flat[index],c=cases[index],local=store(), failures=[];let skip=false;
 const replace=text=>String(text).replace(/\{\{([^}]+)\}\}/g,(_,k)=>k==='$guid'?crypto.randomUUID():String(local.get(k)??environment.get(k)??state.get(k)??'{{'+k+'}}'));
 const request=clone(item.request);request.body??={};request.headers={upsert:()=>{}};
 const pm={collectionVariables:state,environment,variables:{...local,replaceIn:replace},info:{requestName:item.name},request,
 execution:{skipRequest:()=>{skip=true;}},expect:chai.expect,test:(n,fn)=>{assertions++;try{fn();}catch(e){failures.push(n+': '+e.message);}}};
 const context={pm,console:{warn:()=>{}},Date,JSON,Error,Object,Number,String,decodeURIComponent};
 vm.runInNewContext(item.event[0].script.exec.join('\n'),context,{timeout:2000});
 assert(!failures.length,'Prerequest '+c.name+' '+failures.join(';'));
 if(skip){assert(c.gate,'Unexpected skip '+c.name);skipped++;continue;}
 request.url.raw=replace(request.url.raw);assert(!/\{\{/.test(request.url.raw),'Unresolved URL '+c.name);
 for(const h of request.header)h.value=replace(h.value);
 if(request.body.mode==='raw')request.body.raw=replace(request.body.raw);
 const op=spec.paths[contracts[c.operation].path][contracts[c.operation].method.toLowerCase()];
 for(const h of (op.parameters||[]).filter(x=>x.in==='header'&&x.required))assert(request.header.some(x=>x.key===h.name&&x.value),'Missing required header '+c.name);
 if(request.body.mode==='raw'){
  const schema=op.requestBody.content['application/json'].schema;
  const validate=ajv.compile({...convert(schema),definitions,$schema:'http://json-schema.org/draft-07/schema#'});
  const valid=validate(JSON.parse(request.body.raw));
  const intentionallyInvalid=/Недопустимый день|ноль подходов|Нулевой вес/.test(c.name);
  assert(intentionallyInvalid?!valid:valid,'Request schema '+c.name+' '+JSON.stringify(validate.errors));checkedBodies++;
  const input=JSON.parse(request.body.raw);
  if(input.days&&input.schedule)input.schedule.entries.forEach(entry=>assert(input.days.some(d=>d.id===entry.plan_day_id),'Schedule refers to another plan day '+c.name));
  if(input.client_session_id){
    assert(Date.parse(input.finished_at)>Date.parse(input.started_at),'Workout must finish after its start '+c.name);
    const ready=R(input.plan_id===R('fork_id')?'fork_ready_at':'plan_ready_at');
    assert(Date.parse(input.started_at)>=Date.parse(ready),'Workout predates the plan snapshot '+c.name);
  }
 }
 if(c.expected[0]>=400)negatives++;
 const responseSchema=contracts[c.operation].responses[String(c.expected[0])]?.['application/json'];
 const data=fixture(c,request,responseSchema||{});
 const text=data===null?'':JSON.stringify(data);
 function run(code,responseData,responseText){
  const errs=[];const p={...pm,test:(n,fn)=>{assertions++;try{fn();}catch(e){errs.push(n+': '+e.message);}},response:{code,text:()=>responseText,json:()=>responseData,headers:{get:()=> 'application/json'},to:{have:{jsonSchema:schema=>{const ok=ajv.validate(schema,responseData);chai.expect(ok,JSON.stringify(ajv.errors)).to.eql(true);}}}}};
  const ctx={...context,pm:p};vm.runInNewContext(common,ctx,{timeout:2000});vm.runInNewContext(item.event[1].script.exec.join('\n'),ctx,{timeout:2000});return errs;
 }
 const errs=run(c.expected[0],data,text);assert(!errs.length,'Response fixture '+c.name+' '+errs.join(';'));
 function probe(code,body,text){const backup=state.toObject();try{return run(code,body,text);}finally{Object.keys(state.toObject()).forEach(k=>state.unset(k));Object.entries(backup).forEach(([k,v])=>state.set(k,v));}}
 assert(probe(500,{error:'wrong'},'{"error":"wrong"}').length,'500 falsely passes '+c.name);killedStatus++;
 if(responseSchema){const bad=42;assert(probe(c.expected[0],bad,'42').length,'Invalid schema falsely passes '+c.name);killedSchema++;}
 // Mutate all specifically asserted primitive response fields independently.
 if(data&&c.test.includes('1080')){const mutant=data.metrics?{...data,metrics:{volume_kg:999}}:{...data,volume_kg:999};assert(probe(c.expected[0],mutant,JSON.stringify(mutant)).length,'Volume assertion ineffective '+c.name);}
 const businessMutations={
  'Профиль действительно сохранён':x=>x.weight_kg=90,
  'Первый план стал неактивным':x=>x.is_active=true,
  'Проверить сохранённые изменения':x=>x.title='wrong title',
  'В избранном одна запись':x=>x.items.push(clone(x.items[0])),
  'Оригинал не изменился':x=>x.days[0].exercises[0].sets=99,
  'Публичный проверенный профиль':x=>x.verified=false,
  'Создать заказ с фиксацией цены':x=>x.amount_minor=1,
  'До использования программы отзыв недоступен':x=>x.eligible=true,
  'Право на отзыв после 2 занятий':x=>x.eligible=true,
  'Право на отзыв после 3 занятий':x=>x.eligible=false,
  'Редактировать свой отзыв':x=>x.overall=1,
  'Закрыть жалобу':x=>x.status='open',
  'Мои тренеры без дублей':x=>x.items.push(clone(x.items[0]))
 };
 const key=Object.keys(businessMutations).find(k=>c.name.endsWith(k));
 if(key){const mutant=clone(data);businessMutations[key](mutant);assert(probe(c.expected[0],mutant,JSON.stringify(mutant)).length,'Business mutation falsely passes '+c.name);killedBusiness++;}
}
assert(skipped===4,'Exactly four explicit integration fixtures should be disabled');
const report={offline_only:true,official_collection_schema:schemaFile?'PASS':'NOT_CHECKED',compiled_scripts:1+flat.length*2,requests:flat.length,executed_synthetic_cases:flat.length-skipped,skipped_integration_templates:skipped,
 validated_request_bodies:checkedBodies,negative_cases:negatives,assertion_executions:assertions,rejected_500_mutations:killedStatus,rejected_schema_mutations:killedSchema,rejected_business_mutations:killedBusiness};
fs.writeFileSync(path.join(__dirname,'verification.json'),JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify(report));

// Real Postman sandbox smoke check: bootstrap plus disabled later stages, only an in-memory localhost fixture server.
const http=require('http');let registered=0;
const server=http.createServer((req,res)=>{
 assert(req.url==='/api/v1/auth/register','Unexpected Newman smoke request '+req.url);
 let raw='';req.on('data',b=>raw+=b);req.on('end',()=>{const input=JSON.parse(raw);registered++;
 res.writeHead(200,{'Content-Type':'application/json'});res.end(JSON.stringify({token:'sandbox-fixture-token',user:{id:registered,email:input.email,is_anonymous:false}}));});
});
server.listen(0,'127.0.0.1',()=>{
 const smokeEnv=clone(env);smokeEnv.values.find(x=>x.key==='base_url').value='http://127.0.0.1:'+server.address().port;
 smokeEnv.values.find(x=>x.key==='max_stage').value='0';
 newman.run({collection,environment:smokeEnv,reporters:[],timeoutRequest:5000},(error,summary)=>{
  server.close();if(error||summary.run.failures.length){console.error(error||summary.run.failures);process.exitCode=1;return;}
  assert(registered===3,'Newman should register exactly three fixture accounts');
  report.newman_sandbox_bootstrap='PASS';report.newman_assertions=summary.run.stats.assertions.total;
  fs.writeFileSync(path.join(__dirname,'verification.json'),JSON.stringify(report,null,2)+'\n');console.log('Newman sandbox: PASS (3 bootstrap requests; later stages skipped)');
 });
});
