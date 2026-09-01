package tokens

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
