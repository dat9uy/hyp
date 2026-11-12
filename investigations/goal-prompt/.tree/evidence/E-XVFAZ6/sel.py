import json,subprocess,sys,tarfile,tempfile
for od in sys.argv[1:]:
    st=dict(l.rstrip('\n').split('=',1) for l in open(od+'/state.txt') if '=' in l)
    tmp=tempfile.mkdtemp(); tarfile.open(od+'/tree.tar').extractall(tmp,filter='data')
    evs=json.loads(subprocess.run(['hyp','events','--json'],cwd=tmp,capture_output=True,text=True).stdout)['events']
    i=[e['id'] for e in evs].index(st['cut_event'])
    out=[]
    for e in evs[i+1:]:
        if e['type']=='result.record':
            v=e['payload'].get('verdicts') or {}
            out.append('R:'+','.join(k[-6:]+'='+x[0] for k,x in v.items())+('[rev]' if any('review' in x for x in e['payload'].get('evidence') or []) else ''))
        if e['type']=='dq.select': out.append('SELECT '+e['payload']['hypothesis'][-6:])
    print(od.split('/')[-1], ' '.join(out))
