package tokens

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

// Admin face paths / fixture identity (mock shell until server-rendered UI is contracted).
const (
	AdminFaceTitleSuffix = "Admin"
	AdminMockClockLabel  = "MOCK clock"
)

// Admin UI copy (RU) — SoT for generated face.html.
const (
	AdminCopyAuthLede       = "Админка платформы. Вход только по Magic Link — без паролей."
	AdminCopyEmailLabel     = "Email"
	AdminCopyEmailPlaceholder = "you@company.ru"
	AdminCopySendLink       = "Прислать ссылку"
	AdminCopyOpenMock       = "Открыть макет админки →"
	AdminCopyAuthHint       = "Макет лица. Реальный Magic Link HTTP — по контракту позже."
	AdminCopyOverviewTitle  = "Обзор"
	AdminCopyOverviewLede   = "Платформа сервисов. Shipped и reserved — из каталога токенов."
	AdminCopyServicesTitle  = "Сервисы"
	AdminCopyServicesLede   = "Каталог product_services (SoT = tokens + БД)."
	AdminCopyAvitoTitle     = "Аккаунты Авито"
	AdminCopyAvitoLede      = "Пул аккаунтов. Данные появятся после admin API."
	AdminCopyProxiesTitle   = "Прокси"
	AdminCopyProxiesLede    = "Управление прокси. Данные появятся после admin API."
	AdminCopyUsersTitle     = "Пользователи"
	AdminCopyUsersLede      = "USER / ADMIN. Magic Link only. Данные появятся после admin API."
	AdminCopyEmptyPanel     = "Пусто — ждём контрактный API."
	AdminCopyLogout         = "Выйти"
	AdminCopyStatAvito      = "Аккаунты Авито"
	AdminCopyStatProxies    = "Прокси (schema)"
	AdminCopyStatShipped    = "Сервисы shipped"
	AdminCopyColCode        = "Code"
	AdminCopyColTitle       = "Title"
	AdminCopyColShipped     = "Shipped"
)

// AdminNavItem is a shell navigation entry.
type AdminNavItem struct {
	ID    string
	Label string
}

// AdminNav is the admin shell navigation SoT.
var AdminNav = []AdminNavItem{
	{ID: "overview", Label: "Обзор"},
	{ID: "services", Label: "Сервисы"},
	{ID: "avito", Label: "Аккаунты Авито"},
	{ID: "proxies", Label: "Прокси"},
	{ID: "users", Label: "Пользователи"},
}

// FixtureAdminEmail is the mock side-meta identity (ModulePath-derived).
func FixtureAdminEmail() string {
	return "admin@" + ModulePath + ".io"
}

// ProductBrandStem / ProductBrandAccent split ProductNameLocal for display accent.
func ProductBrandStem() string {
	r := []rune(ProductNameLocal)
	if len(r) < 2 {
		return ProductNameLocal
	}
	return string(r[:len(r)-2])
}

func ProductBrandAccent() string {
	r := []rune(ProductNameLocal)
	if len(r) < 2 {
		return ""
	}
	return string(r[len(r)-2:])
}

