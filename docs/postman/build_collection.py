"""Build reproducible Postman scenarios from the team's OpenAPI contract."""
import copy
import hashlib
import json
import re
import uuid
from pathlib import Path

HERE = Path(__file__).resolve().parent
SPEC = json.loads((HERE.parent / 'api/openapi.json').read_text(encoding='utf-8'))
CASES = []
FOLDERS = {0: {'name': '00 · Подготовка нового прогона', 'item': []}}
for tag in SPEC['tags']:
    if re.match(r'^\d{2} · ', tag['name']):
        FOLDERS[int(tag['name'][:2])] = {'name': tag['name'], 'description': tag.get('description', ''), 'item': []}
FOLDERS[16] = {'name': '16 · Интеграции — явные тестовые фикстуры', 'description': 'Отключены по умолчанию. Для провайдера нужны согласованный адаптер подписи и тестовый баланс; для уведомления — доставленное событие. См. README.', 'item': []}

def convert(value):
    if isinstance(value, list): return [convert(v) for v in value]
    if not isinstance(value, dict): return value
    out = {k: convert(v) for k, v in value.items() if k not in ('nullable', 'example', 'readOnly', 'writeOnly')}
    if '$ref' in out: out['$ref'] = out['$ref'].replace('#/components/schemas/', '#/definitions/')
    # Postman's Ajv validates draft-07; OpenAPI 3.0 uses boolean exclusive bounds.
    for keyword, bound in [('exclusiveMinimum', 'minimum'), ('exclusiveMaximum', 'maximum')]:
        if isinstance(value.get(keyword), bool):
            out.pop(keyword, None)
            if value[keyword]: out[keyword] = value[bound]
    if value.get('nullable'):
        out = {'anyOf': [out, {'type': 'null'}]}
    return out

DEFINITIONS = convert(SPEC['components']['schemas'])
CONTRACTS = {}
for path, item in SPEC['paths'].items():
    for method, op in item.items():
        if not isinstance(op, dict) or 'operationId' not in op: continue
        CONTRACTS[op['operationId']] = {
            'path': path, 'method': method.upper(),
            'responses': {code: {media: convert(entry['schema']) for media, entry in response.get('content', {}).items()}
                          for code, response in op['responses'].items()}
        }

COMMON_TEST = r'''
(function () {
  const c = JSON.parse(pm.variables.get('__case'));
  pm.test('HTTP: ожидаемый статус ' + c.codes.join('/'), () => pm.expect(pm.response.code).to.be.oneOf(c.codes));
  const contract = JSON.parse(pm.collectionVariables.get('__contracts'))[c.operation];
  const responses = contract.responses;
  pm.test('Статус описан в OpenAPI', () => pm.expect(Object.keys(responses)).to.include(String(pm.response.code)));
  const content = responses[String(pm.response.code)];
  if (pm.response.code === 204) {
    pm.test('204: тело отсутствует', () => pm.expect(pm.response.text()).to.eql(''));
  } else if (content && content['application/json']) {
    pm.test('Content-Type: application/json', () => pm.expect(pm.response.headers.get('Content-Type') || '').to.match(/application\/json\b/i));
    pm.test('Тело соответствует OpenAPI', () => pm.response.to.have.jsonSchema(Object.assign({}, content['application/json'], {
      $schema: 'http://json-schema.org/draft-07/schema#', definitions: JSON.parse(pm.collectionVariables.get('__definitions'))
    })));
    if (pm.response.code < 300 && content['application/json'].$ref === '#/definitions/OK') {
      pm.test('Операция подтверждена: ok=true', () => pm.expect(pm.response.json().ok).to.eql(true));
    }
  }
})();
'''

PRE_HELPERS = r'''
const S = pm.collectionVariables;
function R(name) {
  const value = S.get('state_' + name);
  if (value === undefined || value === null || value === '') throw Error('Нет state_' + name + ': выполните предыдущие этапы');
  return value;
}
function J(name) { return JSON.parse(R(name)); }
function save(name, value) { S.set('state_' + name, typeof value === 'object' ? JSON.stringify(value) : value); }
function guid() { return pm.variables.replaceIn('{{$guid}}'); }
function day(offset=0) { return new Date(Date.now()+offset*86400000).toISOString().slice(0,10); }
function planInput(title) {
  return {title, days:[{id:guid(),title:'День A',position:0,exercises:[{exercise_id:Number(R('exercise_id')),position:0,sets:2,reps_min:8,reps_max:10,weight_kg:60,rest_sec:90}]}],
    schedule:{time_zone:'Europe/Moscow',start_date:day(),entries:[]}};
}
function draft(price=0) {
  return {title:'Postman '+R('run_id')+(price?' paid':' free'),description:'Тестовая программа',goal:'muscle_gain',level:'beginner',duration_weeks:4,sessions_per_week:3,
    equipment_ids:[],category_ids:[R('category_id')],price_minor:price,currency:'RUB',allow_public_remix:true,days:planInput('x').days};
}
function workout(plan) {
  if (!plan.days[0].id) throw Error('GET плана должен возвращать ID дня для записи тренировки');
  const ready=new Date(R(plan.id===S.get('state_fork_id')?'fork_ready_at':'plan_ready_at')).getTime();
  const last=S.get('state_last_workout_finished');
  const start=new Date(Math.max(ready,last?new Date(last).getTime()+1:ready));
  const finish=new Date(Math.max(Date.now(),start.getTime()+1));
  save('last_workout_finished',finish.toISOString());
  return {client_session_id:guid(),plan_id:plan.id,plan_version:plan.version,plan_day_id:plan.days[0].id,
    started_at:start.toISOString(),finished_at:finish.toISOString(),
    slots:[{exercise_id:Number(R('exercise_id')),position:0,sets:[{reps:10,weight_kg:20,is_warmup:true},{reps:10,weight_kg:60,is_warmup:false},{reps:8,weight_kg:60,is_warmup:false}]}]};
}
'''
POST_HELPERS = r'''
const S=pm.collectionVariables;
function R(n){const v=S.get('state_'+n);if(v===undefined)throw Error('Нет state_'+n);return v;}
function J(n){return JSON.parse(R(n));}
function save(n,v){S.set('state_'+n,typeof v==='object'?JSON.stringify(v):v);}
function check(name, fn){pm.test(name,fn);}
function eq(a,b){pm.expect(a).to.eql(b);}
function present(v){pm.expect(v).not.to.be.oneOf([null,undefined,'']);}
function includesItem(items,id){pm.expect(items.some(x=>String(x.id)===String(id))).to.eql(true);}
let data=null;
if(pm.response.code!==204){try{data=pm.response.json();}catch(e){check('Ответ должен быть JSON',()=>{throw e;});return;}}
const submitted=pm.request.body && pm.request.body.mode==='raw' ? JSON.parse(pm.variables.replaceIn(pm.request.body.raw)) : null;
'''

