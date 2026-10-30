package tokens

import (
	"encoding/json"
	"strconv"
)

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
	regionLabelsJSON, _ := json.Marshal(WatchRegionLabels)
	return `  <script>
    const watchRegionLabels = ` + string(regionLabelsJSON) + `;
    const pathMeTasks = '` + PathV1MeTasks + `';
    const pathMeWatches = '` + PathV1MeWatches + `';
    const pathTaskResults = (id) => '` + PathV1Tasks + `/' + id + '` + PathV1TaskResultsSuffix + `';
    const pathWatchResults = (id) => '` + PathV1MeWatches + `/' + id + '` + PathV1WatchResultsSuffix + `';
    const pathWatchReset = (id) => '` + PathV1MeWatches + `/' + id + '` + PathV1WatchResetBaselineSuffix + `';
    const pathWatch = (id) => '` + PathV1MeWatches + `/' + id;
    let watchPollTimer = null;
    let activeWatchID = null;
    function esc(s) {
      return String(s ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
    }
    function formatResultRow(it) {
      const time = it.` + JSONFieldFoundAt + ` ? fmtTime(it.` + JSONFieldFoundAt + `) : '';
      const title = esc(it.` + JSONFieldTitle + ` || '');
      const url = it.` + JSONFieldURL + ` || '';
      const link = url ? '<a href="' + esc(url) + '" target="_blank" rel="noopener noreferrer">` + AppCopyLinkOpen + `</a>' : '';
      return '<tr><td>' + esc(time) + '</td><td>' + title + '</td><td>' + link + '</td></tr>';
    }
    function renderResults(box, rows) {
      box.innerHTML = '<table><thead><tr><th>` + AppCopyColFoundAt + `</th><th>` + AppCopyColDescription + `</th><th>` + AppCopyColLink + `</th></tr></thead><tbody>' + rows + '</tbody></table>';
    }
    function humanizeSlug(s) {
      return String(s || '').split('-').filter(Boolean).map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    }
    function regionLabel(slug) {
      return watchRegionLabels[slug] || humanizeSlug(slug) || '` + AppCopyWatchDetailUnknown + `';
    }
    function fmtTime(iso) {
      if (!iso) return '';
      try {
        return new Date(iso).toLocaleString('` + LocaleBCP47 + `', { hour12: false });
      } catch (_) { return String(iso); }
    }
    function humanizeWatchError(msg) {
      const m = String(msg || '').toLowerCase();
      if (!m) return '';
      if (m.includes('` + ErrMsgListingSearchRateLimited + `')) return '` + AppCopyWatchErrRateLimited + `';
      if (m.includes('` + ErrMsgListingSearchAvitoEmpty + `')) return '` + AppCopyWatchErrEmpty + `';
      if (m.includes('` + ErrMsgListingSearchAvitoFetch + `')) return '` + AppCopyWatchErrFetch + `';
      if (m.includes('` + ErrMsgListingSearchNoBinding + `')) return '` + AppCopyWatchErrNoBinding + `';
      return String(msg);
    }
    function watchStatusLabel(status) {
      switch (status) {
        case '` + ListingWatchStatusActive + `': return '` + AppCopyWatchStatusActive + `';
        case '` + ListingWatchStatusPaused + `': return '` + AppCopyWatchStatusPaused + `';
        case '` + ListingWatchStatusDisabled + `': return '` + AppCopyWatchStatusDisabled + `';
        default: return status || '` + AppCopyWatchDetailUnknown + `';
      }
    }
    function watchMetaStatusLabel(st) {
      switch (st) {
        case '` + WatchMetaStatusPending + `': return '` + AppCopyWatchMetaPending + `';
        case '` + WatchMetaStatusReady + `': return '` + AppCopyWatchMetaReady + `';
        case '` + WatchMetaStatusFailed + `': return '` + AppCopyWatchMetaFailed + `';
        default: return st || '` + AppCopyWatchDetailUnknown + `';
      }
    }
    function formatPrice(extras, params) {
      const from = extras.from ?? extras.fromPrice ?? params.from;
      const to = extras.to ?? extras.toPrice ?? params.to;
      const parts = [];
      if (from !== undefined && from !== null && String(from) !== '') {
        parts.push('` + AppCopyWatchDetailPriceFrom + ` ' + Number(from).toLocaleString('` + LocaleBCP47 + `') + ' ` + AppCopyWatchDetailPriceRub + `');
      }
      if (to !== undefined && to !== null && String(to) !== '') {
        parts.push('` + AppCopyWatchDetailPriceTo + ` ' + Number(to).toLocaleString('` + LocaleBCP47 + `') + ' ` + AppCopyWatchDetailPriceRub + `');
      }
      return parts.join(' · ') || '';
    }
    function formatWatchMeta(w) {
      const lines = [];
      const cats = (w.` + JSONFieldCategories + ` || []).map(humanizeSlug).join(' → ');
      if (w.` + JSONFieldRegion + `) lines.push('<strong>` + AppCopyWatchDetailRegion + `:</strong> ' + esc(regionLabel(w.` + JSONFieldRegion + `)));
      if (cats) lines.push('<strong>` + AppCopyWatchDetailCategory + `:</strong> ' + esc(cats));
      if (w.` + JSONFieldLabel + `) lines.push('<strong>` + AppCopyWatchDetailBrand + `:</strong> ' + esc(humanizeSlug(w.` + JSONFieldLabel + `)));
      const price = formatPrice(w.` + JSONFieldExtras + ` || {}, w.` + JSONFieldParams + ` || {});
      if (price) lines.push('<strong>` + AppCopyWatchDetailPrice + `:</strong> ' + esc(price));
      lines.push('<strong>` + AppCopyWatchDetailMonitor + `:</strong> ' + esc(watchStatusLabel(w.` + JSONFieldStatus + `)));
      lines.push('<strong>` + AppCopyWatchDetailMeta + `:</strong> ' + esc(watchMetaStatusLabel(w.` + JSONFieldMetaStatus + `)));
      const pollOK = !w.` + JSONFieldLastError + ` && (w.` + JSONFieldConsecutiveFailures + ` || 0) === 0;
      lines.push('<strong>` + AppCopyWatchDetailPoll + `:</strong> ' + esc(pollOK ? '` + AppCopyWatchDetailPollOK + `' : '` + AppCopyWatchDetailPollFail + `'));
      if (w.` + JSONFieldLastSuccessAt + `) {
        lines.push('<strong>` + AppCopyWatchDetailLastOK + `:</strong> ' + esc(fmtTime(w.` + JSONFieldLastSuccessAt + `)));
      }
      if (w.` + JSONFieldLastError + `) {
        lines.push('<strong>` + AppCopyWatchDetailLastFail + `:</strong> ' + esc(humanizeWatchError(w.` + JSONFieldLastError + `)));
      }
      if ((w.` + JSONFieldConsecutiveFailures + ` || 0) > 0) {
        lines.push('<strong>` + AppCopyWatchDetailFailStreak + `:</strong> ' + esc(String(w.` + JSONFieldConsecutiveFailures + `)));
      }
      if (w.` + JSONFieldQuery + `) {
        lines.push('<strong>` + AppCopyWatchDetailLink + `:</strong> <span class="mono">' + esc(w.` + JSONFieldQuery + `) + '</span>');
      }
      return lines.join('<br>');
    }
    async function patchWatch(id, status) {
      await fetch(pathWatch(id), {
        method: 'PATCH',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        credentials: 'same-origin',
        body: JSON.stringify({ ` + JSONFieldStatus + `: status })
      });
      await loadWatches(activeWatchID);
    }
    async function resetWatch(id) {
      await fetch(pathWatchReset(id), { method: 'POST', credentials: 'same-origin' });
      await loadWatches(id);
      if (activeWatchID === id) selectWatch(id);
    }
    async function stopWatch(id) {
      await fetch(pathWatch(id), { method: 'DELETE', credentials: 'same-origin' });
      if (activeWatchID === id) activeWatchID = null;
      await loadWatches();
    }
    function selectWatch(watchID) {
      activeWatchID = watchID;
      document.querySelectorAll('#` + AppDOMSearchWatches + ` .watch').forEach(el => {
        el.classList.toggle('` + AppClassIsActive + `', el.dataset.id === watchID);
      });
      const status = document.getElementById('` + AppDOMSearchStatus + `');
      if (watchPollTimer) { clearInterval(watchPollTimer); watchPollTimer = null; }
      loadWatchResults(watchID, status);
      watchPollTimer = setInterval(async () => {
        await loadWatches(watchID);
        await loadWatchResults(watchID, status);
      }, ` + strconv.Itoa(ListingSearchWatchPollIntervalMs) + `);
    }
    async function loadWatches(selectID) {
      const box = document.getElementById('` + AppDOMSearchWatches + `');
      if (!box) return;
      const res = await fetch(pathMeWatches, { credentials: 'same-origin' });
      if (!res.ok) { box.innerHTML = '<p class="hint">` + AppCopySearchFailed + `</p>'; return; }
      const data = await res.json();
      const watches = data.` + JSONFieldWatches + ` || [];
      if (!watches.length) {
        box.innerHTML = '<p class="hint">` + AppCopySearchWatchesEmpty + `</p>';
        return;
      }
      box.innerHTML = '<div class="panel-head" style="padding:0 0 8px;border:0"><h2>` + AppCopySearchWatchesTitle + `</h2></div>' +
        watches.map(w => {
          const id = w.` + JSONFieldID + `;
          const active = id === (selectID || activeWatchID) ? ' ` + AppClassIsActive + `' : '';
          const errClass = w.` + JSONFieldLastError + ` ? ' ` + AppClassWatchError + `' : '';
          const metaClass = w.` + JSONFieldMetaStatus + ` === '` + WatchMetaStatusPending + `' ? ' ` + AppClassMetaPending + `' :
            (w.` + JSONFieldMetaStatus + ` === '` + WatchMetaStatusFailed + `' ? ' ` + AppClassMetaFail + `' : ' ` + AppClassMetaReady + `');
          const pauseLabel = w.` + JSONFieldStatus + ` === '` + ListingWatchStatusPaused + `' ? '` + AppCopySearchWatchResume + `' : '` + AppCopySearchWatchPause + `';
          const pauseStatus = w.` + JSONFieldStatus + ` === '` + ListingWatchStatusPaused + `' ? '` + ListingWatchStatusActive + `' : '` + ListingWatchStatusPaused + `';
          return '<div class="watch' + active + errClass + metaClass + '" data-id="' + esc(id) + '" onclick="selectWatch(\'' + esc(id) + '\')">' +
            '<p class="watch-title">' + esc(w.` + JSONFieldLabel + ` || w.` + JSONFieldRegion + ` || id) + '</p>' +
            '<p class="watch-meta">' + formatWatchMeta(w) + '</p>' +
            '<p class="` + AppClassWatchActions + `" onclick="event.stopPropagation()">' +
              '<button type="button" class="btn ghost" onclick="patchWatch(\'' + esc(id) + '\',\'' + pauseStatus + '\')">' + pauseLabel + '</button> ' +
              '<button type="button" class="btn ghost" onclick="resetWatch(\'' + esc(id) + '\')">` + AppCopySearchWatchReset + `</button> ' +
              '<button type="button" class="btn ghost" onclick="stopWatch(\'' + esc(id) + '\')">` + AppCopySearchWatchStop + `</button>' +
            '</p></div>';
        }).join('');
      if (selectID) selectWatch(selectID);
      else if (!activeWatchID && watches[0]) selectWatch(watches[0].` + JSONFieldID + `);
    }
    async function loadWatchResults(watchID, status) {
      const box = document.getElementById('` + AppDOMSearchResults + `');
      const res = await fetch(pathWatchResults(watchID), { credentials: 'same-origin' });
      if (!res.ok) return;
      const data = await res.json();
      const rows = (data.` + JSONFieldResults + ` || []).map(formatResultRow).join('');
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
        await loadWatches(task.` + JSONFieldID + `);
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
      const rows = (data.` + JSONFieldResults + ` || []).map(formatResultRow).join('');
      renderResults(box, rows);
    }
    if (document.getElementById('` + AppDOMSearchWatches + `')) loadWatches();
  </script>
`
}
