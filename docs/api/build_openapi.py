"""Validate canonical OpenAPI and export three Swagger views; no Markdown dependency."""
import copy
import json
import re
from pathlib import Path

HERE=Path(__file__).resolve().parent
METHODS={'get','post','put','patch','delete','head','options','trace'}

def validate(spec):
    assert spec['openapi']=='3.0.3'
    assert spec['info']['title'] and spec['info']['version']
    ids=set()
    for path,item in spec['paths'].items():
        assert path.startswith('/'),path
        for method,op in item.items():
            if method not in METHODS: continue
            assert op['operationId'] not in ids,op['operationId']
            ids.add(op['operationId'])
            assert op['summary'] and op['description'],path
            assert op['x-status'] in ['existing','planned']
            assert op['responses']
            needed=set(re.findall(r'\{([^}]+)\}',path))
            actual={p['name'] for p in op.get('parameters',[]) if p['in']=='path'}
            assert needed==actual,(path,needed,actual)
            for p in op.get('parameters',[]):
                assert 'schema' in p
                if p['in']=='path': assert p['required']
            for code,res in op['responses'].items():
                assert code=='default' or re.fullmatch(r'[1-5][0-9X]{2}',code),code
                assert 'description' in res
                if code=='204': assert 'content' not in res
    def walk(value):
        if isinstance(value,dict):
            if '$ref' in value:
                pointer=value['$ref'];assert pointer.startswith('#/'),pointer
                cur=spec
                for part in pointer[2:].split('/'): cur=cur[part.replace('~1','/').replace('~0','~')]
            if isinstance(value.get('required'),list): assert set(value['required'])<=set(value.get('properties',{}))
            for v in value.values(): walk(v)
        elif isinstance(value,list):
            for v in value: walk(v)
    walk(spec)
    return len(ids)

def subset(spec,predicate,title):
    out=copy.deepcopy(spec);out['info']['title']=title;out['paths']={};tags=set()
    for path,item in spec['paths'].items():
        selected={m:copy.deepcopy(o) for m,o in item.items() if m in METHODS and predicate(o)}
        if selected:
            out['paths'][path]=selected
            for op in selected.values(): tags.update(op['tags'])
    out['tags']=[copy.deepcopy(tag) for tag in spec.get('tags',[]) if tag['name'] in tags]
    known={tag['name'] for tag in out['tags']}
    out['tags'].extend({'name':tag} for tag in sorted(tags-known))
    needed=set()
    def collect(value):
        if isinstance(value,dict):
            if '$ref' in value and value['$ref'].startswith('#/components/schemas/'):
                name=value['$ref'].split('/')[-1]
                if name not in needed: needed.add(name);collect(spec['components']['schemas'][name])
            for v in value.values(): collect(v)
        elif isinstance(value,list):
            for v in value: collect(v)
    collect(out['paths'])
    out['components']['schemas']={n:s for n,s in out['components']['schemas'].items() if n in needed}
    if 'ML' in title:
        out['security']=[]
        out['servers']=[{'url':'/proxy/ml-plan','description':'Локальный ML-план, порт 8000'},{'url':'/proxy/ml-chat','description':'Локальный ML-чат, порт 8001'}]
    used_auth={name for path in out['paths'].values() for op in path.values() for requirement in op.get('security',out.get('security',[])) for name in requirement}
    out['components']['securitySchemes']={name:scheme for name,scheme in out['components']['securitySchemes'].items() if name in used_auth}
    validate(out)
    return out

def yaml_lines(value, indent=0):
    """Serialize JSON-compatible OpenAPI as standalone YAML without dependencies."""
    pad=' '*indent
    scalar=lambda v:json.dumps(v,ensure_ascii=False)
    if isinstance(value,dict) and value:
        for key,item in value.items():
            if isinstance(item,(dict,list)) and item:
                yield pad+scalar(key)+':'
                yield from yaml_lines(item,indent+2)
            else: yield pad+scalar(key)+': '+scalar(item)
    elif isinstance(value,list) and value:
        for item in value:
            if isinstance(item,(dict,list)) and item:
                yield pad+'-'
                yield from yaml_lines(item,indent+2)
            else: yield pad+'- '+scalar(item)
    else: yield pad+scalar(value)

def main():
    spec=json.loads((HERE/'openapi.json').read_text(encoding='utf-8'));count=validate(spec)
    (HERE/'openapi.yml').write_text('\n'.join(yaml_lines(spec))+'\n',encoding='utf-8')
    exports=[('current.json',lambda o:o['x-status']=='existing' and o['x-service']=='go','Спортик — действующий Go API'),('ml.json',lambda o:o['x-status']=='existing' and o['x-service'].startswith('ml-'),'Спортик — действующие ML-сервисы'),('semester.json',lambda o:o['x-status']=='planned','Спортик — контракт нового семестра')]
    result={}
    for file,select,title in exports:
        view=subset(spec,select,title)
        (HERE/file).write_text(json.dumps(view,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
        result[file]=sum(len(v) for v in view['paths'].values())
    print(json.dumps({'validated_operations':count,'views':result},ensure_ascii=False))

if __name__=='__main__': main()