def add(stage, name, method, path, *, role='user', codes=None, body=None, pre='', test='', query=None, ids=None, idem=None, gate=None, capture=None, description=''):
    op = SPEC['paths'][path][method.lower()]
    if codes is None: codes = [int(c) for c in op['responses'] if c.startswith('2')]
    assert all(str(c) in op['responses'] for c in codes), (name, codes)
    ids = ids or {}
    resolved = path
    dependencies = []
    for param in re.findall(r'\{([^}]+)\}', path):
        var = ids.get(param, param)
        resolved = resolved.replace('{'+param+'}', '{{state_'+var+'}}')
        dependencies.append(var)
    headers = [{'key':'Accept','value':'application/json'}]
    if body is not None: headers.append({'key':'Content-Type','value':'application/json'})
    if idem:
        headers.append({'key':'Idempotency-Key','value':'{{state_'+idem+'}}'})
    request = {'method': method.upper(), 'header': headers, 'url': {'raw':'{{base_url}}'+resolved, 'host':['{{base_url}}'], 'path':resolved.strip('/').split('/')},
               'description':description or op['summary'], 'auth':{'type':'noauth'} if role is None else {'type':'bearer','bearer':[{'key':'token','value':'{{__token}}','type':'string'}]}}
    if query:
        request['url']['query'] = [{'key':k,'value':str(v)} for k,v in query.items()]
        request['url']['raw'] += '?'+'&'.join(k+'='+str(v) for k,v in query.items())
    if body is not None: request['body'] = {'mode':'raw','raw':'{{__body}}','options':{'raw':{'language':'json'}}}
    config={'operation':op['operationId'],'codes':codes}
    script = '(function(){\n'+PRE_HELPERS+'\n'
    script += "try {\nconst last=Number(pm.environment.get('max_stage')||15);if(!Number.isInteger(last)||last<0||last>15)throw Error('max_stage должен быть целым числом 0–15');\n"
    script += f"if ({stage} > Number(pm.environment.get('max_stage') || 15) && {stage} !== 16) {{ console.warn('Этап отключён: '+pm.info.requestName); pm.execution.skipRequest(); return; }}\n"
    if gate:
        script += f"if(pm.environment.get({json.dumps(gate)})!=='true'){{console.warn('ИНТЕГРАЦИЯ НЕ ПРОВЕРЕНА: '+pm.info.requestName);pm.execution.skipRequest();return;}}\n"
    script += "const base=pm.environment.get('base_url');if(!base||!/^https?:\\/\\//.test(base))throw Error('Укажите base_url');pm.variables.set('base_url',base.replace(/\\/+$/,''));\n"
    script += "if(pm.environment.get('test_environment')!=='true')throw Error('Подтвердите тестовую БД: test_environment=true');\n"
    if role == 'admin': script += "const token=pm.environment.get('admin_token');if(!token)throw Error('Нужен admin_token тестового администратора');pm.variables.set('__token',token);\n"
    elif role: script += f"pm.variables.set('__token',R('{role}_token'));\n"
    # Prepare fresh data before resolving path/body dependencies.
    script += pre+'\n'
    for var in dependencies: script += f"pm.variables.set('state_{var}',R('{var}'));\n"
    if idem: script += f"pm.variables.set('state_{idem}',R('{idem}'));\n"
    if query:
        for value in query.values():
            for var in re.findall(r'\{\{state_([^}]+)\}\}',str(value)): script += f"pm.variables.set('state_{var}',R('{var}'));\n"
    if body is not None: script += 'pm.variables.set(\'__body\',JSON.stringify('+body+'));\n'
    script += 'pm.variables.set(\'__case\','+json.dumps(json.dumps(config))+');\n'
    script += "}catch(e){pm.test('Подготовка запроса: '+e.message,()=>{throw e;});pm.execution.skipRequest();}\n})();"
    post='(function(){\nif(!'+json.dumps(codes)+'.includes(pm.response.code))return;\n'+POST_HELPERS+'\nif('+json.dumps(codes)+'.includes(pm.response.code)){\n'+test+'\n}\n})();'
    if capture:
        # Captures are specific success fields, not arbitrary UUID placeholders.
        post=post.replace(test+'\n}',test+'\n'+capture+'\n}')
    case={'name':f'{len(FOLDERS[stage]["item"])+1:02d} · {name}', 'id':str(uuid.uuid5(uuid.NAMESPACE_URL,f'postman:{stage}:{name}')),
          'request':request,'event':[{'listen':'prerequest','script':{'type':'text/javascript','exec':script.splitlines()}},{'listen':'test','script':{'type':'text/javascript','exec':post.splitlines()}}]}
    FOLDERS[stage]['item'].append(case)
    CASES.append({'stage':stage,'name':case['name'],'operation':op['operationId'],'expected':codes,'gate':gate,'body_expression':body,'test':test,'pre':pre,'ids':ids,'idem':idem,'role':role})
    return case

def get(stage,name,path,**kw): return add(stage,name,'GET',path,**kw)
def post(stage,name,path,**kw): return add(stage,name,'POST',path,**kw)
def patch(stage,name,path,**kw): return add(stage,name,'PATCH',path,**kw)
def delete(stage,name,path,**kw): return add(stage,name,'DELETE',path,**kw)
def put(stage,name,path,**kw): return add(stage,name,'PUT',path,**kw)
def store(name,field='id'): return f"check('Сохраняем {name}',()=>{{present(data.{field});save('{name}',data.{field});}});"
def snapshot(name):
    result=f"save('{name}',data);"
    if name in ('plan','fork'): result+=f"save('{name}_ready_at',new Date().toISOString());"
    return result

