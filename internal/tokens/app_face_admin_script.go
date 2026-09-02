package tokens

import "strconv"

func appFaceAdminScriptBlock() string {
	return `  <script>
    const pathAdminProxies = '` + PathV1AdminProxies + `';
    const pathAdminAvito = '` + PathV1AdminAvitoAccounts + `';
    const adminResource = (base, id) => base + '/' + id;

    function esc(s) {
      return String(s ?? '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/"/g,'&quot;');
    }

    async function loadProxies() {
      const tbody = document.getElementById('` + AppDOMProxiesTable + `');
      const status = document.getElementById('` + AppDOMProxiesFormStatus + `');
      if (!tbody) return;
      const res = await fetch(pathAdminProxies, { credentials: 'same-origin' });
      if (!res.ok) { if (status) status.textContent = '` + AppCopyAdminLoadFailed + `'; return; }
      const data = await res.json();
      const rows = (data.` + JSONFieldProxies + ` || []);
      if (!rows.length) {
        tbody.innerHTML = '<tr><td colspan="4">` + AppCopyAdminEmpty + `</td></tr>';
        return;
      }
      tbody.innerHTML = rows.map(p =>
        '<tr data-id="' + esc(p.` + JSONFieldID + `) + '">' +
        '<td>' + esc(p.` + JSONFieldLabel + `) + '</td>' +
        '<td class="mono">' + esc(p.` + JSONFieldEndpoint + `) + '</td>' +
        '<td><span class="status">' + esc(p.` + JSONFieldStatus + `) + ' / ' + esc(p.` + JSONFieldHealth + ` || '') + '</span></td>' +
        '<td class="row-actions">' +
        '<button type="button" class="btn btn-sm btn-ghost" onclick="editProxy(this)">` + AppCopyAdminEdit + `</button>' +
        '<button type="button" class="btn btn-sm btn-ghost" onclick="deleteProxy(this)">` + AppCopyAdminDelete + `</button>' +
        '</td></tr>'
      ).join('');
    }

    function resetProxyForm() {
      document.getElementById('` + AppDOMProxiesEditID + `').value = '';
      document.getElementById('` + AppDOMProxiesLabel + `').value = '';
      document.getElementById('` + AppDOMProxiesEndpoint + `').value = '';
      document.getElementById('` + AppDOMProxiesStatus + `').value = '` + ProxyStatusValues[0] + `';
      document.getElementById('` + AppDOMProxiesForm + `-submit').textContent = '` + AppCopyAdminCreate + `';
      document.getElementById('` + AppDOMProxiesFormStatus + `').textContent = '';
      return false;
    }

    function editProxy(btn) {
      const tr = btn.closest('tr');
      document.getElementById('` + AppDOMProxiesEditID + `').value = tr.dataset.id;
      document.getElementById('` + AppDOMProxiesLabel + `').value = tr.children[0].textContent;
      document.getElementById('` + AppDOMProxiesEndpoint + `').value = tr.children[1].textContent;
      document.getElementById('` + AppDOMProxiesStatus + `').value = tr.children[2].textContent.trim();
      document.getElementById('` + AppDOMProxiesForm + `-submit').textContent = '` + AppCopyAdminSave + `';
      return false;
    }

    async function saveProxy(e) {
      if (e) e.preventDefault();
      const status = document.getElementById('` + AppDOMProxiesFormStatus + `');
      const editID = document.getElementById('` + AppDOMProxiesEditID + `').value.trim();
      const body = {
        ` + JSONFieldLabel + `: document.getElementById('` + AppDOMProxiesLabel + `').value.trim(),
        ` + JSONFieldEndpoint + `: document.getElementById('` + AppDOMProxiesEndpoint + `').value.trim(),
        ` + JSONFieldStatus + `: document.getElementById('` + AppDOMProxiesStatus + `').value
      };
      const url = editID ? adminResource(pathAdminProxies, editID) : pathAdminProxies;
      const res = await fetch(url, {
        method: editID ? 'PATCH' : 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        credentials: 'same-origin',
        body: JSON.stringify(body)
      });
      if (!res.ok) { status.textContent = '` + AppCopyAdminSaveFailed + `'; return false; }
      resetProxyForm();
      await loadProxies();
      return false;
    }

    async function deleteProxy(btn) {
      if (!confirm('` + AppCopyAdminConfirmDelete + `')) return false;
      const id = btn.closest('tr').dataset.id;
      const status = document.getElementById('` + AppDOMProxiesFormStatus + `');
      const res = await fetch(adminResource(pathAdminProxies, id), { method: 'DELETE', credentials: 'same-origin' });
      if (!res.ok) { status.textContent = '` + AppCopyAdminDeleteFailed + `'; return false; }
      resetProxyForm();
      await loadProxies();
      return false;
    }

    async function loadAvito() {
      const tbody = document.getElementById('` + AppDOMAvitoTable + `');
      const status = document.getElementById('` + AppDOMAvitoFormStatus + `');
      if (!tbody) return;
      const res = await fetch(pathAdminAvito, { credentials: 'same-origin' });
      if (!res.ok) { if (status) status.textContent = '` + AppCopyAdminLoadFailed + `'; return; }
      const data = await res.json();
      const rows = (data.` + JSONFieldAccounts + ` || []);
      if (!rows.length) {
        tbody.innerHTML = '<tr><td colspan="4">` + AppCopyAdminEmpty + `</td></tr>';
        return;
      }
      tbody.innerHTML = rows.map(a =>
        '<tr data-id="' + esc(a.` + JSONFieldID + `) + '">' +
        '<td>' + esc(a.` + JSONFieldLabel + `) + '</td>' +
        '<td class="mono">' + esc(a.` + JSONFieldExternalRef + `) + '</td>' +
        '<td><span class="status">' + esc(a.` + JSONFieldStatus + `) + '</span></td>' +
        '<td class="row-actions">' +
        '<button type="button" class="btn btn-sm btn-ghost" onclick="editAvito(this)">` + AppCopyAdminEdit + `</button>' +
        '<button type="button" class="btn btn-sm btn-ghost" onclick="deleteAvito(this)">` + AppCopyAdminDelete + `</button>' +
        '</td></tr>'
      ).join('');
    }

    function resetAvitoForm() {
      document.getElementById('` + AppDOMAvitoEditID + `').value = '';
      document.getElementById('` + AppDOMAvitoLabel + `').value = '';
      document.getElementById('` + AppDOMAvitoExternalRef + `').value = '';
      document.getElementById('` + AppDOMAvitoStatus + `').value = '` + AvitoAccountStatusValues[0] + `';
      document.getElementById('` + AppDOMAvitoPassword + `').value = '';
      document.getElementById('` + AppDOMAvitoPassword + `').required = true;
      document.getElementById('` + AppDOMAvitoPassword + `-label').textContent = '` + AppCopyAdminPasswordRequired + `';
      document.getElementById('` + AppDOMAvitoForm + `-submit').textContent = '` + AppCopyAdminCreate + `';
      document.getElementById('` + AppDOMAvitoFormStatus + `').textContent = '';
      return false;
    }

    function editAvito(btn) {
      const tr = btn.closest('tr');
      document.getElementById('` + AppDOMAvitoEditID + `').value = tr.dataset.id;
      document.getElementById('` + AppDOMAvitoLabel + `').value = tr.children[0].textContent;
      document.getElementById('` + AppDOMAvitoExternalRef + `').value = tr.children[1].textContent;
      document.getElementById('` + AppDOMAvitoStatus + `').value = tr.children[2].textContent.trim();
      document.getElementById('` + AppDOMAvitoPassword + `').value = '';
      document.getElementById('` + AppDOMAvitoPassword + `').required = false;
      document.getElementById('` + AppDOMAvitoPassword + `-label').textContent = '` + AppCopyAdminPasswordOptional + `';
      document.getElementById('` + AppDOMAvitoForm + `-submit').textContent = '` + AppCopyAdminSave + `';
      return false;
    }

    async function saveAvito(e) {
      if (e) e.preventDefault();
      const status = document.getElementById('` + AppDOMAvitoFormStatus + `');
      const editID = document.getElementById('` + AppDOMAvitoEditID + `').value.trim();
      const body = {
        ` + JSONFieldLabel + `: document.getElementById('` + AppDOMAvitoLabel + `').value.trim(),
        ` + JSONFieldExternalRef + `: document.getElementById('` + AppDOMAvitoExternalRef + `').value.trim(),
        ` + JSONFieldStatus + `: document.getElementById('` + AppDOMAvitoStatus + `').value
      };
      const pwd = document.getElementById('` + AppDOMAvitoPassword + `').value;
      if (!editID || pwd) body.` + JSONFieldPassword + ` = pwd;
      const url = editID ? adminResource(pathAdminAvito, editID) : pathAdminAvito;
      const res = await fetch(url, {
        method: editID ? 'PATCH' : 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        credentials: 'same-origin',
        body: JSON.stringify(body)
      });
      if (!res.ok) { status.textContent = '` + AppCopyAdminSaveFailed + `'; return false; }
      resetAvitoForm();
      await loadAvito();
      return false;
    }

    async function deleteAvito(btn) {
      if (!confirm('` + AppCopyAdminConfirmDelete + `')) return false;
      const id = btn.closest('tr').dataset.id;
      const status = document.getElementById('` + AppDOMAvitoFormStatus + `');
      const res = await fetch(adminResource(pathAdminAvito, id), { method: 'DELETE', credentials: 'same-origin' });
      if (!res.ok) { status.textContent = '` + AppCopyAdminDeleteFailed + `'; return false; }
      resetAvitoForm();
      await loadAvito();
      return false;
    }

    loadProxies();
    loadAvito();
    setInterval(() => { loadProxies(); loadAvito(); }, ` + strconv.Itoa(AdminLivePollIntervalMs) + `);
    const adminNav = document.getElementById('` + AppDOMNav + `');
    if (adminNav) adminNav.addEventListener('click', (e) => {
      const a = e.target.closest('a[data-view]');
      if (!a) return;
      if (a.dataset.view === '` + AppNavIDProxies + `') loadProxies();
      if (a.dataset.view === '` + AppNavIDAvito + `') loadAvito();
    });
  </script>
`
}
