"""One-time migration of the original prose contract into canonical OpenAPI."""
import copy
import hashlib
import json
import re
from pathlib import Path

HERE=Path(__file__).resolve().parent
DOCS=HERE.parent
spec=json.loads((HERE/'openapi.json').read_text(encoding='utf-8'))
old=(DOCS/'02_API.md').read_text(encoding='utf-8')
protected={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in DOCS.glob('*.md') if p.name!='02_API.md'}
S=spec['components']['schemas']
st={'type':'string'}; it={'type':'integer'}; nu={'type':'number'}; bo={'type':'boolean'}
uid={'type':'string','format':'uuid'}; dt={'type':'string','format':'date-time'}; day={'type':'string','format':'date'}
def obj(p,req=(),description=None):
    out={'type':'object','properties':p}
    if req: out['required']=list(req)
    if description: out['description']=description
    return out
def ar(s): return {'type':'array','items':s}
def ref(n): return {'$ref':'#/components/schemas/'+n}
def nullable(s): return {'allOf':[s],'nullable':True} if '$ref' in s else {**s,'nullable':True}
def collection(n,cursor=True): return obj({'items':ar(ref(n)),**({'next_cursor':nullable(st)} if cursor else {})},['items'])
def schema(name,props,required=(),description=None): S[name]=obj(props,required,description)
def example(n,v): S[n]['example']=v
def response(op,name,code='200'):
    op['responses']={code:{'description':'Успешный результат','content':{'application/json':{'schema':ref(name)}}}}
def request(op,name,required=True): op['requestBody']={'required':required,'content':{'application/json':{'schema':ref(name)}}}

spec['info'].update(title='Спортик — общий контракт Go, frontend и ML',version='2026.10.05-2',description='Единый OpenAPI-контракт команды. Действующие операции описаны по исходникам; операции семестра помечены planned и ещё не реализованы. Go-разработчик реализует обработчики по контракту, frontend использует DTO и статусы, ML — межсервисные запросы, технический писатель — описания сценариев. Источник истины: docs/api/openapi.json. После изменения запустить python docs/api/build_openapi.py. Комиссия семестра: 20% (2000 bps), порог отзыва: 3 завершённые тренировки; взято из пользовательской редакции 01_PROJECT.md. Существующие операции проверены статически, доступность сервисов не гарантируется.')
spec['servers']=[{'url':'/proxy/backend','description':'Локальный Go через сервер Swagger: порт 5050'},{'url':'http://localhost:5050','description':'Прямой локальный Go (нужен CORS)'}]
spec['externalDocs']={'description':'Официальная документация Swagger / OpenAPI','url':'https://swagger.io/docs/specification/v3_0/about/'}
spec['x-conventions']={'ids':'Новые UUID, существующие users/exercises integer','time':'RFC3339 UTC; calendar date YYYY-MM-DD, IANA time_zone','money':'Целые минимальные единицы; RUB — копейки','commission_bps':2000,'review_required_workouts':3,'idempotency':'Для v2 заказа, генерации, Fork, complete и выплаты обязателен Idempotency-Key. Тот же ключ+тело возвращает исходный результат, другое тело 409; хранение минимум 24 часа — проектный параметр. provider_event_id дедуплицируется независимо.'}

# Keep business rules in Swagger rather than depending on a Markdown document.
rules={}
for fid,section in re.findall(r'<a id="(f\d+)"></a>\s*(.*?)(?=<a id="f\d+"></a>|\Z)',old,re.S):
    paras=[p.strip() for p in section.split('\n\n') if p.strip() and not p.lstrip().startswith(('#','|'))]
    rules[fid]='\n\n'.join(paras).replace('[ML-контракт](05_ML.md#ml-contracts)','ML-контракт доступен в разделе ML этой спецификации.').replace('[04_DATABASE](04_DATABASE.md#progress-rules)','Формулы прогресса: 04_DATABASE.md, раздел «Расчёты прогресса».').replace('[frontend-формат](02_API.md)','frontend-формат в текущем Go API')
rules['f02']=rules['f02'].replace('Это пример, не установленная ставка.','Ставка проекта — 20% (2000 bps); числовой пример выше относится к иной ставке и не используется.').replace('100000 копеек при 1500 bps → 15000 комиссии и 85000 автору','100000 копеек при 2000 bps → 20000 комиссии и 80000 автору')
rules['f02']=rules['f02'].replace('**Ставка проекта — 20% (2000 bps); числовой пример выше относится к иной ставке и не используется.**','Ставка проекта — 20% (2000 bps).')
def group(path):
    p=path.split('/api/v2')[-1]
    if p.startswith('/admin/'): return 'f12','Администрирование'
    if p.startswith(('/orders','/me/orders','/payments','/trainer/payouts','/trainer/analytics')): return 'f02','Покупки и выплаты'
    if p.startswith(('/recommendations','/events')): return 'f03','Рекомендации'
    if p in ['/plans/generate','/me/profile','/equipment']: return 'f04','Профиль и AI-план'
    if p.startswith(('/workouts','/progress')): return 'f05','Тренировки и прогресс'
    if p.startswith(('/me/plans','/exercises')): return 'f06','Личные планы'
    if p.endswith('/fork') and p.startswith('/programs'): return 'f07','Fork / Remix'
    if p.startswith('/shares'): return 'f08','Публичные ссылки'
    if '/reviews' in p or '/review-eligibility' in p: return 'f09','Отзывы'
    if p.startswith(('/trainers','/trainer/profile','/trainer/verification','/me/trainers','/notifications')): return 'f10','Тренеры и подписки'
    if p=='/dashboard': return 'f11','Дашборд'
    if p=='/reports': return 'f12','Жалобы'
    return 'f01','Маркетплейс'