# 00: one administrator token is supplied; all ordinary test accounts are new each run.
for index,role in enumerate(['user','stranger','trainer']):
    reset="Object.keys(S.toObject()).filter(k=>k.startsWith('state_')).forEach(k=>S.unset(k));save('run_id',guid());" if index==0 else ''
    post(0,f'Создать аккаунт {role}','/api/v1/auth/register',role=None,
         pre=reset+f"save('{role}_email','pm-{role}-'+R('run_id')+'@example.test');save('{role}_password',guid()+'Aa1!');",
         body=f"{{email:R('{role}_email'),password:R('{role}_password')}}",
         test=f"check('Новый аккаунт и токен',()=>{{present(data.token);present(data.user.id);save('{role}_token',data.token);save('{role}_id',data.user.id);}});")

# 01: roles and denial are checked against a fresh non-admin account.
get(1,'Администратор видит пользователей','/api/v2/admin/users',role='admin',query={'limit':100})
get(1,'Обычный пользователь не видит админку','/api/v2/admin/users',codes=[403])
get(1,'Без JWT нет доступа к профилю','/api/v2/me/profile',role=None,codes=[401])
patch(1,'Назначить созданному автору роль тренера','/api/v2/admin/users/{user_id}',role='admin',ids={'user_id':'trainer_id'},body="{roles:['user','trainer'],reason:'Postman fixture'}",test="check('Роль trainer сохранена',()=>pm.expect(data.roles).to.include('trainer'));")
patch(1,'Заблокировать постороннего пользователя','/api/v2/admin/users/{user_id}',role='admin',ids={'user_id':'stranger_id'},body="{blocked:true,reason:'Postman block test'}",test="check('blocked=true',()=>eq(data.blocked,true));")
get(1,'Заблокированный токен отвергается','/api/v2/me/profile',role='stranger',codes=[401,403])
patch(1,'Разблокировать постороннего пользователя','/api/v2/admin/users/{user_id}',role='admin',ids={'user_id':'stranger_id'},body="{blocked:false,reason:'Postman unblock'}",test="check('blocked=false',()=>eq(data.blocked,false));")

# 02: own reference records avoid relying on production fixtures or magic IDs.
post(2,'Создать упражнение без оборудования','/api/v2/admin/exercises',role='admin',body="{name:'Postman '+R('run_id'),muscle_group:'chest',equipment_ids:[],is_active:true}",test=store('exercise_id'))
get(2,'Справочник упражнений администратора','/api/v2/admin/exercises',role='admin')
patch(2,'Изменить созданное упражнение','/api/v2/admin/exercises/{exercise_id}',role='admin',body="{name:'Postman updated '+R('run_id'),equipment_ids:[],is_active:true}",test="check('Имя упражнения обновлено',()=>eq(data.name,submitted.name));")
get(2,'Поиск созданного упражнения','/api/v2/exercises',query={'query':'{{state_run_id}}','limit':100},test="check('Поиск возвращает созданное упражнение',()=>includesItem(data.items,R('exercise_id')));")
get(2,'Оборудование','/api/v2/equipment')
post(2,'Создать категорию','/api/v2/admin/categories',role='admin',body="{slug:'pm-'+R('run_id'),name:'Postman category',position:0,is_active:true}",test=store('category_id'))
get(2,'Категории администратора','/api/v2/admin/categories',role='admin')
patch(2,'Переименовать категорию','/api/v2/admin/categories/{category_id}',role='admin',body="{slug:'pm-'+R('run_id'),name:'Postman updated',position:0,is_active:true}",test="check('Имя категории обновлено',()=>eq(data.name,submitted.name));")
get(2,'Категории каталога','/api/v2/catalog/categories')
get(2,'Недопустимая пагинация','/api/v2/exercises',query={'limit':0},codes=[400,422])

# 03: optimistic concurrency, fractional body weight and persistence.
get(3,'Прочитать версию профиля','/api/v2/me/profile',test=store('profile_version','version'))
patch(3,'Заполнить профиль','/api/v2/me/profile',body="{version:Number(R('profile_version')),age:25,height_cm:180,weight_kg:80.5,gender:'male',activity_level:'moderate',goal:'muscle_gain',level:'beginner',training_days:[1,3,5],equipment_ids:[],session_duration_min:60,excluded_exercise_ids:[]}",
      pre="save('old_profile_version',R('profile_version'));",test="check('Версия выросла',()=>pm.expect(data.version).to.be.above(Number(R('old_profile_version'))));check('Вес дробный',()=>eq(data.weight_kg,80.5));"+store('profile_version','version'))
get(3,'Профиль действительно сохранён','/api/v2/me/profile',test="check('Сохранённые параметры',()=>{eq(data.weight_kg,80.5);eq(data.training_days,[1,3,5]);eq(data.version,Number(R('profile_version')));});")
patch(3,'Устаревшая версия профиля','/api/v2/me/profile',body="{version:Number(R('old_profile_version')),weight_kg:90}",codes=[409])
patch(3,'Недопустимый день недели','/api/v2/me/profile',body="{version:Number(R('profile_version')),training_days:[8]}",codes=[400,422])

