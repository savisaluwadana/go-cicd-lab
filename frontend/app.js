const $ = sel => document.querySelector(sel)
const $$ = sel => document.querySelectorAll(sel)

async function api(path, opts={}){
  const res = await fetch(path, opts)
  if (res.status === 204) return null
  return res.json()
}

async function load(){
  const books = await api('/api/books')
  const tbody = $('#books tbody')
  tbody.innerHTML = ''
  books.forEach(b=>{
    const tr = document.createElement('tr')
    tr.innerHTML = `<td>${b.id}</td><td>${b.title}</td><td>${b.author}</td><td>${b.year}</td><td><a class="action" data-id="${b.id}">Edit</a> | <a class="action del" data-id="${b.id}">Delete</a></td>`
    tbody.appendChild(tr)
  })
  $$('.action').forEach(a=>a.onclick = e=>{
    const id = e.target.dataset.id
    const row = [...$('#books tbody tr')].find(r=>r.children[0].textContent===id)
    if (e.target.classList.contains('del')){
      if(confirm('Delete book '+id+'?')){ api(`/api/book?id=${id}`,{method:'DELETE'}).then(load) }
      return
    }
    $('#id').value = id
    $('#title').value = row.children[1].textContent
    $('#author').value = row.children[2].textContent
    $('#year').value = row.children[3].textContent
  })
}

$('#create').onclick = async ()=>{
  const payload = { id: $('#id').value.trim(), title: $('#title').value.trim(), author: $('#author').value.trim(), year: Number($('#year').value) }
  if(!payload.id){ alert('ID is required'); return }
  // try create; if exists, use PUT to update
  try{
    await api('/api/book?id='+encodeURIComponent(payload.id))
    // exists -> update
    await api('/api/books',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  }catch(e){
    // create
    await api('/api/books',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
  }
  $('#clear').click()
  load()
}

$('#clear').onclick = ()=>{ $('#id').value=''; $('#title').value=''; $('#author').value=''; $('#year').value='' }

load()