schema('PasswordInput',{'current_password':st,'new_password':{'type':'string','minLength':8,'maxLength':72}},['new_password'])
S['Note']['properties'].update(user_id=it,updated_at=nullable(dt),deleted_at=nullable(dt))
S['NoteInput'].pop('required',None)
S['NoteInput']['description']='Пустые title/body принимаются транспортным слоем; ограничения сервиса применяются отдельно.'
schema('NotesV1',{'items':nullable(ar(ref('Note')))},['items'])
schema('AchievementV1',{'id':it,'title':st,'description':st,'category':st,'unlocked_at':dt})
S['AchievementsV1']=ar(ref('AchievementV1'))
schema('OK',{'ok':bo},['ok'])
schema('CreatedOK',{'id':it,'ok':bo},['id','ok'])
macros={k:nu for k in ['protein_g','fat_g','carbs_g','calories']}
schema('Macros',macros)
schema('NutritionEntryInput',{'dish_id':nullable(it),'grams':nullable(nu),'meal_type':{'type':'string','default':'snack','description':'breakfast/lunch/dinner/snack; неизвестное значение превращается в snack'},'title':st,'protein_g':nu,'fat_g':nu,'carbs_g':nu,'calories':nullable(nu),'consumed_at':nullable(dt),'day':day},['protein_g','fat_g','carbs_g'],description='БЖУ обязательны даже при dish_id. POST масштабирует пищевую ценность по grams; при dish_id берёт значения блюда. day имеет приоритет над consumed_at; использует Europe/Moscow. grams по умолчанию 100. PATCH — не произвольный partial update: БЖУ также обязательны.')
schema('NutritionEntry',{**macros,'id':it,'dish_id':it,'grams':nu,'meal_type':st,'title':st,'consumed_at':dt,'created_at':dt})
schema('NutritionEntries',{'items':ar(ref('NutritionEntry')),'day':day},['items','day'])
schema('NutritionFavoriteInput',{'title':st,'protein_g':nu,'fat_g':nu,'carbs_g':nu,'unit_type':{'type':'string','default':'gram'}},['protein_g','fat_g','carbs_g'])
schema('NutritionFavorite',{'id':it,'title':st,'protein_g':nu,'fat_g':nu,'carbs_g':nu,'unit_type':st})
S['NutritionFavorites']=collection('NutritionFavorite',False)
schema('NutritionGoal',{**macros,'updated_at':dt})
schema('NutritionGoalInput',{'target':st,'target_delta_kg':nullable(nu)},description='Пересчёт по профилю. Параметры необязательны; неизвестный target обрабатывается текущими правилами сервиса.')
schema('NutritionStats',{'day':ref('Macros'),'week_avg':ref('Macros'),'month_avg':ref('Macros')})
schema('NutritionDishInput',{'title':st,**macros,'base_grams':{'type':'number','default':100}},['title'],description='PATCH требует title; отсутствующие БЖУ превращаются в 0, это не partial PATCH. calories по умолчанию рассчитываются как 4P+9F+4C. При POST существующее название обновляет общую запись через upsert.')
schema('NutritionDish',{'id':it,'title':st,**macros,'base_grams':nu})
S['NutritionDishes']=collection('NutritionDish',False)
schema('NutritionWaterInput',{'amount_ml':it,'day':day},description='Заменяет итог за день, не прибавляет. Отрицательное количество превращается в 0; невалидная дата заменяется текущей датой Moscow.')
schema('NutritionWeightInput',{'weight_kg':{'type':'number','minimum':0,'exclusiveMinimum':True},'day':day},['weight_kg'])
schema('NutritionDashboard',{'day':day,'today':ref('Macros'),'goal':obj({**macros,'remaining_calories':nu}),'water':obj({'amount_ml':it,'goal_ml':it,'goal_liters':nu}),'weight':obj({'last_weight_kg':nullable(nu),'last_weight_day':nullable(dt),'need_weight_reminder':bo}),'streak_days':it,'burned_calories':nu})
schema('NutritionAnalytics',{'days':it,'food':ar(obj({'day':day,**macros})),'weight':ar(obj({'day':day,'weight_kg':nu})),'water':ar(obj({'day':day,'amount_ml':it}))})
example('Credentials',{'email':'user@example.com','password':'example-password'})
example('ProfileUpdate',{'version':1,'age':26,'height_cm':175,'weight_kg':75,'goal':'Набрать мышцы','fitness_level':'новичок','training_days_map':['Пн','Ср','Пт']})
example('WorkoutV1',{'day_code':'A','finished_at':'2026-10-05T15:00:00Z','slots':[{'slot_index':0,'exercise_id':46,'sets':[{'weight_kg':60,'reps':10}]}]})
example('ChatInput',{'model':'default','messages':[{'role':'user','content':'Как заменить упражнение в моём плане?'}]})
example('NutritionEntryInput',{'title':'Куриная грудка','grams':150,'meal_type':'lunch','protein_g':22.5,'fat_g':1.9,'carbs_g':0,'day':'2026-10-05'})