# 04: aggregate edits, ownership, concurrency, single active plan, archive.
post(4,'Создать личный план','/api/v2/me/plans',body="planInput('Postman plan '+R('run_id'))",test=store('plan_id')+snapshot('plan'))
get(4,'Прочитать созданный план','/api/v2/me/plans/{plan_id}',test="check('ID, дни и упражнения',()=>{eq(data.id,R('plan_id'));eq(data.days[0].exercises[0].exercise_id,Number(R('exercise_id')));present(data.days[0].id);});"+snapshot('plan'))
get(4,'Чужой пользователь не читает план','/api/v2/me/plans/{plan_id}',role='stranger',codes=[403,404])
patch(4,'Чужой пользователь не меняет план','/api/v2/me/plans/{plan_id}',role='stranger',body="{version:J('plan').version,title:'hacked',days:J('plan').days,schedule:J('plan').schedule}",codes=[403,404])
patch(4,'Редактировать план и расписание','/api/v2/me/plans/{plan_id}',pre="save('old_plan_version',J('plan').version);",body="{version:J('plan').version,title:'Postman edited',days:J('plan').days,schedule:{time_zone:'Europe/Moscow',start_date:day(),entries:[{weekday:1,plan_day_id:J('plan').days[0].id}]}}",test="check('Название, расписание и новая версия',()=>{eq(data.title,'Postman edited');pm.expect(data.version).to.be.above(Number(R('old_plan_version')));});"+snapshot('plan'))
get(4,'Проверить сохранённые изменения','/api/v2/me/plans/{plan_id}',test="check('Изменения сохранились',()=>{eq(data.title,'Postman edited');eq(data.schedule.entries[0].plan_day_id,data.days[0].id);});"+snapshot('plan'))
patch(4,'Устаревшая версия плана','/api/v2/me/plans/{plan_id}',body="{version:Number(R('old_plan_version')),title:'stale',days:J('plan').days,schedule:J('plan').schedule}",codes=[409])
patch(4,'Нельзя задать ноль подходов','/api/v2/me/plans/{plan_id}',pre="const invalid=J('plan');invalid.days[0].exercises[0].sets=0;save('invalid_plan',invalid);",body="{version:J('plan').version,title:'invalid',days:J('invalid_plan').days,schedule:J('plan').schedule}",codes=[400,422])
post(4,'Активировать первый план','/api/v2/me/plans/{plan_id}/activate',body="{version:J('plan').version}",test="check('План активен',()=>eq(data.is_active,true));"+snapshot('plan'))
post(4,'Создать второй план','/api/v2/me/plans',body="planInput('Postman spare')",test=store('spare_id')+snapshot('spare'))
post(4,'Активировать второй план','/api/v2/me/plans/{plan_id}/activate',ids={'plan_id':'spare_id'},body="{version:J('spare').version}",test="check('Второй план активен',()=>eq(data.is_active,true));"+snapshot('spare'))
get(4,'Первый план стал неактивным','/api/v2/me/plans/{plan_id}',test="check('Первый план деактивирован',()=>eq(data.is_active,false));"+snapshot('plan'))
get(4,'Список личных планов','/api/v2/me/plans',query={'limit':100},test="check('Оба плана присутствуют',()=>{includesItem(data.items,R('plan_id'));includesItem(data.items,R('spare_id'));});")
get(4,'Первая страница личных планов','/api/v2/me/plans',query={'limit':1},test="check('Одна запись и курсор следующей страницы',()=>{eq(data.items.length,1);present(data.next_cursor);save('plans_cursor',data.next_cursor);save('first_page_plan_id',data.items[0].id);});")
get(4,'Следующая страница без повторов','/api/v2/me/plans',query={'limit':1,'cursor':'{{state_plans_cursor}}'},test="check('Курсор не повторяет первую страницу',()=>{eq(data.items.length,1);pm.expect(data.items[0].id).not.to.eql(R('first_page_plan_id'));pm.expect([R('plan_id'),R('spare_id')]).to.include(data.items[0].id);});")
delete(4,'Архивировать второй план','/api/v2/me/plans/{plan_id}',ids={'plan_id':'spare_id'})

# 05: real ML/stub must be configured by the backend; no invented exact model output.
post(5,'Сгенерировать AI-план','/api/v2/plans/generate',pre="save('generate_key',guid());",idem='generate_key',body="{profile_version:Number(R('profile_version'))}",test=store('ai_plan_id','plan_id')+snapshot('generated'))
post(5,'Повтор генерации не создаёт второй план','/api/v2/plans/generate',idem='generate_key',body="{profile_version:Number(R('profile_version'))}",test="check('Тот же результат генерации',()=>eq(data,J('generated')));")
get(5,'AI-план сохранён и редактируем','/api/v2/me/plans/{plan_id}',ids={'plan_id':'ai_plan_id'},test="check('Источник AI и непустая программа',()=>{eq(data.source_type,'ai');pm.expect(data.days.length).to.be.above(0);data.days.forEach(d=>pm.expect(d.exercises.length).to.be.above(0));});")

# 06: exact volume, replay and payload-conflict, history and progress delta.
get(6,'Базовый прогресс до занятия','/api/v2/progress',query={'period':'all','time_zone':'Europe/Moscow'},test=snapshot('progress_before'))
post(6,'Завершить тренировку','/api/v2/workouts/complete',pre="save('workout_key',guid());save('workout_body',workout(J('plan')));",idem='workout_key',body="J('workout_body')",test="check('Объём без разминки = 1080 кг',()=>eq(data.metrics.volume_kg,1080));"+store('session_id','session_id')+snapshot('workout_result'))
post(6,'Повтор занятия возвращает исходный результат','/api/v2/workouts/complete',idem='workout_key',body="J('workout_body')",test="check('Нет дублирования',()=>eq(data,J('workout_result')));")
post(6,'Тот же client_session_id с новым ключом не дублируется','/api/v2/workouts/complete',pre="save('workout_alternate_key',guid());",idem='workout_alternate_key',body="J('workout_body')",codes=[201,409],test="if(pm.response.code===201)check('Сессия уже существует',()=>eq(data.session_id,R('session_id')));")
post(6,'Тот же ключ с другим телом — конфликт','/api/v2/workouts/complete',idem='workout_key',pre="const b=J('workout_body');b.slots[0].sets[1].weight_kg=65;save('conflicting_workout',b);",body="J('conflicting_workout')",codes=[409])
get(6,'История тренировок','/api/v2/workouts',query={'limit':100},test="check('Занятие ровно одно',()=>eq(data.items.filter(x=>x.id===R('session_id')).length,1));")
get(6,'Подходы завершённого занятия','/api/v2/workouts/{session_id}',test="check('Подходы не потеряны',()=>{eq(data.id,R('session_id'));eq(data.slots,J('workout_body').slots);});")
get(6,'Чужую тренировку нельзя прочитать','/api/v2/workouts/{session_id}',role='stranger',codes=[403,404])
post(6,'Записать дробный вес тела','/api/v2/progress/weight',body="{day:day(),weight_kg:80.5}")
post(6,'Нулевой вес тела запрещён','/api/v2/progress/weight',body="{day:day(),weight_kg:0}",codes=[400,422])
get(6,'Прогресс после повторного запроса','/api/v2/progress',query={'period':'all','time_zone':'Europe/Moscow'},test="check('Повтор не увеличил счётчик и объём',()=>{eq(data.workout_count,J('progress_before').workout_count+1);eq(data.volume_kg,J('progress_before').volume_kg+1080);});")
get(6,'Динамика упражнения','/api/v2/progress/exercises/{exercise_id}',test="check('Точка занятия присутствует',()=>pm.expect(data.points.some(x=>x.session_id===R('session_id'))).to.eql(true));")
get(6,'Недельные сводки','/api/v2/progress/summaries',query={'period':'week'})
get(6,'Месячные сводки','/api/v2/progress/summaries',query={'period':'month'})