// RenderAdminFaceHTML returns the canonical web/admin/face.html body (tokens only).
func RenderAdminFaceHTML() string {
	var b strings.Builder
	esc := html.EscapeString
	stem := esc(ProductBrandStem())
	accent := esc(ProductBrandAccent())
	title := esc(ProductNameLocal + " — " + AdminFaceTitleSuffix)
	emailPh := esc(AdminCopyEmailPlaceholder)
	fixtureEmail := esc(FixtureAdminEmail())

	b.WriteString("<!DOCTYPE html>\n")
	b.WriteString("<html lang=\"ru\">\n<head>\n")
	b.WriteString("  <meta charset=\"utf-8\" />\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	fmt.Fprintf(&b, "  <title>%s</title>\n", title)
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" />\n", esc(FontsGooglePreconnect))
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" crossorigin />\n", esc(FontsGoogleStatic))
	fmt.Fprintf(&b, "  <link href=\"%s\" rel=\"stylesheet\" />\n", esc(FontsGoogleCSSURL))
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\" />\n", esc(PathAdminFaceTokensHref))
	b.WriteString(adminFaceStyleBlock())
	b.WriteString("</head>\n<body>\n")

	// Auth
	b.WriteString("  <section id=\"screen-auth\" class=\"screen is-active\">\n")
	b.WriteString("    <div class=\"auth\">\n      <div class=\"auth-card\">\n")
	fmt.Fprintf(&b, "        <p class=\"brand\">%s<span>%s</span></p>\n", stem, accent)
	fmt.Fprintf(&b, "        <p class=\"lede\">%s</p>\n", esc(AdminCopyAuthLede))
	b.WriteString("        <form id=\"magic-form\" onsubmit=\"return enterAdmin(event)\">\n")
	b.WriteString("          <div class=\"field\">\n")
	fmt.Fprintf(&b, "            <label for=\"email\">%s</label>\n", esc(AdminCopyEmailLabel))
	fmt.Fprintf(&b, "            <input id=\"email\" type=\"email\" required placeholder=\"%s\" autocomplete=\"email\" />\n", emailPh)
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <button class=\"btn\" type=\"submit\">%s</button>\n", esc(AdminCopySendLink))
	fmt.Fprintf(&b, "          <button class=\"btn btn-ghost\" type=\"button\" onclick=\"enterAdmin(event, true)\">%s</button>\n", esc(AdminCopyOpenMock))
	b.WriteString("        </form>\n")
	fmt.Fprintf(&b, "        <p class=\"hint\">%s</p>\n", esc(AdminCopyAuthHint))
	b.WriteString("      </div>\n    </div>\n  </section>\n\n")

	// Admin shell
	b.WriteString("  <section id=\"screen-admin\" class=\"screen\">\n    <div class=\"shell\">\n")
	b.WriteString("      <aside class=\"side\">\n")
	fmt.Fprintf(&b, "        <div class=\"side-brand\">%s<em>%s</em></div>\n", stem, accent)
	b.WriteString("        <nav class=\"nav\" id=\"nav\">\n")
	for i, item := range AdminNav {
		active := ""
		if i == 0 {
			active = " class=\"is-active\""
		}
		fmt.Fprintf(&b, "          <a href=\"#%s\"%s data-view=\"%s\">%s</a>\n",
			esc(item.ID), active, esc(item.ID), esc(item.Label))
	}
	b.WriteString("        </nav>\n        <div class=\"side-meta\">\n")
	fmt.Fprintf(&b, "          <span class=\"pulse\"><i></i> %s</span>\n", esc(AdminMockClockLabel))
	fmt.Fprintf(&b, "          <span>%s</span>\n", fixtureEmail)
	fmt.Fprintf(&b, "          <a href=\"#\" onclick=\"return backAuth()\">%s</a>\n", esc(AdminCopyLogout))
	b.WriteString("        </div>\n      </aside>\n\n      <main class=\"main\">\n")

	// Overview
	shippedN := strconv.Itoa(len(ShippedServiceCodes()))
	b.WriteString("        <div class=\"view is-active\" id=\"view-overview\">\n")
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(&b, "            <h1>%s</h1>\n", esc(AdminCopyOverviewTitle))
	fmt.Fprintf(&b, "            <p>%s</p>\n", esc(AdminCopyOverviewLede))
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <div class=\"live-bar\">%s <samp id=\"tick\">--:--:--</samp></div>\n", esc(AdminMockClockLabel))
	b.WriteString("          </div>\n          <div class=\"grid\">\n")
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b>—</b></div>\n", esc(AdminCopyStatAvito))
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b>—</b></div>\n", esc(AdminCopyStatProxies))
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b>%s</b></div>\n", esc(AdminCopyStatShipped), esc(shippedN))
	b.WriteString("          </div>\n          <div class=\"panel\">\n")
	b.WriteString("            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AdminCopyServicesTitle))
	b.WriteString("</h2><span class=\"chip\">")
	b.WriteString(esc(AuthMethodMagicLink))
	b.WriteString("</span></div>\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Services view
	b.WriteString("        <div class=\"view\" id=\"view-services\">\n")
	writeViewHeader(&b, AdminCopyServicesTitle, AdminCopyServicesLede, false)
	b.WriteString("          <div class=\"panel\">\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Empty sections
	for _, sec := range []struct{ id, title, lede string }{
		{"avito", AdminCopyAvitoTitle, AdminCopyAvitoLede},
		{"proxies", AdminCopyProxiesTitle, AdminCopyProxiesLede},
		{"users", AdminCopyUsersTitle, AdminCopyUsersLede},
	} {
		fmt.Fprintf(&b, "        <div class=\"view\" id=\"view-%s\">\n", esc(sec.id))
		writeViewHeader(&b, sec.title, sec.lede, false)
		b.WriteString("          <div class=\"panel empty\">")
		b.WriteString(esc(AdminCopyEmptyPanel))
		b.WriteString("</div>\n        </div>\n\n")
	}

	b.WriteString("      </main>\n    </div>\n  </section>\n")
	b.WriteString(adminFaceScriptBlock())
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

func writeViewHeader(b *strings.Builder, title, lede string, _ bool) {
	esc := html.EscapeString
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(b, "            <h1>%s</h1>\n", esc(title))
	fmt.Fprintf(b, "            <p>%s</p>\n", esc(lede))
	b.WriteString("          </div></div>\n")
}

func writeServicesTable(b *strings.Builder) {
	esc := html.EscapeString
	b.WriteString("            <table>\n              <thead><tr>")
	fmt.Fprintf(b, "<th>%s</th><th>%s</th><th>%s</th>", esc(AdminCopyColCode), esc(AdminCopyColTitle), esc(AdminCopyColShipped))
	b.WriteString("</tr></thead>\n              <tbody>\n")
	for _, s := range ProductServiceCatalog {
		shippedClass := "status-off"
		shippedVal := "false"
		if s.Shipped {
			shippedClass = "status-ok"
			shippedVal = "true"
		}
		fmt.Fprintf(b, "                <tr><td class=\"mono\">%s</td><td>%s</td><td><span class=\"status %s\">%s</span></td></tr>\n",
			esc(s.Code), esc(s.Title), shippedClass, shippedVal)
	}
	b.WriteString("              </tbody>\n            </table>\n")
}

func adminFaceStyleBlock() string {
	// Layout chrome only; colors/fonts/spacing via CSS variables from tokens.css.
	return `  <style>
    * { box-sizing: border-box; }
    html, body {
      margin: 0; min-height: 100%;
      background:
        radial-gradient(1200px 600px at 10% -10%, color-mix(in srgb, var(--color-accent) 18%, transparent), transparent 60%),
        radial-gradient(900px 500px at 100% 0%, color-mix(in srgb, var(--color-success) 10%, transparent), transparent 55%),
        linear-gradient(180deg, var(--color-canvas-hi) 0%, var(--color-canvas) 45%, var(--color-canvas-lo) 100%);
      color: var(--color-text); font-family: var(--font-sans);
    }
    button, input { font: inherit; }
    a { color: inherit; text-decoration: none; }
    .screen { display: none; min-height: 100vh; }
    .screen.is-active { display: block; }
    .auth { min-height: 100vh; display: grid; place-items: center; padding: var(--space-lg); position: relative; overflow: hidden; }
    .auth::before {
      content: ""; position: absolute; inset: auto -20% -30% -20%; height: 55%;
      background:
        linear-gradient(90deg, transparent, color-mix(in srgb, var(--color-accent) 12%, transparent), transparent),
        repeating-linear-gradient(90deg, color-mix(in srgb, var(--color-border) 55%, transparent) 0 1px, transparent 1px 48px);
      transform: perspective(500px) rotateX(58deg); pointer-events: none; opacity: .55;
    }
    .auth-card { width: min(440px, 100%); position: relative; z-index: 1; }
    .brand { font-family: var(--font-display); font-weight: 800; font-size: clamp(3rem, 8vw, 4.5rem); letter-spacing: -.04em; line-height: .95; margin: 0 0 var(--space-md); }
    .brand span { color: var(--color-accent); }
    .lede { margin: 0 0 var(--space-xl); color: var(--color-text-muted); font-size: 1.05rem; line-height: 1.5; max-width: 34ch; }
    .field { display: grid; gap: var(--space-sm); margin-bottom: var(--space-md); }
    .field label { font-size: .8rem; letter-spacing: .08em; text-transform: uppercase; color: var(--color-text-muted); }
    .field input {
      width: 100%; padding: 14px 16px; border-radius: var(--radius-md);
      border: 1px solid var(--color-border);       background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas));
      color: var(--color-text); outline: none;
    }
    .field input:focus { border-color: color-mix(in srgb, var(--color-accent) 70%, var(--color-border)); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 22%, transparent); }
    .btn {
      display: inline-flex; align-items: center; justify-content: center; gap: var(--space-sm);
      width: 100%; padding: 14px 18px; border: 0; border-radius: var(--radius-md);
      background: var(--color-accent); color: var(--color-on-accent); font-weight: 700; cursor: pointer;
    }
    .btn:hover { filter: brightness(1.06); }
    .btn-ghost { background: transparent; color: var(--color-text-muted); border: 1px solid var(--color-border); margin-top: var(--space-sm); }
    .hint { margin-top: var(--space-lg); font-size: .9rem; color: var(--color-text-muted); }
    .shell { display: grid; grid-template-columns: 260px 1fr; min-height: 100vh; }
    @media (max-width: 900px) { .shell { grid-template-columns: 1fr; } .side { position: sticky; top: 0; z-index: 5; } }
    .side { padding: var(--space-lg); border-right: 1px solid var(--color-border); background: color-mix(in srgb, var(--color-surface) 70%, transparent); backdrop-filter: blur(10px); }
    .side-brand { font-family: var(--font-display); font-weight: 800; font-size: 1.6rem; letter-spacing: -.03em; margin-bottom: var(--space-xl); }
    .side-brand em { font-style: normal; color: var(--color-accent); }
    .nav { display: grid; gap: 4px; }
    .nav a { padding: 10px 12px; border-radius: var(--radius-sm); color: var(--color-text-muted); font-weight: 500; }
    .nav a.is-active, .nav a:hover { background: color-mix(in srgb, var(--color-accent) 12%, transparent); color: var(--color-text); }
    .side-meta { margin-top: var(--space-xl); padding-top: var(--space-lg); border-top: 1px solid var(--color-border); font-size: .85rem; color: var(--color-text-muted); display: grid; gap: 6px; }
    .pulse { display: inline-flex; align-items: center; gap: 8px; color: var(--color-success); font-family: var(--font-mono); font-size: .78rem; }
    .pulse i { width: 8px; height: 8px; border-radius: 50%; background: var(--color-success); animation: ping 1.6s infinite; }
    @keyframes ping { 0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 55%, transparent); } 70% { box-shadow: 0 0 0 10px transparent; } 100% { box-shadow: 0 0 0 0 transparent; } }
    .main { padding: var(--space-xl) var(--space-lg); }
    .top { display: flex; flex-wrap: wrap; align-items: end; justify-content: space-between; gap: var(--space-md); margin-bottom: var(--space-xl); }
    .top h1 { margin: 0; font-family: var(--font-display); font-size: clamp(1.8rem, 3vw, 2.4rem); letter-spacing: -.03em; }
    .top p { margin: 6px 0 0; color: var(--color-text-muted); max-width: 48ch; }
    .chip { font-family: var(--font-mono); font-size: .75rem; padding: 6px 10px; border-radius: var(--radius-md); border: 1px solid var(--color-border); color: var(--color-text-muted); background: color-mix(in srgb, var(--color-surface) 80%, transparent); }
    .grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-md); margin-bottom: var(--space-xl); }
    @media (max-width: 900px) { .grid { grid-template-columns: 1fr; } }
    .stat { padding: var(--space-lg); border: 1px solid var(--color-border); border-radius: var(--radius-md); background: linear-gradient(160deg, color-mix(in srgb, var(--color-surface) 92%, var(--color-text)), var(--color-surface)); }
    .stat b { display: block; font-family: var(--font-display); font-size: 2rem; letter-spacing: -.03em; margin-top: var(--space-sm); }
    .stat span { color: var(--color-text-muted); font-size: .9rem; }
    .panel { border: 1px solid var(--color-border); border-radius: var(--radius-md); background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas)); overflow: hidden; }
    .panel.empty { padding: var(--space-xl); color: var(--color-text-muted); }
    .panel-head { display: flex; justify-content: space-between; align-items: center; gap: var(--space-md); padding: var(--space-md) var(--space-lg); border-bottom: 1px solid var(--color-border); }
    .panel-head h2 { margin: 0; font-size: 1rem; font-weight: 600; }
    table { width: 100%; border-collapse: collapse; }
    th, td { text-align: left; padding: 14px 18px; border-bottom: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent); font-size: .92rem; }
    th { color: var(--color-text-muted); font-weight: 500; font-size: .75rem; letter-spacing: .06em; text-transform: uppercase; }
    tr:last-child td { border-bottom: 0; }
    .status { font-family: var(--font-mono); font-size: .78rem; padding: 4px 8px; border-radius: var(--radius-sm); }
    .status-ok { color: var(--color-success); background: color-mix(in srgb, var(--color-success) 14%, transparent); }
    .status-off { color: var(--color-text-muted); background: color-mix(in srgb, var(--color-border) 35%, transparent); }
    .mono { font-family: var(--font-mono); font-size: .85rem; color: var(--color-text-muted); }
    .view { display: none; }
    .view.is-active { display: block; animation: in .35s ease; }
    @keyframes in { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
    .live-bar { display: flex; align-items: center; gap: var(--space-sm); color: var(--color-text-muted); font-size: .85rem; }
    .live-bar samp { font-family: var(--font-mono); color: var(--color-accent); }
  </style>
`
}

func adminFaceScriptBlock() string {
	return `  <script>
    const auth = document.getElementById('screen-auth');
    const admin = document.getElementById('screen-admin');
    const tick = document.getElementById('tick');
    function enterAdmin(e, skip) {
      if (e) e.preventDefault();
      if (!skip) {
        const email = document.getElementById('email').value.trim();
        if (!email.includes('@')) return false;
      }
      auth.classList.remove('is-active');
      admin.classList.add('is-active');
      return false;
    }
    function backAuth() {
      admin.classList.remove('is-active');
      auth.classList.add('is-active');
      return false;
    }
    document.getElementById('nav').addEventListener('click', (e) => {
      const a = e.target.closest('a[data-view]');
      if (!a) return;
      e.preventDefault();
      document.querySelectorAll('#nav a').forEach(x => x.classList.toggle('is-active', x === a));
      document.querySelectorAll('.view').forEach(v => v.classList.toggle('is-active', v.id === 'view-' + a.dataset.view));
    });
    setInterval(() => {
      tick.textContent = new Date().toLocaleTimeString('ru-RU', { hour12: false });
    }, 1000);
  </script>
`
}
