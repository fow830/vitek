package tokens

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

// Admin face title / live clock label.
const (
	AdminFaceTitleSuffix = "Admin"
	AdminCopyLiveClock   = "LIVE"
)

// AdminNav IDs (shell views).
const (
	AdminNavIDOverview = "overview"
	AdminNavIDServices = "services"
	AdminNavIDAvito    = "avito"
	AdminNavIDProxies  = "proxies"
)

// Admin DOM ids — shared by face HTML, face JS, and Datastar SSE patches.
const (
	AdminDOMScreenAuth  = "screen-auth"
	AdminDOMScreenAdmin = "screen-admin"
	AdminDOMStatAvito   = "stat-avito"
	AdminDOMStatProxy   = "stat-proxy"
	AdminDOMStatShipped = "stat-shipped"
	AdminDOMTick        = "tick"
	AdminDOMAuthHint    = "auth-hint"
	AdminDOMMagicForm   = "magic-form"
	AdminDOMEmailInput  = "email"
	AdminDOMNav         = "nav"
	AdminClassIsActive  = "is-active"
	AdminClassStatusOk  = "status-ok"
	AdminClassStatusOff = "status-off"
	AdminDOMViewPrefix  = "view-"
)

// Admin UI copy (RU) — SoT for generated face.html.
const (
	AdminCopyAuthLede         = "Админка платформы. Вход только по Magic Link — без паролей."
	AdminCopyEmailLabel       = "Email"
	AdminCopyEmailPlaceholder = "you@company.ru"
	AdminCopySendLink         = "Прислать ссылку"
	AdminCopyAuthHint         = "Magic Link: письмо/локальный token → сессия → " + PathAdmin + "."
	AdminCopySentHint         = "Ссылка создана. Локально token вернётся в ответе API."
	AdminCopyRequestFailed    = "Ошибка запроса Magic Link"
	AdminCopyConsumeFailed    = "Ошибка consume"
	AdminCopyOverviewTitle    = "Обзор"
	AdminCopyOverviewLede     = "Платформа сервисов. Shipped и reserved — из каталога токенов."
	AdminCopyServicesTitle    = "Сервисы"
	AdminCopyServicesLede     = "Каталог product_services (SoT = tokens + БД)."
	AdminCopyAvitoTitle       = "Аккаунты Авито"
	AdminCopyAvitoLede        = "Пул аккаунтов. CRUD через " + PathV1AdminAvitoAccounts + "."
	AdminCopyProxiesTitle     = "Прокси"
	AdminCopyProxiesLede      = "Управление прокси. CRUD через " + PathV1AdminProxies + "."
	AdminCopyEmptyPanel       = "Нет строк — создайте через admin API."
	AdminCopyLogout           = "Выйти"
	AdminCopyStatAvito        = "Аккаунты Авито"
	AdminCopyStatProxies      = "Прокси ACTIVE"
	AdminCopyStatShipped      = "Сервисы shipped"
	AdminCopyColCode          = "Code"
	AdminCopyColTitle         = "Title"
	AdminCopyColShipped       = "Shipped"
)

// AdminSSEStatsPatch is the Datastar element patch for overview counters.
func AdminSSEStatsPatch(avitoN, proxyN, shippedN int) string {
	return `<b id="` + AdminDOMStatAvito + `">` + strconv.Itoa(avitoN) + `</b>` +
		`<b id="` + AdminDOMStatProxy + `">` + strconv.Itoa(proxyN) + `</b>` +
		`<b id="` + AdminDOMStatShipped + `">` + strconv.Itoa(shippedN) + `</b>`
}

// AdminNavItem is a shell navigation entry.
type AdminNavItem struct {
	ID    string
	Label string
}

// AdminNav is the admin shell navigation SoT.
var AdminNav = []AdminNavItem{
	{ID: AdminNavIDOverview, Label: AdminCopyOverviewTitle},
	{ID: AdminNavIDServices, Label: AdminCopyServicesTitle},
	{ID: AdminNavIDAvito, Label: AdminCopyAvitoTitle},
	{ID: AdminNavIDProxies, Label: AdminCopyProxiesTitle},
}