# 07: end-to-end verification with a supplied storage reference.
patch(7,'Заполнить профиль тренера','/api/v2/trainer/profile',role='trainer',body="{display_name:'Postman trainer '+R('run_id'),bio:'Тестовый автор',specializations:['strength'],experience_years:3,qualifications:'Тестовая квалификация'}",test="check('Профиль автора',()=>eq(data.display_name,submitted.display_name));")
get(7,'Профиль тренера','/api/v2/trainer/profile',role='trainer',test="check('ID автора',()=>eq(data.id,Number(R('trainer_id'))));")
post(7,'Подать заявку на проверку','/api/v2/trainer/verification',role='trainer',pre="if(!pm.environment.get('verification_document_ref'))throw Error('Нужен verification_document_ref загруженного тестового документа');",body="{qualification:'Test certificate',document_refs:[pm.environment.get('verification_document_ref')]}",test=store('verification_id')+"check('Заявка pending',()=>eq(data.status,'pending'));")
get(7,'Прочитать свою заявку','/api/v2/trainer/verification',role='trainer',test="check('Это созданная заявка',()=>eq(data.id,R('verification_id')));")
get(7,'Очередь проверки тренеров','/api/v2/admin/verifications',role='admin')
patch(7,'Подтвердить тренера','/api/v2/admin/verifications/{verification_id}',role='admin',body="{status:'approved',reason:'Postman verification'}",test="check('Статус approved',()=>eq(data.status,'approved'));")
get(7,'Список тренеров администратора','/api/v2/admin/trainers',role='admin')
get(7,'Публичный проверенный профиль','/api/v2/trainers/{trainer_id}',test="check('Проверенный тренер',()=>{eq(data.id,Number(R('trainer_id')));eq(data.verified,true);});")

def publish_program(stage,prefix,price):
    post(stage,f'Создать {prefix} программу','/api/v2/trainer/programs',role='trainer',body=f'draft({price})',test=store(prefix+'_program_id')+snapshot(prefix+'_author'))
    get(stage,f'Прочитать {prefix} черновик','/api/v2/trainer/programs/{program_id}',role='trainer',ids={'program_id':prefix+'_program_id'},test=snapshot(prefix+'_author'))
    post(stage,f'Отправить {prefix} программу на модерацию','/api/v2/trainer/programs/{program_id}/publish',role='trainer',ids={'program_id':prefix+'_program_id'},body=f"{{version:J('{prefix}_author').version}}",test="check('Заявка принята',()=>present(data.status));")
    patch(stage,f'Одобрить {prefix} программу','/api/v2/admin/programs/{program_id}/moderation',role='admin',ids={'program_id':prefix+'_program_id'},body="{status:'approved',reason:'Postman publication'}")
    get(stage,f'Зафиксировать опубликованную {prefix} версию','/api/v2/catalog/programs/{program_id}',ids={'program_id':prefix+'_program_id'},test=store(prefix+'_version_id','published_version_id')+f"check('Цена опубликованной версии',()=>eq(data.price_minor,{price}));"+snapshot(prefix+'_card'))

# 08: a free program supports the full deterministic marketplace flow; a paid one tests access and pending order.
publish_program(8,'free',0)
publish_program(8,'paid',100000)
get(8,'Список программ автора','/api/v2/trainer/programs',role='trainer',query={'limit':100},test="check('Обе программы автора',()=>{includesItem(data.items,R('free_program_id'));includesItem(data.items,R('paid_program_id'));});")
get(8,'Обычный пользователь не читает черновик автора','/api/v2/trainer/programs/{program_id}',ids={'program_id':'free_program_id'},codes=[403,404])
get(8,'Очередь модерации программ','/api/v2/admin/programs',role='admin')
post(8,'Создать черновик новой версии','/api/v2/trainer/programs/{program_id}/versions',role='trainer',ids={'program_id':'free_program_id'},test=snapshot('free_new_author'))
patch(8,'Изменить новый черновик','/api/v2/trainer/programs/{program_id}',role='trainer',ids={'program_id':'free_program_id'},body="{version:J('free_new_author').version,title:'Postman next version'}",test=snapshot('free_new_author'))
get(8,'Черновик не меняет опубликованный оригинал','/api/v2/catalog/programs/{program_id}',ids={'program_id':'free_program_id'},test="check('Публичная версия и название прежние',()=>{eq(data.published_version_id,R('free_version_id'));eq(data.title,J('free_card').title);});")

