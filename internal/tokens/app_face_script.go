package tokens

import "strconv"

func appFaceScriptBlock() string {
	return `  <script>
    const tick = document.getElementById('` + AppDOMTick + `');
    const hint = document.getElementById('` + AppDOMAuthHint + `');
    const pathRequest = '` + PathV1AuthMagicLink + `';
    const pathConsume = '` + PathV1AuthMagicLinkConsume + `';
    const pathLogout = '` + PathV1AuthLogout + `';
    const pathRoot = '` + PathRoot + `';
    async function enterApp(e) {
      if (e) e.preventDefault();
      const email = document.getElementById('` + AppDOMEmailInput + `').value.trim();
      if (!email.includes('@')) return false;
      const res = await fetch(pathRequest, {
        method: 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        body: JSON.stringify({ ` + JSONFieldEmail + `: email })
      });
      if (!res.ok) { hint.textContent = '` + AppCopyRequestFailed + `'; return false; }
      const data = await res.json();
      if (data.` + JSONFieldMagicLinkURL + `) {
        window.location.href = data.` + JSONFieldMagicLinkURL + `;
        return false;
      }
      if (data.` + JSONFieldToken + `) {
        const creq = await fetch(pathConsume, {
          method: 'POST',
          headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
          body: JSON.stringify({ ` + JSONFieldToken + `: data.` + JSONFieldToken + ` }),
          credentials: 'same-origin'
        });
        if (!creq.ok) { hint.textContent = '` + AppCopyConsumeFailed + `'; return false; }
        window.location.href = pathRoot;
        return false;
      }
      hint.textContent = '` + AppCopySentHint + `';
      return false;
    }
    async function backAuth() {
      try {
        await fetch(pathLogout, { method: 'POST', credentials: 'same-origin' });
      } catch (_) {}
      window.location.href = pathRoot;
      return false;
    }
    const nav = document.getElementById('` + AppDOMNav + `');
    if (nav) nav.addEventListener('click', (e) => {
      const a = e.target.closest('a[data-view]');
      if (!a) return;
      e.preventDefault();
      document.querySelectorAll('#` + AppDOMNav + ` a').forEach(x => x.classList.toggle('` + AppClassIsActive + `', x === a));
      document.querySelectorAll('.view').forEach(v => v.classList.toggle('` + AppClassIsActive + `', v.id === '` + AppDOMViewPrefix + `' + a.dataset.view));
    });
    if (tick) setInterval(() => {
      tick.textContent = new Date().toLocaleTimeString('` + LocaleBCP47 + `', { hour12: false });
    }, 1000);
  </script>
`
}

func appFaceSearchScriptBlock() string {
	return `  <script>
    const pathMeTasks = '` + PathV1MeTasks + `';
    const pathTaskResults = (id) => '` + PathV1Tasks + `/' + id + '` + PathV1TaskResultsSuffix + `';
    const pathWatchResults = (id) => '` + PathV1MeWatches + `/' + id + '` + PathV1WatchResultsSuffix + `';
    let watchPollTimer = null;
    function renderResults(box, rows) {
      box.innerHTML = '<table><thead><tr><th>` + AppCopyColCode + `</th><th>` + AppCopyColTitle + `</th></tr></thead><tbody>' + rows + '</tbody></table>';
    }
    async function loadWatchResults(watchID, status) {
      const box = document.getElementById('` + AppDOMSearchResults + `');
      const res = await fetch(pathWatchResults(watchID), { credentials: 'same-origin' });
      if (!res.ok) return;
      const data = await res.json();
      const rows = (data.` + JSONFieldResults + ` || []).map(it =>
        '<tr><td class="mono">' + it.` + JSONFieldAvitoID + ` + '</td><td>' + it.` + JSONFieldTitle + ` + '</td></tr>'
      ).join('');
      renderResults(box, rows);
      if (status) status.textContent = '` + AppCopySearchStatusWatch + `';
    }
    async function submitSearch(e) {
      if (e) e.preventDefault();
      const url = document.getElementById('` + AppDOMSearchURL + `').value.trim();
      const status = document.getElementById('` + AppDOMSearchStatus + `');
      const box = document.getElementById('` + AppDOMSearchResults + `');
      status.textContent = '…';
      box.innerHTML = '';
      if (watchPollTimer) { clearInterval(watchPollTimer); watchPollTimer = null; }
      const res = await fetch(pathMeTasks, {
        method: 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        credentials: 'same-origin',
        body: JSON.stringify({ ` + JSONFieldQuery + `: url })
      });
      if (!res.ok) { status.textContent = '` + AppCopySearchFailed + `'; return false; }
      const task = await res.json();
      if (task.` + JSONFieldKind + ` === '` + ListingSearchKindWatch + `') {
        await loadWatchResults(task.` + JSONFieldID + `, status);
        watchPollTimer = setInterval(() => loadWatchResults(task.` + JSONFieldID + `, status), ` + strconv.Itoa(ListingSearchWatchPollIntervalMs) + `);
        return false;
      }
      status.textContent = task.` + JSONFieldStatus + `;
      pollResults(task.` + JSONFieldID + `);
      return false;
    }
    async function pollResults(taskID) {
      const status = document.getElementById('` + AppDOMSearchStatus + `');
      const box = document.getElementById('` + AppDOMSearchResults + `');
      for (let i = 0; i < ` + strconv.Itoa(ListingSearchPollMaxAttempts) + `; i++) {
        const t = await fetch('` + PathV1Tasks + `/' + taskID, { credentials: 'same-origin' });
        if (t.ok) {
          const task = await t.json();
          status.textContent = task.` + JSONFieldStatus + `;
          if (task.` + JSONFieldStatus + ` === '` + TaskStatusCompleted + `' || task.` + JSONFieldStatus + ` === '` + TaskStatusFailed + `') break;
        }
        await new Promise(r => setTimeout(r, ` + strconv.Itoa(ListingSearchPollIntervalMs) + `));
      }
      const res = await fetch(pathTaskResults(taskID), { credentials: 'same-origin' });
      if (!res.ok) return;
      const data = await res.json();
      const rows = (data.` + JSONFieldResults + ` || []).map(it =>
        '<tr><td class="mono">' + it.` + JSONFieldAvitoID + ` + '</td><td>' + it.` + JSONFieldTitle + ` + '</td></tr>'
      ).join('');
      renderResults(box, rows);
    }
  </script>
`
}