// FixtureAdminEmail is the side-meta identity shown before login (ModulePath-derived).
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
	fmt.Fprintf(&b, "<html lang=\"%s\" %s>\n<head>\n", LocaleHTML, AttrDataStar)
	b.WriteString("  <meta charset=\"utf-8\" />\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	fmt.Fprintf(&b, "  <title>%s</title>\n", title)
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" />\n", esc(FontsGooglePreconnect))
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" crossorigin />\n", esc(FontsGoogleStatic))
	fmt.Fprintf(&b, "  <link href=\"%s\" rel=\"stylesheet\" />\n", esc(FontsGoogleCSSURL))
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\" />\n", esc(PathTokensCSS))
	fmt.Fprintf(&b, "  <script type=\"module\" src=\"%s\" data-star></script>\n", esc(DatastarCDNURL))
	b.WriteString(adminFaceStyleBlock())
	b.WriteString("</head>\n<body>\n")

	// Auth
	fmt.Fprintf(&b, "  <section id=\"%s\" class=\"screen %s\">\n", AdminDOMScreenAuth, AdminClassIsActive)
	b.WriteString("    <div class=\"auth\">\n      <div class=\"auth-card\">\n")
	fmt.Fprintf(&b, "        <p class=\"brand\">%s<span>%s</span></p>\n", stem, accent)
	fmt.Fprintf(&b, "        <p class=\"lede\">%s</p>\n", esc(AdminCopyAuthLede))
	fmt.Fprintf(&b, "        <form id=\"%s\" onsubmit=\"return enterAdmin(event)\">\n", AdminDOMMagicForm)
	b.WriteString("          <div class=\"field\">\n")
	fmt.Fprintf(&b, "            <label for=\"%s\">%s</label>\n", AdminDOMEmailInput, esc(AdminCopyEmailLabel))
	fmt.Fprintf(&b, "            <input id=\"%s\" type=\"email\" required placeholder=\"%s\" autocomplete=\"email\" />\n", AdminDOMEmailInput, emailPh)
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <button class=\"btn\" type=\"submit\">%s</button>\n", esc(AdminCopySendLink))
	b.WriteString("        </form>\n")
	fmt.Fprintf(&b, "        <p class=\"hint\" id=\"%s\">%s</p>\n", AdminDOMAuthHint, esc(AdminCopyAuthHint))
	b.WriteString("      </div>\n    </div>\n  </section>\n\n")

	// Admin shell
	fmt.Fprintf(&b, "  <section id=\"%s\" class=\"screen\">\n    <div class=\"shell\">\n", AdminDOMScreenAdmin)
	b.WriteString("      <aside class=\"side\">\n")
	fmt.Fprintf(&b, "        <div class=\"side-brand\">%s<em>%s</em></div>\n", stem, accent)
	fmt.Fprintf(&b, "        <nav class=\"nav\" id=\"%s\">\n", AdminDOMNav)
	for i, item := range AdminNav {
		active := ""
		if i == 0 {
			active = ` class="` + AdminClassIsActive + `"`
		}
		fmt.Fprintf(&b, "          <a href=\"#%s\"%s data-view=\"%s\">%s</a>\n",
			esc(item.ID), active, esc(item.ID), esc(item.Label))
	}
	b.WriteString("        </nav>\n        <div class=\"side-meta\">\n")
	fmt.Fprintf(&b, "          <span class=\"pulse\"><i></i> %s</span>\n", esc(AdminCopyLiveClock))
	fmt.Fprintf(&b, "          <span>%s</span>\n", fixtureEmail)
	fmt.Fprintf(&b, "          <a href=\"#\" onclick=\"return backAuth()\">%s</a>\n", esc(AdminCopyLogout))
	b.WriteString("        </div>\n      </aside>\n\n      <main class=\"main\">\n")

	// Overview
	shippedN := strconv.Itoa(len(ShippedServiceCodes()))
	fmt.Fprintf(&b, "        <div class=\"view %s\" id=\"%s%s\">\n", AdminClassIsActive, AdminDOMViewPrefix, AdminNavIDOverview)
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(&b, "            <h1>%s</h1>\n", esc(AdminCopyOverviewTitle))
	fmt.Fprintf(&b, "            <p>%s</p>\n", esc(AdminCopyOverviewLede))
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <div class=\"live-bar\">%s <samp id=\"%s\">--:--:--</samp></div>\n", esc(AdminCopyLiveClock), AdminDOMTick)
	b.WriteString("          </div>\n          <div class=\"grid\">\n")
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">—</b></div>\n", esc(AdminCopyStatAvito), AdminDOMStatAvito)
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">—</b></div>\n", esc(AdminCopyStatProxies), AdminDOMStatProxy)
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">%s</b></div>\n", esc(AdminCopyStatShipped), AdminDOMStatShipped, esc(shippedN))
	b.WriteString("          </div>\n          <div class=\"panel\">\n")
	b.WriteString("            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AdminCopyServicesTitle))
	b.WriteString("</h2><span class=\"chip\">")
	b.WriteString(esc(AuthMethodMagicLink))
	b.WriteString("</span></div>\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Services view
	fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AdminDOMViewPrefix, AdminNavIDServices)
	writeViewHeader(&b, AdminCopyServicesTitle, AdminCopyServicesLede)
	b.WriteString("          <div class=\"panel\">\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Remaining nav views (empty panels until list UI is contracted).
	for _, item := range AdminNav[2:] {
		fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AdminDOMViewPrefix, esc(item.ID))
		writeViewHeader(&b, item.Label, ledeForNav(item.ID))
		b.WriteString("          <div class=\"panel empty\">")
		b.WriteString(esc(AdminCopyEmptyPanel))
		b.WriteString("</div>\n        </div>\n\n")
	}

	b.WriteString("      </main>\n    </div>\n  </section>\n")
	b.WriteString(adminFaceScriptBlock())
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

// RenderAdminFaceHTMLLoggedIn returns face with admin shell active and SSE boot.
func RenderAdminFaceHTMLLoggedIn(email string) string {
	out := RenderAdminFaceHTML()
	out = strings.Replace(out,
		`id="`+AdminDOMScreenAuth+`" class="screen `+AdminClassIsActive+`"`,
		`id="`+AdminDOMScreenAuth+`" class="screen"`, 1)
	out = strings.Replace(out,
		`id="`+AdminDOMScreenAdmin+`" class="screen"`,
		`id="`+AdminDOMScreenAdmin+`" class="screen `+AdminClassIsActive+`" data-on:load="@get('`+PathAdminSSE+`')"`, 1)
	out = strings.Replace(out, html.EscapeString(FixtureAdminEmail()), html.EscapeString(email), 1)
	return out
}

func writeViewHeader(b *strings.Builder, title, lede string) {
	esc := html.EscapeString
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(b, "            <h1>%s</h1>\n", esc(title))
	fmt.Fprintf(b, "            <p>%s</p>\n", esc(lede))
	b.WriteString("          </div></div>\n")
}

func ledeForNav(id string) string {
	switch id {
	case AdminNavIDAvito:
		return AdminCopyAvitoLede
	case AdminNavIDProxies:
		return AdminCopyProxiesLede
	case AdminNavIDServices:
		return AdminCopyServicesLede
	default:
		return AdminCopyOverviewLede
	}
}

func writeServicesTable(b *strings.Builder) {
	esc := html.EscapeString
	b.WriteString("            <table>\n              <thead><tr>")
	fmt.Fprintf(b, "<th>%s</th><th>%s</th><th>%s</th>", esc(AdminCopyColCode), esc(AdminCopyColTitle), esc(AdminCopyColShipped))
	b.WriteString("</tr></thead>\n              <tbody>\n")
	for _, s := range ProductServiceCatalog {
		shippedClass := AdminClassStatusOff
		shippedVal := BoolStringFalse
		if s.Shipped {
			shippedClass = AdminClassStatusOk
			shippedVal = BoolStringTrue
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
    .screen.` + AdminClassIsActive + ` { display: block; }
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
    .hint { margin-top: var(--space-lg); font-size: .9rem; color: var(--color-text-muted); }
    .shell { display: grid; grid-template-columns: 260px 1fr; min-height: 100vh; }
    @media (max-width: 900px) { .shell { grid-template-columns: 1fr; } .side { position: sticky; top: 0; z-index: 5; } }
    .side { padding: var(--space-lg); border-right: 1px solid var(--color-border); background: color-mix(in srgb, var(--color-surface) 70%, transparent); backdrop-filter: blur(10px); }
    .side-brand { font-family: var(--font-display); font-weight: 800; font-size: 1.6rem; letter-spacing: -.03em; margin-bottom: var(--space-xl); }
    .side-brand em { font-style: normal; color: var(--color-accent); }
    .nav { display: grid; gap: 4px; }
    .nav a { padding: 10px 12px; border-radius: var(--radius-sm); color: var(--color-text-muted); font-weight: 500; }
    .nav a.` + AdminClassIsActive + `, .nav a:hover { background: color-mix(in srgb, var(--color-accent) 12%, transparent); color: var(--color-text); }
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
    .` + AdminClassStatusOk + ` { color: var(--color-success); background: color-mix(in srgb, var(--color-success) 14%, transparent); }
    .` + AdminClassStatusOff + ` { color: var(--color-text-muted); background: color-mix(in srgb, var(--color-border) 35%, transparent); }
    .mono { font-family: var(--font-mono); font-size: .85rem; color: var(--color-text-muted); }
    .view { display: none; }
    .view.` + AdminClassIsActive + ` { display: block; animation: in .35s ease; }
    @keyframes in { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
    .live-bar { display: flex; align-items: center; gap: var(--space-sm); color: var(--color-text-muted); font-size: .85rem; }
    .live-bar samp { font-family: var(--font-mono); color: var(--color-accent); }
  </style>
`
}