# 09: filtering, favorites and entitlement.
get(9,'Каталог по автору, цели, уровню и длительности','/api/v2/catalog/programs',query={'trainer_id':'{{state_trainer_id}}','goal':'muscle_gain','level':'beginner','duration_weeks':4,'limit':100},test="check('Фильтры соблюдены',()=>{includesItem(data.items,R('free_program_id'));data.items.forEach(x=>{eq(x.goal,'muscle_gain');eq(x.level,'beginner');eq(x.duration_weeks,4);eq(x.trainer.id,Number(R('trainer_id')));});});")
put(9,'Добавить в избранное','/api/v2/me/favorites/{program_id}',ids={'program_id':'free_program_id'})
put(9,'Повтор добавления в избранное','/api/v2/me/favorites/{program_id}',ids={'program_id':'free_program_id'})
get(9,'В избранном одна запись','/api/v2/me/favorites',query={'limit':100},test="check('Избранное не дублируется',()=>eq(data.items.filter(x=>x.id===R('free_program_id')).length,1));")
delete(9,'Удалить из избранного','/api/v2/me/favorites/{program_id}',ids={'program_id':'free_program_id'})
get(9,'Удаление из избранного сохранилось','/api/v2/me/favorites',query={'limit':100},test="check('Программа удалена',()=>eq(data.items.some(x=>x.id===R('free_program_id')),false));")
post(9,'Получить бесплатный доступ','/api/v2/programs/{program_id}/access',ids={'program_id':'free_program_id'},test="check('Доступ active к нужной программе',()=>{eq(data.program_id,R('free_program_id'));eq(data.status,'active');});")
get(9,'Бесплатная программа доступна полностью','/api/v2/catalog/programs/{program_id}',ids={'program_id':'free_program_id'},test="check('Состав программы доступен',()=>pm.expect(data.days.length).to.be.above(0));"+snapshot('free_access_card'))
post(9,'Нельзя бесплатно получить платную программу','/api/v2/programs/{program_id}/access',ids={'program_id':'paid_program_id'},codes=[403,422])

# 10: no fabricated payment confirmation. Pending orders must not unlock a program.
get(10,'Прочитать комиссию','/api/v2/admin/settings/commission',role='admin',test=snapshot('commission'))
patch(10,'Сохранить комиссию 20 процентов','/api/v2/admin/settings/commission',role='admin',body="{commission_bps:2000,effective_at:new Date().toISOString(),version:J('commission').version}",test="check('Ставка 2000 bps',()=>eq(data.commission_bps,2000));")
post(10,'Создать заказ с фиксацией цены','/api/v2/orders',pre="save('order_key',guid());",idem='order_key',body="{program_id:R('paid_program_id'),program_version_id:R('paid_version_id')}",test=store('order_id')+snapshot('order')+"check('Цена и валюта заказа',()=>{eq(data.amount_minor,100000);eq(data.currency,'RUB');pm.expect(data.status).to.be.oneOf(['created','pending']);});")
post(10,'Повтор заказа с тем же ключом','/api/v2/orders',idem='order_key',body="{program_id:R('paid_program_id'),program_version_id:R('paid_version_id')}",test="check('Заказ не дублируется',()=>eq(data.id,R('order_id')));")
get(10,'Статус созданного заказа','/api/v2/orders/{order_id}',test="check('Заказ пользователя',()=>eq(data.id,R('order_id')));")
get(10,'Чужой заказ недоступен','/api/v2/orders/{order_id}',role='stranger',codes=[403,404])
get(10,'История покупок пользователя','/api/v2/me/orders',query={'limit':100},test="check('Заказ в истории',()=>includesItem(data.items,R('order_id')));")
get(10,'Заказы администратора','/api/v2/admin/orders',role='admin')
get(10,'Реестр выплат','/api/v2/admin/payouts',role='admin')
get(10,'Выплаты тренеру','/api/v2/trainer/payouts',role='trainer')
post(10,'Pending заказ не даёт права копировать программу','/api/v2/programs/{program_id}/fork',ids={'program_id':'paid_program_id'},pre="save('pending_fork_key',guid());",idem='pending_fork_key',body="{source_version_id:R('paid_version_id'),title:'unpaid copy'}",codes=[403,404])

# 11: clone independence, source lineage, public sharing and revocation.
post(11,'Создать личную копию бесплатной программы','/api/v2/programs/{program_id}/fork',ids={'program_id':'free_program_id'},pre="save('fork_key',guid());",idem='fork_key',body="{source_version_id:R('free_version_id'),title:'Postman fork'}",test=store('fork_id','plan_id')+snapshot('fork_result')+"check('Происхождение сохранено',()=>eq(data.source_program_version_id,R('free_version_id')));")
post(11,'Повтор Fork возвращает ту же копию','/api/v2/programs/{program_id}/fork',ids={'program_id':'free_program_id'},idem='fork_key',body="{source_version_id:R('free_version_id'),title:'Postman fork'}",test="check('Не создана вторая копия',()=>eq(data,J('fork_result')));")
get(11,'Прочитать копию','/api/v2/me/plans/{plan_id}',ids={'plan_id':'fork_id'},test="check('Новые ID дней',()=>{present(data.days[0].id);eq(data.source_program_version_id,R('free_version_id'));pm.expect(data.days[0].id).not.to.eql(J('free_access_card').days[0].id);});"+snapshot('fork'))
patch(11,'Изменить упражнения в копии','/api/v2/me/plans/{plan_id}',ids={'plan_id':'fork_id'},pre="const b=J('fork');b.days[0].exercises[0].sets=4;save('fork_edit',b);",body="{version:J('fork').version,title:'Postman remixed',days:J('fork_edit').days,schedule:J('fork').schedule}",test=snapshot('fork'))
get(11,'Правки копии сохранены','/api/v2/me/plans/{plan_id}',ids={'plan_id':'fork_id'},test="check('Копия изменена',()=>{eq(data.title,'Postman remixed');eq(data.days[0].exercises[0].sets,4);});"+snapshot('fork'))
get(11,'Оригинал не изменился','/api/v2/catalog/programs/{program_id}',ids={'program_id':'free_program_id'},test="check('Исходные название, дни и версия',()=>{eq(data.title,J('free_access_card').title);eq(data.days,J('free_access_card').days);eq(data.published_version_id,R('free_version_id'));});")
post(11,'Создать публичную ссылку на собственный план','/api/v2/shares',pre="save('share_key',guid());",idem='share_key',body="{target_type:'plan',target_id:R('plan_id'),expires_at:null}",test=store('share_id')+"check('Извлечь токен ссылки',()=>{const url=data.url;present(url);const token=url.split('?')[0].replace(/\/$/,'').split('/').pop();present(token);save('share_token',decodeURIComponent(token));});")
get(11,'Открыть ссылку без авторизации','/api/v2/shares/{token}',role=None,ids={'token':'share_token'},test="check('Снимок плана доступен',()=>{eq(data.title,J('plan').title);pm.expect(data.days.length).to.be.above(0);});check('Публичный ответ без токена и email',()=>{pm.expect(data).not.to.have.property('email');pm.expect(data).not.to.have.property('token');});")
post(11,'Получатель сохраняет ссылку','/api/v2/shares/{token}/save',role='stranger',ids={'token':'share_token'})
post(11,'Получатель создаёт свою копию','/api/v2/shares/{token}/fork',role='stranger',ids={'token':'share_token'},pre="save('share_fork_key',guid());",idem='share_fork_key',test=store('recipient_plan_id','plan_id'))
get(11,'Копия принадлежит получателю','/api/v2/me/plans/{plan_id}',role='stranger',ids={'plan_id':'recipient_plan_id'},test="check('Это отдельный план',()=>pm.expect(data.id).not.to.eql(R('plan_id')));")
delete(11,'Отозвать публичную ссылку','/api/v2/shares/{share_id}')
get(11,'Отозванная ссылка недоступна','/api/v2/shares/{token}',role=None,ids={'token':'share_token'},codes=[404])

