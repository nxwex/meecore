const $=s=>document.querySelector(s);

const S={
    instances:[],
    containers:[],
    templates:[],
    nodes:[]
};

async function api(path,opt={}){
    const r=await fetch(path,{
        headers:{
            "Content-Type":"application/json",
            ...(opt.headers||{})
        },
        ...opt
    });

    const t=await r.text();

    let d=null;

    try{
        d=t?JSON.parse(t):null
    }catch{
        d=t
    }

    if(!r.ok){
        throw Error(
            typeof d==='string'
                ?d
                :(d?.error||d?.message||`HTTP ${r.status}`)
        )
    }

    return d
}

const arr=(d,ks=[])=>Array.isArray(d)
    ?d
    :ks.map(k=>d?.[k]).find(Array.isArray)||[];

const cid=c=>c?.id??c?.ID??c?.container_id??c?.Id??'';

const iid=i=>i?.id??i?.ID??'';

const icid=i=>i?.container_id??i?.ContainerID??'';

const status=c=>String(
    c?.status??
    c?.Status??
    c?.state??
    c?.State??
    ''
).toLowerCase();

const esc=v=>String(v??'').replace(
    /[&<>"']/g,
    x=>({
        '&':'&amp;',
        '<':'&lt;',
        '>':'&gt;',
        '"':'&quot;',
        "'":'&#39;'
    }[x])
);

function toast(m,err=false){
    const x=$('#toast');

    x.textContent=m;
    x.className='toast'+(err?' error':'');
    x.classList.remove('hidden');

    clearTimeout(toast.t);

    toast.t=setTimeout(
        ()=>x.classList.add('hidden'),
        3000
    );
}

function date(v){
    if(!v)return'—';

    const d=new Date(v);

    return isNaN(d)
        ?'—'
        :d.toLocaleString([],{
            dateStyle:'medium',
            timeStyle:'short'
        });
}

function findC(i){
    const id=icid(i);

    return S.containers.find(
        c=>String(cid(c))===String(id)||
        (
            cid(c)&&
            id&&
            (
                String(cid(c)).startsWith(String(id))||
                String(id).startsWith(String(cid(c)))
            )
        )
    );
}

function badge(c){
    const s=status(c);
    const run=s.includes('running');

    return `<span class="badge ${run?'running':''}">${esc(
        run?'running':(s||'unknown')
    )}</span>`;
}

async function load(){
    try{
        const [
            i,
            c,
            t
        ]=await Promise.all([
            api('/api/instances'),
            api('/api/containers'),
            api('/api/templates')
        ]);

        S.instances=arr(i,['instances','data']).filter(Boolean);
        S.containers=arr(c,['containers','data']).filter(Boolean);
        S.templates=arr(t,['templates','data']).filter(Boolean);

        renderInstances();
        renderContainers();
        renderTemplates();

        await nodes();
    }catch(e){
        $('#instancesState').textContent=e.message;
        toast(e.message,true);
    }
}

async function nodes(){
    try{
        S.nodes=arr(
            await api('/api/nodes'),
            ['nodes','data']
        );

        populateNodes();
    }catch{
        S.nodes=[];

        $('#nodeSelect').innerHTML=
            '<option value="">Nodes API unavailable</option>';

        $('#submitCreate').disabled=true;
    }
}

function populateNodes(){
    const x=$('#nodeSelect');

    if(!S.nodes.length){
        x.innerHTML=
            '<option value="">No nodes available</option>';

        $('#submitCreate').disabled=true;

        return;
    }

    $('#submitCreate').disabled=false;

    x.innerHTML=S.nodes.map(
        n=>`<option value="${esc(n.id??n.ID)}">${
            esc(n.name??n.Name??`Node ${n.id??n.ID}`)
        }</option>`
    ).join('');
}

function populateTemplates(){
    const x=$('#templateSelect');

    x.innerHTML=
        '<option value="">Select template...</option>'+
        S.templates.map(
            t=>`<option value="${esc(t.name??t.Name??'')}">${
                esc(
                    t.display_name??
                    t.DisplayName??
                    t.name??
                    t.Name
                )
            }</option>`
        ).join('')+
        '<option value="__custom">Custom</option>';
}

function renderInstances(){
    const list=S.instances;

    $('#instanceCount').textContent=list.length;

    let r=0;

    for(const i of list){
        if(status(findC(i)).includes('running')){
            r++;
        }
    }

    $('#runningCount').textContent=r;
    $('#stoppedCount').textContent=list.length-r;

    const body=$('#instancesBody');

    if(!list.length){
        $('#instancesState').textContent='No instances yet.';
        $('#instancesState').classList.remove('hidden');
        $('#instancesTableWrap').classList.add('hidden');

        body.innerHTML='';

        return;
    }

    $('#instancesState').classList.add('hidden');
    $('#instancesTableWrap').classList.remove('hidden');

    body.innerHTML=list.map(i=>{
        const c=findC(i);
        const id=icid(i);
        const iidv=iid(i);
        const run=status(c).includes('running');

        return `
            <tr>
                <td>
                    <strong>
                        ${esc(i.name??i.Name??'Unnamed')}
                    </strong>
                </td>

                <td>
                    ${esc(
                        i.node_name??
                        i.NodeName??
                        i.node_id??
                        i.NodeID??
                        '—'
                    )}
                </td>

                <td>
                    ${esc(i.template??i.Template??'—')}
                </td>

                <td>
                    <code>
                        ${esc(String(id).slice(0,12)||'—')}
                    </code>
                </td>

                <td>
                    ${badge(c)}
                </td>

                <td>
                    ${esc(date(i.created_at??i.CreatedAt))}
                </td>

                <td>
                    <div class="actions">

                        <button
                            class="mini"
                            data-a="console"
                            data-i="${esc(iidv)}"
                        >
                            Console
                        </button>

                        <button
                            class="mini"
                            data-a="start"
                            data-c="${esc(id)}"
                            ${!id||run?'disabled':''}
                        >
                            Start
                        </button>

                        <button
                            class="mini"
                            data-a="stop"
                            data-c="${esc(id)}"
                            ${!id||!run?'disabled':''}
                        >
                            Stop
                        </button>

                        <button
                            class="mini"
                            data-a="restart"
                            data-c="${esc(id)}"
                            ${!id?'disabled':''}
                        >
                            Restart
                        </button>

                        <button
                            class="mini danger"
                            data-a="delete-i"
                            data-i="${esc(iidv)}"
                        >
                            Delete
                        </button>

                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

function renderContainers(){
    const b=$('#containersBody');

    if(!S.containers.length){
        $('#containersState').textContent='No containers.';
        $('#containersState').classList.remove('hidden');
        $('#containersTableWrap').classList.add('hidden');

        return;
    }

    $('#containersState').classList.add('hidden');
    $('#containersTableWrap').classList.remove('hidden');

    b.innerHTML=S.containers.map(c=>{
        const id=cid(c);

        return `
            <tr>
                <td>
                    <strong>
                        ${esc(c.name??c.Name??'—')}
                    </strong>
                </td>

                <td>
                    <code>
                        ${esc(String(id).slice(0,12))}
                    </code>
                </td>

                <td>
                    ${badge(c)}
                </td>

                <td>
                    ${esc(c.image??c.Image??'—')}
                </td>

                <td>
                    <div class="actions">

                        <button
                            class="mini"
                            data-a="start"
                            data-c="${esc(id)}"
                        >
                            Start
                        </button>

                        <button
                            class="mini"
                            data-a="stop"
                            data-c="${esc(id)}"
                        >
                            Stop
                        </button>

                        <button
                            class="mini"
                            data-a="restart"
                            data-c="${esc(id)}"
                        >
                            Restart
                        </button>

                        <button
                            class="mini danger"
                            data-a="delete-c"
                            data-c="${esc(id)}"
                        >
                            Delete
                        </button>

                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

function renderTemplates(){
    $('#templatesGrid').innerHTML=S.templates.length
        ?S.templates.map(
            t=>`
                <div class="template">
                    <h3>
                        ${esc(
                            t.display_name??
                            t.DisplayName??
                            t.name??
                            t.Name
                        )}
                    </h3>

                    <p>
                        ${esc(t.name??t.Name??'')}
                    </p>

                    <code>
                        ${esc(t.image??t.Image??'')}
                    </code>
                </div>
            `
        ).join('')
        :'<div class="state">No templates.</div>';

    populateTemplates();
}

async function action(a,c){
    if(!c)return;

    try{
        await api(
            `/api/containers/${encodeURIComponent(c)}/${a}`,
            {
                method:'POST'
            }
        );

        toast(a+' completed');

        await load();
    }catch(e){
        toast(e.message,true);
    }
}

async function delI(id){
    if(!id||!confirm('Delete this instance?')){
        return;
    }

    try{
        await api(
            `/api/instances/${encodeURIComponent(id)}`,
            {
                method:'DELETE'
            }
        );

        toast('Instance deleted');

        await load();
    }catch(e){
        toast(e.message,true);
    }
}

async function delC(id){
    if(!id||!confirm('Delete this container?')){
        return;
    }

    try{
        await api(
            `/api/containers/${encodeURIComponent(id)}`,
            {
                method:'DELETE'
            }
        );

        toast('Container deleted');

        await load();
    }catch(e){
        toast(e.message,true);
    }
}

function open(){
    populateNodes();
    populateTemplates();

    $('#createForm').reset();
    $('#formError').classList.add('hidden');
    $('#customNote').classList.add('hidden');
    $('#createModal').classList.remove('hidden');
}

function close(){
    $('#createModal').classList.add('hidden');
}

$('#createForm').addEventListener(
    'submit',
    async e=>{
        e.preventDefault();

        const err=$('#formError');
        const template=$('#templateSelect').value;

        err.classList.add('hidden');

        if(template==='__custom'){
            err.textContent=
                'Custom creation is not supported by the current backend yet.';

            err.classList.remove('hidden');

            return;
        }

        try{
            $('#submitCreate').disabled=true;

            await api(
                '/api/instances',
                {
                    method:'POST',
                    body:JSON.stringify({
                        node_id:Number($('#nodeSelect').value),
                        name:$('#nameInput').value.trim(),
                        template
                    })
                }
            );

            close();

            toast('Instance created');

            await load();
        }catch(x){
            err.textContent=x.message;
            err.classList.remove('hidden');
        }finally{
            $('#submitCreate').disabled=false;
        }
    }
);

$('#templateSelect').addEventListener(
    'change',
    e=>$('#customNote').classList.toggle(
        'hidden',
        e.target.value!=='__custom'
    )
);

document.addEventListener(
    'click',
    e=>{
        const n=e.target.closest('.nav-item');

        if(n){
            document
                .querySelectorAll('.nav-item')
                .forEach(x=>x.classList.remove('active'));

            n.classList.add('active');

            document
                .querySelectorAll('.page')
                .forEach(x=>x.classList.add('hidden'));

            $(`#${n.dataset.page}Page`).classList.remove('hidden');

            $('#pageTitle').textContent=
                n.dataset.page[0].toUpperCase()+
                n.dataset.page.slice(1);
        }

        const a=e.target.closest('[data-a]');

        if(a){
            if(a.dataset.a==='console'){
                window.location.href=
                    `/console/${encodeURIComponent(a.dataset.i)}`;

                return;
            }

            if(a.dataset.a==='delete-i'){
                delI(a.dataset.i);

                return;
            }

            if(a.dataset.a==='delete-c'){
                delC(a.dataset.c);

                return;
            }

            action(
                a.dataset.a,
                a.dataset.c
            );
        }

        if(e.target.matches('[data-close-modal]')){
            close();
        }
    }
);

$('#createBtn').onclick=open;
$('#refreshBtn').onclick=load;

async function health(){
    try{
        await api('/health');

        $('#healthDot').className='ok';
        $('#healthText').textContent='Connected';
    }catch{
        $('#healthDot').className='bad';
        $('#healthText').textContent='Offline';
    }
}

health();
load();