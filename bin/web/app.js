const API = '/api';

let state = {
    activeCaseId: null,
    cases: []
};

const els = {
    search: document.getElementById('search'),
    status: document.getElementById('status'),
    date1: document.getElementById('date1'),
    date2: document.getElementById('date2'),
    list: document.getElementById('list'),
    rightBlock: document.getElementById('right-block')
};

function escapeHTML(str) {
    if (!str) return '';
    return str.replace(/[&<>'"]/g, 
        tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag])
    );
}

function debounce(func, wait) {
    let timeout;
    function debounced(...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), wait);
    }
    debounced.cancel = () => clearTimeout(timeout);
    return debounced;
}

function formatDate(isoString, includeTime = false) {
    if (!isoString) return '';
    const date = new Date(isoString);
    const options = { day: '2-digit', month: '2-digit', year: 'numeric' };
    if (includeTime) {
        options.hour = '2-digit';
        options.minute = '2-digit';
    }
    return date.toLocaleString('ru-RU', options);
}

function toLocalDatetime(isoString) {
    if (!isoString) return '';
    const date = new Date(isoString);
    return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
}

async function request(url, options = {}) {
    const response = await fetch(url, options);
    if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(error.error || 'Ошибка сервера');
    }
    if (response.status === 204) return null;
    return response.json();
}

async function loadCases() {
    const params = new URLSearchParams();
    if (els.search.value) params.append('q', els.search.value);
    if (els.status.value) params.append('status', els.status.value);
    
    if (els.date1.value) params.append('from', new Date(els.date1.value + 'T00:00:00').toISOString());
    if (els.date2.value) params.append('to', new Date(els.date2.value + 'T23:59:59').toISOString());

    try {
        const cases = await request(`${API}/cases?${params}`);
        state.cases = cases;
        renderList();
    } catch (error) {
        els.list.innerHTML = `<p style="color:red;">Ошибка: ${error.message}</p>`;
    }
}

async function showCase(id) {
    state.activeCaseId = id;
    renderList();

    try {
        const c = await request(`${API}/cases/${id}`);
        renderCaseDetails(c);
    } catch (error) {
        els.rightBlock.innerHTML = `<p style="color:red;">Ошибка: ${error.message}</p>`;
    }
}

async function changeStatus(id, status) {
    try {
        if (status === 'open') {
            await request(`${API}/cases/${id}/reopen`, { method: 'POST' });
        } else {
            await request(`${API}/cases/${id}/close`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status })
            });
        }
        await showCase(id);
        await loadCases();
    } catch (error) {
        alert('Ошибка смены статуса: ' + error.message);
    }
}

async function addComment(id) {
    const input = document.getElementById('comment-text');
    const text = input.value.trim();
    if (!text) return;

    try {
        await request(`${API}/cases/${id}/comments`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ body: text })
        });
        input.value = '';
        await loadComments(id);
    } catch (error) {
        alert('Ошибка добавления комментария: ' + error.message);
    }
}

async function loadComments(id) {
    try {
        const comments = await request(`${API}/cases/${id}/comments`);
        renderComments(comments);
    } catch (error) {
        console.error('Ошибка загрузки комментариев:', error);
    }
}

function renderComments(comments) {
    const container = document.getElementById('comments');
    if (!container) return;

    if (comments.length === 0) {
        container.innerHTML = '<span style="color:gray;">Пока нет комментариев</span>';
        return;
    }

    container.innerHTML = comments.map(c => `
        <div style="border-bottom: 1px solid #eee; padding: 5px 0;">
            <small style="color: gray;">${formatDate(c.created_at, true)}</small><br>
            ${escapeHTML(c.body)}
        </div>
    `).join('');
}

async function createNewDraft() {
    try {
        const res = await request(`${API}/cases`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                title: '',
                description: '',
                status: 'draft',
                promised_at: new Date().toISOString(),
                due_at: new Date().toISOString()
            })
        });
        
        await loadCases();
        showForm(res.id); 
    } catch (error) {
        alert('Не удалось создать черновик: ' + error.message);
    }
}

function renderList() {
    if (state.cases.length === 0) {
        els.list.innerHTML = '<p style="color:gray;">Ничего не найдено</p>';
        return;
    }

    const fragment = document.createDocumentFragment();

    state.cases.forEach(c => {
        const div = document.createElement('div');
        div.className = c.id === state.activeCaseId ? 'item item-active' : 'item';
        
        let statusText = 'В работе';
        let statusColor = 'blue';
        if (c.status === 'done') { statusText = 'Готово'; statusColor = 'green'; }
        if (c.status === 'cancelled') { statusText = 'Отменено'; statusColor = 'red'; }
        if (c.status === 'draft') { statusText = 'Черновик'; statusColor = 'orange'; }

        div.innerHTML = `
            <b>${escapeHTML(c.title || 'Без названия')}</b><br>
            До: ${formatDate(c.due_at)}<br>
            <span class="${statusColor}">${statusText}</span>
        `;
        
        div.addEventListener('click', () => {
            if (c.status === 'draft') {
                showForm(c.id); 
            } else {
                showCase(c.id); 
            }
        });
        fragment.appendChild(div);
    });

    els.list.innerHTML = '';
    els.list.appendChild(fragment);
}