# 12: manual workouts do not qualify; only three distinct sessions of the fork do.
get(12,'До использования программы отзыв недоступен','/api/v2/programs/{program_id}/review-eligibility',ids={'program_id':'free_program_id'},test="check('Ручная тренировка не считается использованием программы',()=>{eq(data.eligible,false);eq(data.completed_workouts,0);eq(data.required_workouts,3);});")
post(12,'Нельзя оставить отзыв до тренировок','/api/v2/programs/{program_id}/reviews',ids={'program_id':'free_program_id'},body="{overall:5,text:'too early'}",codes=[403,422])
for i in range(1,4):
    post(12,f'Выполнить тренировку программы {i}/3','/api/v2/workouts/complete',pre="save('review_workout_key',guid());save('review_workout',workout(J('fork')));",idem='review_workout_key',body="J('review_workout')",test="check('Объём занятия',()=>eq(data.metrics.volume_kg,1080));")
    get(12,f'Право на отзыв после {i} занятий','/api/v2/programs/{program_id}/review-eligibility',ids={'program_id':'free_program_id'},test=f"check('Порог отзыва соблюдён',()=>{{eq(data.completed_workouts,{i});eq(data.required_workouts,3);eq(data.eligible,{str(i==3).lower()});}});")
post(12,'Создать отзыв после трёх занятий','/api/v2/programs/{program_id}/reviews',ids={'program_id':'free_program_id'},body="{overall:5,structure:5,description_match:4,text:'Postman verified review'}",test=store('review_id')+"check('Оценка и программа',()=>{eq(data.overall,5);eq(data.program_id,R('free_program_id'));});")
get(12,'Отзывы программы','/api/v2/programs/{program_id}/reviews',ids={'program_id':'free_program_id'})
patch(12,'Чужой пользователь не меняет отзыв','/api/v2/reviews/{review_id}',role='stranger',body="{overall:1,text:'hacked'}",codes=[403,404])
patch(12,'Редактировать свой отзыв','/api/v2/reviews/{review_id}',body="{overall:4,text:'Postman edited review'}",test="check('Оценка обновлена',()=>eq(data.overall,4));")
get(12,'Отзывы для модерации','/api/v2/admin/reviews',role='admin')
patch(12,'Скрыть отзыв администратором','/api/v2/admin/reviews/{review_id}/moderation',role='admin',body="{status:'hidden',reason:'Postman moderation'}",test="check('Отзыв hidden',()=>eq(data.status,'hidden'));")
get(12,'Скрытый отзыв не виден постороннему','/api/v2/programs/{program_id}/reviews',role='stranger',ids={'program_id':'free_program_id'},query={'limit':100},test="check('Скрытый отзыв исключён',()=>eq(data.items.some(x=>x.id===R('review_id')),false));")
post(12,'Подать жалобу на программу','/api/v2/reports',body="{target_type:'program',target_id:R('free_program_id'),reason:'Postman report'}",test=store('report_id'))
get(12,'Жалобы для администратора','/api/v2/admin/reports',role='admin')
patch(12,'Закрыть жалобу','/api/v2/admin/reports/{report_id}',role='admin',body="{status:'resolved',reason:'Postman handled'}",test="check('Жалоба resolved',()=>eq(data.status,'resolved'));")
delete(12,'Удалить собственный отзыв','/api/v2/reviews/{review_id}')

# 13: deterministic subscription state; delivery/read testing requires an actual notification fixture.
put(13,'Подписаться на тренера','/api/v2/trainers/{trainer_id}/follow',test="check('Подписка активна',()=>eq(data.following,true));")
put(13,'Повтор подписки','/api/v2/trainers/{trainer_id}/follow',test="check('Подписка остаётся активной',()=>eq(data.following,true));")
get(13,'Мои тренеры без дублей','/api/v2/me/trainers',query={'limit':100},test="check('Автор в списке один раз',()=>eq(data.items.filter(x=>x.id===Number(R('trainer_id'))).length,1));")
get(13,'Уведомления пользователя','/api/v2/notifications')
delete(13,'Отписаться от тренера','/api/v2/trainers/{trainer_id}/follow')
get(13,'Отписка сохранена','/api/v2/me/trainers',query={'limit':100},test="check('Автор отсутствует',()=>eq(data.items.some(x=>x.id===Number(R('trainer_id'))),false));")

