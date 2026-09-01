package tokens

// Admin UI DOM ids — app face HTML + admin script.
const (
	AppDOMProxiesTable       = "proxies-table"
	AppDOMProxiesForm        = "proxies-form"
	AppDOMProxiesEditID      = "proxy-edit-id"
	AppDOMProxiesEndpoint    = "proxy-endpoint"
	AppDOMProxiesStatus      = "proxy-status"
	AppDOMProxiesLabel       = "proxy-label"
	AppDOMProxiesFormStatus  = "proxies-form-status"
	AppDOMAvitoTable         = "avito-table"
	AppDOMAvitoForm          = "avito-form"
	AppDOMAvitoEditID        = "avito-edit-id"
	AppDOMAvitoLabel         = "avito-label"
	AppDOMAvitoStatus        = "avito-status"
	AppDOMAvitoExternalRef   = "avito-external-ref"
	AppDOMAvitoPassword      = "avito-password"
	AppDOMAvitoFormStatus    = "avito-form-status"
)

// Admin UI copy (RU).
const (
	AppCopyColStatus              = "Статус"
	AppCopyColActions             = "Действия"
	AppCopyColEndpoint            = "Endpoint"
	AppCopyColLabel               = "Метка"
	AppCopyColExternalRef         = "Логин"
	AppCopyAdminCreate            = "Создать"
	AppCopyAdminSave              = "Сохранить"
	AppCopyAdminCancel            = "Отмена"
	AppCopyAdminDelete            = "Удалить"
	AppCopyAdminEdit              = "Изменить"
	AppCopyAdminPasswordRequired  = "Пароль"
	AppCopyAdminPasswordOptional  = "Пароль (необяз. при редактировании)"
	AppCopyAdminConfirmDelete     = "Удалить запись?"
	AppCopyAdminLoadFailed        = "Ошибка загрузки"
	AppCopyAdminSaveFailed        = "Ошибка сохранения"
	AppCopyAdminDeleteFailed      = "Ошибка удаления"
	AppCopyAdminEmpty             = "Нет записей"
)

// ProxyStatusValues / AvitoAccountStatusValues mirror PostgreSQL enums (admin UI + API).
var (
	ProxyStatusValues = []string{"ACTIVE", "BANNED", "DISABLED"}
	AvitoAccountStatusValues = []string{"ACTIVE", "DISABLED", "ERROR"}
)