# Concrete proposed DTOs: explicit proposals, not assertions about current server output.
schema('Equipment',{'id':st,'code':st,'name':st},['id','code','name'])
schema('Category',{'id':uid,'slug':st,'name':st,'position':it,'is_active':bo},['id','name'])
schema('CategoryInput',{'slug':st,'name':st,'position':it,'is_active':bo},['slug','name'])
schema('Exercise',{'id':it,'name':st,'muscle_group':st,'muscle_subgroup':nullable(st),'equipment_ids':ar(st),'safe_for_injuries':bo,'image_url':nullable(st),'video_url':nullable(st),'is_active':bo},['id','name','equipment_ids'])
schema('ExerciseInput',{k:v for k,v in S['Exercise']['properties'].items() if k!='id'},['name','equipment_ids'])
schema('Trainer',{'id':it,'display_name':st,'photo_url':nullable(st),'bio':st,'specializations':ar(st),'experience_years':it,'qualifications':st,'verified':bo,'rating':nullable(nu),'review_count':it,'program_users':it,'published_programs':ar(uid)},['id','display_name','verified'])
schema('TrainerProfileInput',{'display_name':st,'photo_ref':nullable(st),'bio':st,'specializations':ar(st),'experience_years':{'type':'integer','minimum':0},'qualifications':st},['display_name'])
schema('ProgramCard',{'id':uid,'title':st,'trainer':ref('Trainer'),'goal':st,'level':st,'duration_weeks':it,'sessions_per_week':it,'equipment':ar(ref('Equipment')),'price_minor':it,'currency':st,'rating':nullable(nu),'review_count':it,'published_version_id':uid},['id','title','trainer','price_minor','currency','published_version_id'])
schema('ProgramDetail',{**S['ProgramCard']['properties'],'description':st,'access':{'type':'string','enum':['preview','free','paid']},'allowed_actions':ar(st),'days':ar(ref('PlanDay')),'allow_public_remix':bo},description='days отсутствуют в полном виде без entitlement; публично только явно разрешённое превью.')
schema('AuthorProgram',{'id':uid,'version':it,'published_version_id':nullable(uid),'status':{'type':'string','enum':['draft','in_review','published','rejected','hidden']},'draft':ref('ProgramDraft')},['id','version','status'])
schema('Entitlement',{'id':uid,'program_id':uid,'status':{'type':'string','enum':['active','revoked']},'granted_at':dt,'order_id':nullable(uid)},['id','program_id','status'])
schema('PersonalPlan',{'id':uid,'version':it,'title':st,'source_type':{'type':'string','enum':['manual','ai','fork','legacy_ai']},'source_program_version_id':nullable(uid),'days':ar(ref('PlanDay')),'schedule':ref('Schedule'),'is_active':bo},['id','version','title','days','is_active'])
schema('IdResult',{'id':uid,'version':it},['id'])
schema('WorkflowResult',{'id':uid,'status':st,'version':it},['id','status'])
schema('FollowResult',{'trainer_id':it,'following':bo},['trainer_id','following'])
schema('Notification',{'id':uid,'type':st,'created_at':dt,'read_at':nullable(dt),'title':st,'message':st,'target_route':st},['id','type','created_at'])
schema('NotificationUpdate',{'read':bo},['read'])
schema('Verification',{'id':uid,'trainer_id':it,'status':{'type':'string','enum':['pending','approved','rejected']},'qualification':st,'document_refs':ar(st),'reason':nullable(st),'created_at':dt,'reviewed_at':nullable(dt)},['id','status'])
schema('Review',{'id':uid,'program_id':uid,'program_version_id':uid,'user_id':it,**S['ReviewInput']['properties'],'status':{'type':'string','enum':['published','in_review','hidden']},'created_at':dt},['id','program_id','overall','status'])
schema('ReviewPage',{'items':ar(ref('Review')),'next_cursor':nullable(st),'average':nullable(nu),'count':it},['items','count'])
schema('ReportInput',{'target_type':{'type':'string','enum':['program','trainer','review']},'target_id':st,'reason':st},['target_type','target_id','reason'])
schema('Report',{'id':uid,'reporter_id':it,**S['ReportInput']['properties'],'status':{'type':'string','enum':['open','in_progress','resolved','dismissed']},'created_at':dt,'resolution':nullable(st)},['id','status'])
schema('ReportDecision',{'status':{'type':'string','enum':['in_progress','resolved','dismissed']},'reason':st},['status','reason'])
schema('ReviewModeration',{'status':{'type':'string','enum':['published','hidden']},'reason':st},['status','reason'])
schema('Payout',{'id':uid,'trainer_id':it,'amount_minor':it,'currency':st,'status':{'type':'string','enum':['requested','processing','paid','failed','cancelled']},'created_at':dt,'provider_payout_id':nullable(st)},['id','trainer_id','amount_minor','currency','status'])
schema('PayoutDecision',{'action':{'type':'string','enum':['retry','cancel','reconcile']},'reason':st},['action','reason'])
schema('Commission',{'commission_bps':{'type':'integer','minimum':0,'maximum':10000,'default':2000},'effective_at':dt,'version':it},['commission_bps','version'])
schema('UserAdmin',{'id':it,'email':nullable(st),'roles':ar(st),'blocked':bo,'created_at':dt},['id','roles','blocked'])
schema('WorkoutMetrics',{'volume_kg':nu,'working_sets':it,'exercise_count':it,'duration_sec':it},['volume_kg'])
schema('Highlight',{'exercise_id':nullable(it),'metric':st,'previous':nullable(nu),'current':nu,'delta_percent':nullable(nu),'message':st},['metric','current','message'])
schema('WorkoutResult',{'session_id':uid,'metrics':ref('WorkoutMetrics'),'highlights':ar(ref('Highlight')),'new_achievements':ar(ref('AchievementV1'))},['session_id','metrics','highlights'])
schema('Workout',{'id':uid,**S['WorkoutV2']['properties'],'metrics':ref('WorkoutMetrics'),'status':{'type':'string','enum':['completed']}},['id','finished_at','slots'])
schema('Progress',{'period':{'type':'string','enum':['week','month','all']},'start':dt,'end':dt,'time_zone':st,'workout_count':it,'volume_kg':nu,'adherence':nullable(nu),'streak_weeks':it,'weight_change_kg':nullable(nu),'highlights':ar(ref('Highlight')),'computed_at':dt},['period','workout_count','volume_kg','time_zone'])
schema('ExerciseProgress',{'exercise_id':it,'points':ar(obj({'session_id':uid,'finished_at':dt,'working_weight_kg':nu,'reps':it,'volume_kg':nu})),'records':ar(ref('Highlight'))},['exercise_id','points'])
schema('ProgressSummary',{'id':uid,'period':st,'period_start':day,'status':{'type':'string','enum':['ready','pending']},'metrics':nullable(ref('Progress')),'messages':ar(st)},['id','period','status'])
schema('TrainerAnalytics',{'from':dt,'to':dt,'program_users':it,'sales_count':it,'gross_minor':it,'refund_minor':it,'commission_minor':it,'author_earned_minor':it,'available_minor':it,'paid_out_minor':it,'currency':st,'programs':ar(obj({'program_id':uid,'users':it,'sales':it,'gross_minor':it}))})
schema('AdminAnalytics',{'from':dt,'to':dt,'users':it,'active_users':it,'workouts':it,'published_programs':it,'sales':it,'gross_minor':it,'refund_minor':it,'net_minor':it,'commission_accrued_minor':it,'commission_paid_minor':it,'currency':st,'top_programs':ar(obj({'program_id':uid,'title':st,'users':it,'sales':it})),'top_trainers':ar(obj({'trainer_id':it,'display_name':st,'users':it,'sales':it}))})
schema('PublicShare',{'id':uid,'title':st,'target_type':{'type':'string','enum':['plan','plan_day']},'author':ref('Trainer'),'days':ar(ref('PlanDay')),'requires_access':bo,'can_fork':bo,'source_program_id':nullable(uid),'source_program_version_id':nullable(uid),'snapshot_at':dt},['id','title','requires_access','can_fork'],description='Личные измерения и история владельца исключены. Для платного контента без entitlement — только превью.')
schema('SavedShare',{'library_item_id':uid},['library_item_id'])
schema('EventAccepted',{'accepted':it,'duplicate_count':it},['accepted'])
for name in ['Equipment','Category','Exercise','ProgramCard','AuthorProgram','PersonalPlan','Notification','Trainer','Verification','Report','Payout','UserAdmin','Workout','ProgressSummary','Order']:
    S[name+'Page']=collection(name)