# 14–15: no checks that assume a particular ML score, list length or dashboard state.
post(14,'Отправить событие просмотра','/api/v2/events',pre="save('event_body',{events:[{event_id:guid(),type:'view',program_id:R('free_program_id'),occurred_at:new Date().toISOString(),context:{source:'postman'}}]});",body="J('event_body')")
post(14,'Повтор события с тем же event_id','/api/v2/events',body="J('event_body')")
get(14,'Персональные рекомендации','/api/v2/recommendations/programs',query={'limit':10},test="check('Список и версия модели присутствуют',()=>{pm.expect(data.items).to.be.an('array');present(data.model_version);pm.expect(data.fallback).to.be.a('boolean');});check('Каждая рекомендация объяснена',()=>data.items.forEach(x=>{present(x.program_id);pm.expect(x.reason_codes).to.be.an('array');pm.expect(x.reason_codes.length).to.be.above(0);present(x.explanation);}));")
get(15,'Персональный дашборд','/api/v2/dashboard',test="check('Состояние и блоки дашборда',()=>{present(data.state);pm.expect(data.blocks).to.be.an('array');present(data.generated_at);data.blocks.forEach(b=>{present(b.id);present(b.type);pm.expect(b.priority).to.be.a('number');});});")
get(15,'Аналитика автора','/api/v2/trainer/analytics',role='trainer')
get(15,'Аналитика администратора','/api/v2/admin/analytics',role='admin')
get(15,'Аналитика администратора запрещена обычному пользователю','/api/v2/admin/analytics',codes=[403])

# 16: explicit gaps, no guessed provider signature or fabricated seeded money.
post(16,'Подписанный webhook выбранного провайдера','/api/v2/payments/webhook',role=None,gate='enable_provider_tests',
     pre="const raw=pm.environment.get('provider_webhook_body');if(!raw)throw Error('Нужны provider_webhook_body и provider_signature');if(!pm.environment.get('provider_signature'))throw Error('Нужна подпись провайдера');pm.request.headers.upsert({key:pm.environment.get('provider_signature_header')||'X-Signature',value:pm.environment.get('provider_signature')});save('provider_body',JSON.parse(raw));",
     body="J('provider_body')",description='Формат подписи не согласован. Перед запуском адаптировать запрос к выбранному провайдеру; использовать только sandbox. Это шаблон, не готовая проверка расчётов и возвратов.')
post(16,'Создать выплату из тестового подтверждённого баланса','/api/v2/admin/payouts',role='admin',gate='enable_payout_tests',pre="if(!pm.environment.get('payout_trainer_id')||!pm.environment.get('payout_amount_minor'))throw Error('Нужны payout_trainer_id и payout_amount_minor, тестовый доступный баланс');save('payout_key',guid());",idem='payout_key',body="{trainer_id:Number(pm.environment.get('payout_trainer_id')),amount_minor:Number(pm.environment.get('payout_amount_minor')),currency:'RUB'}",test=store('payout_id'))
patch(16,'Сверить созданную тестовую выплату','/api/v2/admin/payouts/{payout_id}',role='admin',gate='enable_payout_tests',body="{action:'reconcile',reason:'Postman sandbox reconciliation'}")
patch(16,'Пометить доставленное уведомление прочитанным','/api/v2/notifications/{notification_id}',gate='enable_notification_tests',pre="const id=pm.environment.get('notification_id');if(!id)throw Error('Нужен notification_id для пользователя этого прогона');save('notification_id',id);",body="{read:true}",test="check('Есть время прочтения',()=>present(data.read_at));")

COLLECTION = {
 'info':{'name':'Спортик · Backend / Frontend · Приёмочные API-тесты','_postman_id':str(uuid.uuid5(uuid.NAMESPACE_URL,'sportik-semester-acceptance')),
         'schema':'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
         'description':'Запускайте Runner сверху вниз, 1 iteration. max_stage ограничивает этапы. Нужны тестовый backend, admin_token и ссылка на тестовый документ. Каждое выполнение создаёт новые аккаунты и данные. Интеграционные шаблоны папки 16 отключены; их пропуск не означает успешное тестирование оплаты/выплат/доставки. Полная инструкция: docs/06_TESTS.md.'},
 'event':[{'listen':'test','script':{'type':'text/javascript','exec':COMMON_TEST.splitlines()}}],
 'variable':[{'key':'__contracts','value':json.dumps(CONTRACTS,ensure_ascii=False),'type':'string'},{'key':'__definitions','value':json.dumps(DEFINITIONS,ensure_ascii=False),'type':'string'}],
 'item':list(FOLDERS.values())}
ENV = {'id':str(uuid.uuid5(uuid.NAMESPACE_URL,'sportik-test-env')), 'name':'Спортик · Локальная тестовая БД', '_postman_variable_scope':'environment',
       'values':[{'key':k,'value':v,'type':'secret' if k=='admin_token' or 'signature' in k else 'default','enabled':True} for k,v in {
           'base_url':'http://localhost:5050','admin_token':'','test_environment':'true','max_stage':'4',
           'verification_document_ref':'','enable_provider_tests':'false','provider_webhook_body':'','provider_signature':'','provider_signature_header':'',
           'enable_payout_tests':'false','payout_trainer_id':'','payout_amount_minor':'','enable_notification_tests':'false','notification_id':''}.items()]}

def main():
    for file, data in [('Sportik.postman_collection.json',COLLECTION),('Local.postman_environment.json',ENV),('cases.json',CASES)]:
        (HERE/file).write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    planned={o['operationId'] for item in SPEC['paths'].values() for o in item.values() if o.get('x-status')=='planned' and o.get('x-service')=='go'}
    covered={c['operation'] for c in CASES}
    assert planned<=covered,planned-covered
    automated={c['operation'] for c in CASES if not c['gate']}
    report={'contract_sha256':hashlib.sha256((HERE.parent/'api/openapi.json').read_bytes()).hexdigest(),
            'requests':len(CASES),'automated_requests':sum(not c['gate'] for c in CASES),'negative_requests':sum(min(c['expected'])>=400 for c in CASES),
            'planned_operations':len(planned),'automated_planned_operations':len(planned & automated),'fixture_only_operations':sorted(planned-automated),
            'backend_execution':'NOT_RUN: semester endpoints are planned; offline validation is not acceptance of backend'}
    (HERE/'coverage.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(json.dumps(report,ensure_ascii=False))

if __name__=='__main__': main()
