package tokens

func adminFaceScriptBlock() string {
	return `  <script>
    const tick = document.getElementById('` + AdminDOMTick + `');
    const hint = document.getElementById('` + AdminDOMAuthHint + `');
    const pathRequest = '` + PathV1AuthMagicLink + `';
    const pathConsume = '` + PathV1AuthMagicLinkConsume + `';
    const pathLogout = '` + PathV1AuthLogout + `';
    const pathAdmin = '` + PathAdmin + `';
    async function enterAdmin(e) {
      if (e) e.preventDefault();
      const email = document.getElementById('` + AdminDOMEmailInput + `').value.trim();
      if (!email.includes('@')) return false;
      const res = await fetch(pathRequest, {
        method: 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        body: JSON.stringify({ ` + JSONFieldEmail + `: email })
      });
      if (!res.ok) { hint.textContent = '` + AdminCopyRequestFailed + `'; return false; }
      const data = await res.json();
      if (data.` + JSONFieldToken + `) {
        const creq = await fetch(pathConsume, {
          method: 'POST',
          headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
          body: JSON.stringify({ ` + JSONFieldToken + `: data.` + JSONFieldToken + ` }),
          credentials: 'same-origin'
        });
        if (!creq.ok) { hint.textContent = '` + AdminCopyConsumeFailed + `'; return false; }
        window.location.href = pathAdmin;
        return false;
      }
      hint.textContent = '` + AdminCopySentHint + `';
      return false;
    }
    async function backAuth() {
      try {
        await fetch(pathLogout, { method: 'POST', credentials: 'same-origin' });
      } catch (_) {}
      window.location.href = pathAdmin;
      return false;
    }
    const nav = document.getElementById('` + AdminDOMNav + `');
    if (nav) nav.addEventListener('click', (e) => {
      const a = e.target.closest('a[data-view]');
      if (!a) return;
      e.preventDefault();
      document.querySelectorAll('#` + AdminDOMNav + ` a').forEach(x => x.classList.toggle('` + AdminClassIsActive + `', x === a));
      document.querySelectorAll('.view').forEach(v => v.classList.toggle('` + AdminClassIsActive + `', v.id === '` + AdminDOMViewPrefix + `' + a.dataset.view));
    });
    if (tick) setInterval(() => {
      tick.textContent = new Date().toLocaleTimeString('` + LocaleBCP47 + `', { hour12: false });
    }, 1000);
  </script>
`
}