example('Commission',{'commission_bps':2000,'effective_at':'2026-10-05T00:00:00Z','version':1})
example('OrderInput',{'program_id':'11111111-1111-4111-8111-111111111111','program_version_id':'22222222-2222-4222-8222-222222222222'})
example('ReviewInput',{'overall':5,'structure':5,'description_match':4,'text':'Выполнил три тренировки: расписание и оборудование подошли.'})
example('ReviewEligibility',{'eligible':False,'completed_workouts':1,'required_workouts':3,'reason':'REVIEW_NOT_ELIGIBLE'})
S['CommissionInput']['properties']['commission_bps']['default']=2000

current_map={
 '/auth/guest':('Создать гостевую учётную запись',None,'Token'),
 '/auth/register':('Зарегистрировать текущего гостя','Credentials','AuthResponse'),
 '/auth/login':('Войти по email и паролю','Credentials','AuthResponse'),
 '/auth/change-password':('Изменить пароль','PasswordInput',None),
 '/profile':('Профиль пользователя','ProfileUpdate','Profile'),
 '/plans/generate':('Сгенерировать и сохранить текущий план',None,'PlanV1'),
 '/workouts/complete':('Сохранить результаты тренировки','WorkoutV1','WorkoutResultV1'),
 '/notes':('Личные заметки','NoteInput','NotesV1'),
 '/notes/{id}':('Изменить или удалить свою заметку','NoteInput','Note'),
 '/achievements/get_achievements':('Получить открытые достижения',None,'AchievementsV1'),
 '/ai_chat/history':('Получить последние 50 сообщений чата',None,None),
 '/ai_chat/generate_answer':('Получить поток ответа AI-тренера','ChatInput',None),
 '/nutrition/entries':('Дневник питания','NutritionEntryInput','NutritionEntries'),
 '/nutrition/entries/{id}':('Изменить или удалить запись питания','NutritionEntryInput','OK'),
 '/nutrition/favorites':('Избранные продукты','NutritionFavoriteInput','NutritionFavorites'),
 '/nutrition/favorites/{id}':('Удалить избранный продукт',None,'OK'),
 '/nutrition/goals':('Получить цели БЖУ и калорий',None,'NutritionGoal'),
 '/nutrition/goals/recalculate':('Пересчитать цели по профилю','NutritionGoalInput','NutritionGoal'),
 '/nutrition/stats':('Сводка сегодня и средние за 7/30 дней',None,'NutritionStats'),
 '/nutrition/dashboard':('Получить дневную сводку питания',None,'NutritionDashboard'),
 '/nutrition/analytics':('Получить ряды питания, воды и веса',None,'NutritionAnalytics'),
 '/nutrition/dishes/search':('Найти блюда',None,'NutritionDishes'),
 '/nutrition/dishes/mine':('Получить собственные блюда',None,'NutritionDishes'),
 '/nutrition/dishes':('Создать или обновить блюдо по названию','NutritionDishInput','NutritionDish'),
 '/nutrition/dishes/{id}':('Изменить или удалить собственное блюдо','NutritionDishInput','NutritionDish'),
 '/nutrition/water':('Установить дневное количество воды','NutritionWaterInput','OK'),
 '/nutrition/weight':('Сохранить измерение веса за день','NutritionWeightInput','OK'),
}
planned_map={
 '/catalog/programs':('Каталог и фильтрация программ',None,'ProgramCardPage'),
 '/catalog/programs/{program_id}':('Карточка и доступный состав программы',None,'ProgramDetail'),
 '/catalog/categories':('Категории каталога',None,'CategoryPage'),
 '/trainer/programs':('Авторские программы','ProgramDraft','AuthorProgramPage'),
 '/trainer/programs/{program_id}':('Черновик авторской программы','ProgramDraftUpdate','AuthorProgram'),
 '/trainer/programs/{program_id}/publish':('Отправить версию на модерацию','VersionInput','WorkflowResult'),
 '/trainer/programs/{program_id}/versions':('Создать черновик новой версии',None,'AuthorProgram'),
 '/me/favorites':('Избранные программы',None,'ProgramCardPage'),
 '/me/favorites/{program_id}':('Сохранить или убрать избранное',None,'OK'),
 '/programs/{program_id}/access':('Получить бесплатный доступ',None,'Entitlement'),
 '/orders':('Создать заказ на программу','OrderInput','Order'),
 '/orders/{order_id}':('Проверить статус заказа',None,'Order'),
 '/me/orders':('История покупок',None,'OrderPage'),
 '/payments/webhook':('Подтверждение платежа от провайдера',None,'OK'),
 '/trainer/analytics':('Использование, продажи и доход автора',None,'TrainerAnalytics'),
 '/trainer/payouts':('Выплаты автору',None,'PayoutPage'),
 '/recommendations/programs':('Персональная подборка с объяснениями',None,'RecommendationResult'),
 '/events':('Принять клиентские события','EventsInput','EventAccepted'),
 '/plans/generate':('Создать улучшенный AI-план','GenerateInput','GenerateResult'),
 '/me/profile':('Расширенный профиль для ML','ProfileV2Update','ProfileV2Update'),
 '/equipment':('Справочник оборудования',None,'EquipmentPage'),
 '/workouts/complete':('Завершить занятие без дублирования','WorkoutV2','WorkoutResult'),
 '/workouts':('История тренировок',None,'WorkoutPage'),
 '/workouts/{session_id}':('Фактические подходы занятия',None,'Workout'),
 '/progress':('Показатели за неделю, месяц или всё время',None,'Progress'),
 '/progress/exercises/{exercise_id}':('Динамика упражнения и рекорды',None,'ExerciseProgress'),
 '/progress/summaries':('Недельные и месячные сводки',None,'ProgressSummaryPage'),
 '/progress/weight':('Измерение веса пользователя','WeightInput','OK'),
 '/me/plans':('Личные планы','PersonalPlanInput','PersonalPlanPage'),
 '/me/plans/{plan_id}':('Просмотр, изменение или архивирование плана','PersonalPlanUpdate','PersonalPlan'),
 '/me/plans/{plan_id}/activate':('Выбрать активный план','VersionInput','PersonalPlan'),
 '/exercises':('Поиск упражнений',None,'ExercisePage'),
 '/programs/{program_id}/fork':('Создать независимую личную копию','ForkInput','ForkResult'),
 '/shares':('Создать публичную ссылку','ShareInput','ShareResult'),
 '/shares/{token}':('Открыть разрешённый снимок',None,'PublicShare'),
 '/shares/{share_id}':('Отозвать свою ссылку',None,None),
 '/shares/{token}/save':('Сохранить ссылку в библиотеке',None,'SavedShare'),
 '/shares/{token}/fork':('Создать копию по разрешённой ссылке',None,'ForkResult'),
 '/programs/{program_id}/reviews':('Отзывы программы','ReviewInput','ReviewPage'),
 '/programs/{program_id}/review-eligibility':('Проверить право на отзыв после 3 занятий',None,'ReviewEligibility'),
 '/reviews/{review_id}':('Изменить или удалить собственный отзыв','ReviewInput','Review'),
 '/trainers/{trainer_id}':('Публичный профиль тренера',None,'Trainer'),
 '/trainer/profile':('Профессиональный профиль автора','TrainerProfileInput','Trainer'),
 '/trainer/verification':('Заявка на проверку тренера','VerificationInput','Verification'),
 '/trainers/{trainer_id}/follow':('Подписаться или отписаться',None,'FollowResult'),
 '/me/trainers':('Мои тренеры',None,'TrainerPage'),
 '/notifications':('Уведомления',None,'NotificationPage'),
 '/notifications/{notification_id}':('Отметить уведомление прочитанным','NotificationUpdate','Notification'),
 '/dashboard':('Персональная главная',None,'Dashboard'),
 '/reports':('Пожаловаться на контент','ReportInput','Report'),
 '/admin/users':('Управление пользователями',None,'UserAdminPage'),
 '/admin/users/{user_id}':('Блокировка и назначение ролей','BlockInput','UserAdmin'),
 '/admin/trainers':('Управление тренерами',None,'TrainerPage'),
 '/admin/verifications':('Очередь проверки тренеров',None,'VerificationPage'),
 '/admin/verifications/{verification_id}':('Подтвердить или отклонить статус','DecisionInput','Verification'),
 '/admin/programs':('Очередь программ',None,'AuthorProgramPage'),
 '/admin/programs/{program_id}/moderation':('Модерация версии программы','ModerationInput','WorkflowResult'),
 '/admin/exercises':('Справочник упражнений','ExerciseInput','ExercisePage'),
 '/admin/exercises/{exercise_id}':('Изменить или архивировать упражнение','ExerciseInput','Exercise'),
 '/admin/categories':('Категории маркетплейса','CategoryInput','CategoryPage'),
 '/admin/categories/{category_id}':('Изменить категорию','CategoryInput','Category'),
 '/admin/reviews':('Отзывы для модерации',None,'ReviewPage'),
 '/admin/reviews/{review_id}/moderation':('Скрыть или восстановить отзыв','ReviewModeration','Review'),
 '/admin/reports':('Жалобы пользователей',None,'ReportPage'),
 '/admin/reports/{report_id}':('Обработать жалобу','ReportDecision','Report'),
 '/admin/orders':('Покупки и платежные состояния',None,'OrderPage'),
 '/admin/payouts':('Реестр и создание выплат','PayoutInput','PayoutPage'),
 '/admin/payouts/{payout_id}':('Повтор, отмена или сверка выплаты','PayoutDecision','Payout'),
 '/admin/settings/commission':('Комиссия платформы: 20%','CommissionInput','Commission'),
 '/admin/analytics':('Статистика платформы',None,'AdminAnalytics'),
}
source_map={'auth':'backend/internal/features/users/transport/http','profile':'backend/internal/features/users/transport/http','plans':'backend/internal/features/users/service/generate_plan.go','workouts':'backend/internal/features/users/transport/http/complete_workout.go','notes':'backend/internal/features/users/transport/http/notes.go','achievements':'backend/internal/features/users/transport/http/get_achievements.go','ai_chat':'backend/internal/features/mlclient/aichat/service.go','nutrition':'backend/internal/features/nutrition/nutrition.go'}
explicit_v1={
 '/auth/guest':'Без JWT. Создаёт пользователя и возвращает гостевой JWT. Лимит 10 запросов в минуту. Это запись в БД, не просмотр.',
 '/auth/register':'Нужен JWT анонимного пользователя. Преобразует текущего гостя, сохраняя его данные; JWT зарегистрированного пользователя здесь даёт 409. Лимит 20/мин. frontend name сейчас не используется Go DTO.',
 '/auth/login':'Без JWT. Возвращает JWT и сокращённые данные пользователя. Лимит 30/мин.',
 '/profile':'GET возвращает текущую анкету. POST принимает изменяемые поля с обязательной числовой version. Пустое изменение даёт 400, несовпадение версии — 409.',
 '/plans/generate':'Тело не требуется: профиль берётся из PostgreSQL. Go вызывает ML /plan/user с deadline 52 секунды, подбирает упражнения и сохраняет текущий план. exercises[][] — альтернативы упражнения в слоте, не подходы. Это изменяет текущую программу.',
 '/workouts/complete':'Сохраняет фактические подходы, обновляет веса и достижения. saved_id генерируется из текущего времени; идемпотентность не гарантирована. plan_updates не включён в HTTP-ответ.',
 '/ai_chat/generate_answer':'Go добавляет профиль и план и обращается к Python. Ответ — text/plain поток из текущего ML, не обязательно SSE. Не парсить как JSON. Лимит 30/мин на пользователя. После начала потока HTTP status нельзя изменить; завершение соединения не доказывает полный ответ.',
 '/nutrition/goals':'При отсутствии целей пытается рассчитать по профилю; если это невозможно, может вернуть 204 без тела.',
 '/nutrition/analytics':'Дневные серии питания, веса и воды. days default 30; текущий Go группирует food по DATE(consumed_at) в БД, а dashboard по Europe/Moscow. Не считать эти временные границы одинаковыми автоматически.',
}
for full,methods in spec['paths'].items():
    version='v1' if full.startswith('/api/v1') else 'v2'
    path=full[len('/api/'+version):]
    for method,op in methods.items():
        mapping=current_map if version=='v1' else planned_map
        title,req,res=mapping[path]
        op['summary']=title
        op['x-service']='go'
        if version=='v1':
            feature=path.lstrip('/').split('/')[0]
            label={'auth':'Авторизация','profile':'Профиль','plans':'AI-план','workouts':'Тренировки','notes':'Заметки','achievements':'Достижения','ai_chat':'AI-чат','nutrition':'Питание'}[feature]
            op['tags']=['Go v1 · '+label]
            op['description']='**Действующий маршрут.** '+explicit_v1.get(path,title+'. Форматы описаны ниже по текущему Go-коду.')
            op['x-source']=source_map[feature]
        else:
            fid,label=group(full)
            op['tags']=['Семестр · '+label]
            op['description']='**Проектируется, пока не реализовано.** '+title+'.\n\n'+rules.get(fid,'')
            op['x-feature']=fid.upper()
            op['description']+='\n\nОбщие правила: JWT и право на конкретный объект проверяет Go. Деньги в минимальных единицах. Изменение плана/черновика с version; конфликт 409. Коллекции items/next_cursor, limit 1–100, default 20. Технические подробности ошибок не выдаются пользователю.'
            if path.startswith('/admin/'): op['description']+=' Только admin; изменение требует причины и аудита.'
            elif path.startswith('/trainer/'): op['description']+=' Только тренер; собственные объекты.'
        if method in ['post','patch']:
            if req: request(op,req)
            else: op.pop('requestBody',None)
        else: op.pop('requestBody',None)
        code=next(c for c in op['responses'] if c!='default')
        if method=='post':
            if path=='/notes' and version=='v1': res='Note'
            elif path=='/nutrition/entries' and version=='v1': res='NutritionEntry'
            elif path=='/nutrition/favorites' and version=='v1': res='CreatedOK'
            elif path=='/trainer/programs': res='AuthorProgram'
            elif path=='/me/plans': res='PersonalPlan'
            elif path=='/programs/{program_id}/reviews': res='Review'
            elif path=='/admin/exercises': res='Exercise'
            elif path=='/admin/categories': res='Category'
            elif path=='/admin/payouts': res='Payout'
        if code=='204': op['responses']={'204':{'description':'Успех, без тела'}}
        elif res: response(op,res,code)
        if path=='/ai_chat/history': op['responses']={'200':{'description':'Последние 50 сообщений','content':{'application/json':{'schema':ar(ref('ChatMessage'))}}}}
        if path=='/ai_chat/generate_answer': op['responses']={'200':{'description':'Текстовый поток; Content-Type берётся из ML upstream','content':{'text/plain':{'schema':st},'text/event-stream':{'schema':st}}}}
        if path=='/nutrition/goals': op['responses']['204']={'description':'Цели отсутствуют, пересчёт невозможен; тела нет'}
        # Exact provider payload is intentionally pending provider selection.
        if path=='/payments/webhook':
            request(op,'ProviderWebhook')
            op['description']+='\n\nФормат и подпись зависят от ещё не выбранного провайдера. ProviderWebhook ниже — согласуемая нормализованная схема, не формат конкретного платёжного API. Нельзя интегрировать провайдера без адаптера и проверки подписи.'
            op['x-provider-contract-pending']=True
        err_schema=ref('ErrorV1' if version=='v1' else 'ErrorV2')
        codes=['400','401','404','409','429','500'] if version=='v1' else ['400','401','403','404','409','422','429','503','504']
        labels={'400':'Некорректный запрос','401':'Нужен действительный JWT','403':'Недостаточно прав','404':'Объект не найден или недоступен','409':'Конфликт состояния или версии','422':'Ограничение предметной области','429':'Превышен лимит','500':'Ошибка текущего сервиса/зависимости','503':'Сервис временно недоступен','504':'Истёк deadline'}
        for code in codes:
            content={'application/json':{'schema':err_schema}}
            if version=='v1': content['text/plain']={'schema':st}
            op['responses'][code]={'description':labels[code]+(' — возможный общий ответ; не все коды применимы к каждой операции' if version=='v1' else ''),'content':content}
        if version=='v1': op['description']+='\n\nОшибки users/чат обычно JSON {message,error}; nutrition/middleware могут вернуть text/plain. Коды ошибок перечислены как общие возможные ответы текущего слоя.'
        if version=='v2':
            idem=any(p['name']=='Idempotency-Key' for p in op.get('parameters',[]))
            if idem: op['description']+='\n\nIdempotency-Key обязателен: одинаковые ключ и тело возвращают исходный результат, другое тело с тем же ключом даёт 409.'
        if version=='v1' and path=='/nutrition/dishes/search':
            op['parameters'].append({'in':'query','name':'limit','schema':{'type':'integer','minimum':1,'maximum':200,'default':30}})
        if version=='v1' and path=='/nutrition/analytics':
            next(p for p in op['parameters'] if p['name']=='days')['schema']={'type':'integer','minimum':1,'maximum':4000,'default':30}
        if version=='v2' and method=='get' and res and res.endswith('Page'):
            names={p['name'] for p in op.get('parameters',[])}
            op.setdefault('parameters',[])
            for name,typ in [('limit',{'type':'integer','minimum':1,'maximum':100,'default':20}),('cursor',st)]:
                if name not in names: op['parameters'].append({'in':'query','name':name,'schema':typ})

