"""Local Swagger server and fixed-target proxies for current Go/ML endpoints."""
import argparse
import http.client
import json
import re
from functools import partial
from http.server import SimpleHTTPRequestHandler,ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlsplit

HERE=Path(__file__).resolve().parent
DOCS=HERE.parent
TARGETS={'backend':('127.0.0.1',5050,'go'),'ml-plan':('127.0.0.1',8000,'ml-plan'),'ml-chat':('127.0.0.1',8001,'ml-chat')}
FORWARD={'authorization','content-type','accept','x-request-id'}
spec=json.loads((HERE/'openapi.json').read_text(encoding='utf-8'))
ALLOWED=[]
for path,methods in spec['paths'].items():
    for method,op in methods.items():
        if op.get('x-status')=='existing':
            parts=re.split(r'(\{[^}]+\})',path)
            pattern='^'+''.join('[^/]+' if p.startswith('{') else re.escape(p) for p in parts)+'$'
            ALLOWED.append((op['x-service'],method.upper(),re.compile(pattern)))

class Handler(SimpleHTTPRequestHandler):
    def proxy(self):
        parsed=urlsplit(self.path);match=re.fullmatch(r'/proxy/([^/]+)(/.*)',parsed.path)
        if not match or match[1] not in TARGETS: self.send_error(404,'Unknown local service');return
        host,port,service=TARGETS[match[1]];path=match[2]
        if not any(s==service and m==self.command and p.fullmatch(path) for s,m,p in ALLOWED):
            self.send_error(404,'Endpoint is not implemented in the current project');return
        try: length=int(self.headers.get('Content-Length','0'))
        except ValueError: self.send_error(400,'Invalid Content-Length');return
        if length<0 or length>8*1024*1024: self.send_error(413,'Request too large');return
        body=self.rfile.read(length) if length else None
        headers={k:v for k,v in self.headers.items() if k.lower() in FORWARD}
        endpoint=path+('?' + parsed.query if parsed.query else '')
        connection=http.client.HTTPConnection(host,port,timeout=90);started=False
        try:
            connection.request(self.command,endpoint,body,headers);upstream=connection.getresponse()
            self.send_response(upstream.status)
            for k,v in upstream.getheaders():
                if k.lower() in ['content-type','cache-control','x-request-id','authorization','retry-after']: self.send_header(k,v)
            self.send_header('Connection','close');self.end_headers();started=True
            while True:
                chunk=upstream.read1(16384)
                if not chunk: break
                self.wfile.write(chunk);self.wfile.flush()
            self.close_connection=True
        except (OSError,http.client.HTTPException):
            if not started:
                data=json.dumps({'message':'Локальный сервис недоступен. Запустите проект через Docker Compose.','service':service},ensure_ascii=False).encode('utf-8')
                self.send_response(502);self.send_header('Content-Type','application/json; charset=utf-8');self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data)
            self.close_connection=True
        finally: connection.close()
    def do_GET(self):
        if self.path.startswith('/proxy/'): self.proxy()
        elif self.path=='/': self.send_response(302);self.send_header('Location','/api/');self.end_headers()
        else: super().do_GET()
    def do_POST(self): self.proxy()
    def do_PATCH(self): self.proxy()
    def do_PUT(self): self.proxy()
    def do_DELETE(self): self.proxy()

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--port',type=int,default=8088);args=parser.parse_args()
    server=ThreadingHTTPServer(('127.0.0.1',args.port),partial(Handler,directory=str(DOCS)))
    print(f'Swagger: http://localhost:{args.port}/api/',flush=True)
    try: server.serve_forever()
    except KeyboardInterrupt: pass
    finally: server.server_close()

if __name__=='__main__': main()
