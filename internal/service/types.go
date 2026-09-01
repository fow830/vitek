package service

import "vitek/internal/repository"

// Repository model aliases — transport must use these, not internal/repository.
type (
	Proxy              = repository.Proxy
	AvitoAccount       = repository.AvitoAccount
	PlanType           = repository.PlanType
	ProxyStatus        = repository.ProxyStatus
	AvitoAccountStatus = repository.AvitoAccountStatus
	UserRole           = repository.UserRole
)

const (
	PlanTypeFree  = repository.PlanTypeFREE
	PlanTypePro   = repository.PlanTypePRO
	PlanTypeUltra = repository.PlanTypeULTRA

	ProxyStatusActive   = repository.ProxyStatusACTIVE
	ProxyStatusDisabled = repository.ProxyStatusDISABLED
	ProxyStatusBanned   = repository.ProxyStatusBANNED

	AvitoStatusActive   = repository.AvitoAccountStatusACTIVE
	AvitoStatusDisabled = repository.AvitoAccountStatusDISABLED
	AvitoStatusError    = repository.AvitoAccountStatusERROR

	UserRoleAdmin = repository.UserRoleADMIN
)

func (u SessionUser) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

func ValidPlanType(p PlanType) bool {
	switch p {
	case PlanTypeFree, PlanTypePro, PlanTypeUltra:
		return true
	default:
		return false
	}
}

func ValidProxyStatus(s ProxyStatus) bool {
	switch s {
	case ProxyStatusActive, ProxyStatusDisabled, ProxyStatusBanned:
		return true
	default:
		return false
	}
}

func ValidAvitoStatus(s AvitoAccountStatus) bool {
	switch s {
	case AvitoStatusActive, AvitoStatusDisabled, AvitoStatusError:
		return true
	default:
		return false
	}
}