schema('ProviderWebhook',{'provider_event_id':st,'provider_payment_id':st,'order_id':uid,'event_type':{'type':'string','enum':['payment_succeeded','payment_failed','refund_succeeded']},'amount_minor':it,'currency':st},['provider_event_id','provider_payment_id','order_id','event_type','amount_minor','currency'],description='Нормализованное предложение; реальный формат и подпись определяет выбранный провайдер. Статус доступа меняется только после проверенного события.')
schema('MLProfileV1',{'age':it,'height_cm':it,'weight_kg':nu,'gender':st,'activity_level':st,'injuries_notes':bo,'goal':st,'fitness_level':st,'training_days_map':ar(st)},description='Фактически FastAPI принимает Dict[str,Any]. Документированы английские ключи Go-профиля; ML также распознаёт русские алиасы и использует defaults. Схема не вводит новую валидацию.')
S['MLProfileV1']['additionalProperties']=True
schema('MLMuscle',{'группа':st,'подгруппа':nullable(st)},['группа','подгруппа'])
schema('MLPlanV1',{'тип_сплита':{'type':'string','enum':['Фулбади','ВерхНиз','Тяни-толкай','Пуш-пул-ноги','Сплит']},'еженедельный_план':ar(obj({'день':st,'тип_дня':st,'упражнения':ar(ref('MLMuscle'))},['день','тип_дня','упражнения']))},['тип_сплита','еженедельный_план'])
schema('MLChatV1',{'messages':ar(obj({'role':st,'content':st})),'user':{'type':'object','additionalProperties':True},'plan':{'type':'object','nullable':True,'additionalProperties':True}},['messages','user'],description='messages — list, user — dict, plan — optional dict в текущем Pydantic. Go посылает русские поля user; часть Python fallback читает английские ключи — известное несоответствие контекста.')
schema('MLValidationError',{'detail':ar(obj({'loc':ar({'oneOf':[st,it]}),'msg':st,'type':st}))},['detail'])
ml_profile={'age':26,'height_cm':175,'weight_kg':75,'gender':'м','activity_level':'Средняя активность','injuries_notes':False,'goal':'Набрать мышцы','fitness_level':'новичок','training_days_map':['Пн','Ср','Пт']}
example('MLProfileV1',ml_profile)
example('MLPlanV1',{'тип_сплита':'Фулбади','еженедельный_план':[{'день':d,'тип_дня':'Фулбоди '+d,'упражнения':[{'группа':'Грудь','подгруппа':'грудные середина'}]} for d in ['A','B','C']]})
example('MLChatV1',{'messages':[{'role':'user','content':'Как заменить упражнение?'}],'user':{'возраст':26,'цель':'Набрать мышцы','уровень_подготовки':'новичок'},'plan':{}})
def ml_operation(path,method,title,req,res,service,planned=False,text=False,desc=''):
    port=8000 if service=='ml-plan' else 8001
    op={'operationId':service.replace('-','_')+'_'+method+'_'+re.sub(r'\W+','_',path).strip('_'),'summary':title,'description':('**Проектируется, пока не реализовано.** ' if planned else '**Действующий ML endpoint.** ')+desc,'tags':['ML · '+('Семестр' if planned else 'Планы' if service=='ml-plan' else 'Чат')],'x-status':'planned' if planned else 'existing','x-service':service,'servers':[{'url':'/proxy/'+service,'description':'Локальный ML через Swagger server'},{'url':'http://localhost:'+str(port),'description':'Прямой ML, требует CORS'}],'security':[]}
    if req: request(op,req,required=path!='/plan/user')
    if text: op['responses']={'200':{'description':'Текстовые фрагменты без SSE framing','content':{'text/plain':{'schema':st}}}}
    else: response(op,res)
    op['responses']['422']={'description':'Ошибка схемы FastAPI/Pydantic' if not planned else 'Невалидные признаки или нет допустимого результата','content':{'application/json':{'schema':ref('MLValidationError' if not planned else 'MLError')}}}
    op['responses']['500']={'description':'Ошибка загрузки модели, inference или формирования ответа','content':{'text/plain':{'schema':st}}}
    if planned:
        op['security']=[{'ServiceAuth':[]}]
        op['parameters']=[{'name':'X-Request-ID','in':'header','required':True,'schema':st}]
        op['responses']['503']={'description':'Модель/зависимость недоступна','content':{'application/json':{'schema':ref('MLError')}}}
    spec['paths'].setdefault(path,{})[method]=op