function renderCaseDetails(c) {
    let actionButtons = '';
    if (c.status === 'open') {
        actionButtons = `
            <button onclick="window.changeStatus(${c.id}, 'done')">Выполнить</button>
            <button onclick="window.changeStatus(${c.id}, 'cancelled')">Отменить</button>
        `;
    } else {
        actionButtons = `<button onclick="window.changeStatus(${c.id}, 'open')">Вернуть</button>`;
    }

    els.rightBlock.innerHTML = `
        <h3>${escapeHTML(c.title)}</h3>
        <p>${escapeHTML(c.description || 'Нет описания')}</p>
        <p><b>Обещано:</b> ${formatDate(c.promised_at, true)}</p>
        <p><b>Срок:</b> ${formatDate(c.due_at, true)}</p>
        
        <div style="margin: 15px 0;">
            <button onclick="window.showForm(${c.id})">Изменить</button>
            ${actionButtons}
        </div>

        <h4>Комментарии:</h4>
        <div id="comments" style="margin-bottom: 15px;">Загрузка...</div>
        
        <input type="text" id="comment-text" placeholder="Текст комментария...">
        <button onclick="window.addComment(${c.id})">Отправить</button>
    `;
    
    loadComments(c.id);
}

async function showForm(id) {
    state.activeCaseId = id;
    renderList();

    let caseData;
    let formTitle = 'Редактирование';

    try {
        caseData = await request(`${API}/cases/${id}`);
        if (caseData.status === 'draft') {
            formTitle = 'Новое дело (Черновик)';
        }
    } catch (error) {
        alert('Не удалось загрузить данные дела');
        return;
    }

    const d1 = toLocalDatetime(caseData.promised_at);
    const d2 = toLocalDatetime(caseData.due_at);

    els.rightBlock.innerHTML = `
        <h3>${formTitle}</h3>
        <form id="case-form" onsubmit="window.saveData(event, ${id})">
            <label>Название:</label>
            <input type="text" id="f-title" value="${escapeHTML(caseData.title)}">
            
            <label>Описание:</label>
            <textarea id="f-desc" rows="6">${escapeHTML(caseData.description || '')}</textarea>
            
            <label>Обещано:</label>
            <input type="datetime-local" id="f-date1" value="${d1}">
            
            <label>Срок:</label>
            <input type="datetime-local" id="f-date2" value="${d2}">
            
            <div style="margin-top: 10px;">
                <button type="submit">Сохранить дело</button>
                <button type="button" onclick="window.cancelForm(${id})">Назад</button>
            </div>
        </form>
    `;

    const form = document.getElementById('case-form');
    const fTitle = document.getElementById('f-title');
    const fDesc = document.getElementById('f-desc');
    const fDate1 = document.getElementById('f-date1');
    const fDate2 = document.getElementById('f-date2');

    if (caseData.status === 'draft') {
        const autoSaveToDB = debounce(async () => {
            const data = {
                title: fTitle.value,
                description: fDesc.value,
                status: 'draft',
                promised_at: new Date(fDate1.value || Date.now()).toISOString(),
                due_at: new Date(fDate2.value || Date.now()).toISOString()
            };
            
            try {
                await request(`${API}/cases/${id}`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                loadCases(); 
            } catch (e) {
                console.error('Ошибка автосохранения', e);
            }
        }, 1000);

        form.addEventListener('input', autoSaveToDB);
        form._autoSave = autoSaveToDB;
    }
}

async function saveData(event, id) {
    event.preventDefault();
    
    const form = document.getElementById('case-form');
    if (form && form._autoSave) {
        form._autoSave.cancel();
    }
    
    const title = document.getElementById('f-title').value.trim();
    if (!title) {
        alert('Название не может быть пустым');
        return;
    }

    const data = {
        title: title,
        description: document.getElementById('f-desc').value,
        status: 'open',
        promised_at: new Date(document.getElementById('f-date1').value).toISOString(),
        due_at: new Date(document.getElementById('f-date2').value).toISOString()
    };

    try {
        await request(`${API}/cases/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        await loadCases();
        await showCase(id);

    } catch (error) {
        alert('Ошибка сохранения: ' + error.message);
    }
}

function cancelForm(id) {
    showCase(id);
}

window.loadCases = loadCases;
window.createNewDraft = createNewDraft;
window.showForm = showForm;
window.showCase = showCase;
window.changeStatus = changeStatus;
window.addComment = addComment; 
window.saveData = saveData;
window.cancelForm = cancelForm;

els.search.addEventListener('input', debounce(loadCases, 300));

loadCases();