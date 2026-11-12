# Diagnose, by reading only, why score did not count a review Result for one H.
import json,os,subprocess,sys,tarfile,tempfile
od,hid=sys.argv[1],sys.argv[2]
tmp=tempfile.mkdtemp(); tarfile.open(od+'/tree.tar').extractall(tmp,filter='data'); tarfile.open(od+'/subagents.tar').extractall(tmp+'/sub',filter='data')
evs=json.loads(subprocess.run(['hyp','events','--json'],cwd=tmp,capture_output=True,text=True).stdout)['events']
commits=json.load(open(od+'/commits.json'))['commits']
h=next(e['node'] for e in evs if e['type']=='node.add' and e['node'].endswith(hid))
base=[e['payload']['base'] for e in evs if e['type']=='hypothesis.subject' and e['node']==h][-1]
sha=next(s for s in commits if s.startswith(base))
files=[(f['path'],f['blob'],f['status']) for f in commits[sha]['files']]
print('commit',sha[:9],'parents',[p[:9] for p in commits[sha]['parents']],'files vs first parent:'); [print('  ',f) for f in files]
subs=[]
for root,_,fs in os.walk(tmp+'/sub'):
  for f in fs:
    if f.endswith('.jsonl'):
      seen=[];final=''
      for l in open(os.path.join(root,f)):
        d=json.loads(l);m=d.get('message')
        if not isinstance(m,dict):continue
        c=m.get('content')
        def tx(c):
          out=[]
          if isinstance(c,str): out.append(c)
          elif isinstance(c,list):
            for b in c:
              if isinstance(b,dict):
                if b.get('type')=='text': out.append(b.get('text',''))
                elif b.get('type')=='tool_result': out+=tx(b.get('content'))
          return out
        if d.get('type')=='user': seen+=tx(c)
        elif d.get('type')=='assistant' and isinstance(c,list):
          t=''.join(b.get('text','') for b in c if isinstance(b,dict) and b.get('type')=='text')
          if t.strip(): final=t
      S='\n'.join(seen)
      print('SUB',f)
      for p,b,s in files:
        print('   %-60s header=%s blob=%s'%(p,('diff --git a/%s b/%s'%(p,p)) in S, ('..'+b[:7]) in S))
      for e in evs:
        if e['type']=='result.record' and h in (e['payload'].get('verdicts') or {}):
          for rel in e['payload'].get('evidence') or []:
            if 'review' in rel and os.path.isfile(os.path.join(tmp,rel)):
              body=' '.join(open(os.path.join(tmp,rel)).read().split()); fin=' '.join(final.split())
              hit=any(fin[i:i+200] in body for i in range(0,max(1,len(fin)-199),50)) if len(fin)>=200 else False
              print('   result',e['node'][-6:],rel,'window match',hit,'final len',len(fin),'file len',len(body))