ml_operation('/health','get','Проверить ML-сервис плана',None,'OK','ml-plan',desc='Проверка доступности HTTP; сама по себе не проверяет качество модели.')
ml_operation('/plan/user','post','Предсказать сплит и выбрать шаблон','MLProfileV1','MLPlanV1','ml-plan',desc='CatBoost на 8 признаках, затем случайный шаблон того же сплита и числа дней. Не возвращает финальные подходы/веса: их собирает Go. При отсутствии шаблона возможен 500 из-за None при заявленном Dict response. Внутри Compose Go использует http://ml:8000. В Swagger — локальный порт 8000.')
ml_operation('/ai/user/chat','post','Ответить по профилю и плану','MLChatV1',None,'ml-chat',text=True,desc='Последние 10 сообщений в prompt. Cerebras/g4f при наличии конфигурации, иначе шаблонный ответ. Поток text/plain, не SSE; клиент читает текстовые chunks. Сервис находится на локальном порту 8001.')

schema('MLHistory',{'completed_28d':it,'adherence_28d':nullable(nu),'exercise_metrics':ar(obj({'exercise_id':it,'weight_kg':nu,'reps':it,'volume_kg':nu,'measured_at':dt}))})
schema('MLProfileV2',{'age':it,'height_cm':nu,'weight_kg':nu,'level':st,'goal':st,'training_days':ar({'type':'integer','minimum':1,'maximum':7}),'equipment_ids':ar(st),'excluded_exercise_ids':ar(it)},['level','goal','training_days','equipment_ids'])
schema('MLGenerateInput',{'request_id':st,'feature_schema_version':st,'profile':ref('MLProfileV2'),'history':nullable(ref('MLHistory'))},['request_id','feature_schema_version','profile'])
schema('MLGenerateResult',{'request_id':st,'model_version':st,'feature_schema_version':st,'split':st,'days':ar(obj({'weekday':it,'title':st,'exercises':ar(ref('Prescription'))})),'warnings':ar(st),'fallback':bo},['request_id','model_version','split','days','fallback'])
schema('MLCandidate',{'program_id':uid,'version_id':uid,'level':st,'goal':st,'sessions_per_week':it,'equipment_ids':ar(st)},['program_id','version_id','level','goal','sessions_per_week','equipment_ids'])
schema('MLRankInput',{'request_id':st,'feature_schema_version':st,'context':obj({'level':st,'goal':st,'sessions_per_week':it,'equipment_ids':ar(st),'progress':ref('MLHistory'),'behavior':obj({'viewed_program_ids':ar(uid),'saved_program_ids':ar(uid),'purchased_program_ids':ar(uid),'ratings':ar(obj({'program_id':uid,'overall':it}))})}),'candidates':ar(ref('MLCandidate')),'limit':{'type':'integer','minimum':1,'maximum':100}},['request_id','feature_schema_version','context','candidates','limit'])
schema('Evidence',{'sessions_per_week':it,'level':st,'equipment_match':bo,'goal_match':bo})
schema('MLRankItem',{'program_id':uid,'version_id':uid,'score':nu,'reason_codes':ar(st),'evidence':ref('Evidence')},['program_id','version_id','score','reason_codes','evidence'])
schema('MLRankResult',{'request_id':st,'model_version':st,'items':ar(ref('MLRankItem')),'fallback':bo},['request_id','model_version','items','fallback'])
schema('MLExplainInput',{'request_id':st,'locale':st,'items':ar(obj({'program_id':uid,'version_id':uid,'reason_codes':ar(st),'evidence':ref('Evidence'),'allowed_context':obj({'program_description':st,'user_summary':st})},['program_id','version_id','reason_codes','evidence']))},['request_id','locale','items'])
schema('MLExplainResult',{'request_id':st,'items':ar(obj({'program_id':uid,'text':st,'source_version_ids':ar(uid),'explanation_mode':{'type':'string','enum':['rag','template']}}))},['request_id','items'])
schema('MLError',{'request_id':st,'error':obj({'code':st,'message':st,'retryable':bo},['code','message','retryable'])},['request_id','error'])
spec['components']['securitySchemes']['ServiceAuth']={'type':'http','scheme':'bearer','description':'Проектируемый служебный токен Go↔ML; не пользовательский JWT'}
ml_operation('/v2/plans/generate','post','Генерация плана с назначениями','MLGenerateInput','MLGenerateResult','ml-plan',True,desc='Правила оборудования/ограничений приоритетнее classifier score. Go повторно валидирует упражнения, частоту и состав. null weight требует подтверждения. Предлагаемый deadline 8 секунд, p95 полной цепочки Go ≤5 секунд. Старый план не теряется при ошибке.')
ml_operation('/v2/recommendations/rank','post','Ранжировать допустимые программы','MLRankInput','MLRankResult','ml-plan',True,desc='Content-based/collaborative гибрид, cold start по профилю. Id только из candidates. Go проверяет публикацию/доступ перед показом. Предлагаемый deadline 1 секунда, p95 ≤500 мс; fallback совместимые правила/cache.')
ml_operation('/v2/recommendations/explain','post','Объяснить рекомендацию по фактам','MLExplainInput','MLExplainResult','ml-plan',True,desc='RAG/LLM не определяет право доступа и не выдумывает facts. Только evidence и разрешённый контекст. Предлагаемый deadline 3 секунды; p95 ≤2 секунды. При ошибке шаблонное объяснение.')

# Remove the generic incomplete-object schema and all references to the old long Markdown.
encoded=json.dumps(spec,ensure_ascii=False)
assert '#/components/schemas/DocumentedObject' not in encoded
S.pop('DocumentedObject',None)
for p in DOCS.glob('*.md'):
    if p.name!='02_API.md': assert hashlib.sha256(p.read_bytes()).hexdigest()==protected[p.name]
(HERE/'openapi.json').write_text(json.dumps(spec,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
print('Canonical OpenAPI migrated: '+str(sum(len(v) for v in spec['paths'].values()))+' operations, '+str(len(S))+' schemas')
