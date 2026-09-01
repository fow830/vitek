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
)

func (u SessionUser) IsAdmin() bool {
	return u.Role == repository.UserRoleADMIN
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
	case repository.ProxyStatusACTIVE, repository.ProxyStatusDISABLED, repository.ProxyStatusBANNED:
		return true
	default:
		return false
	}
}

func ValidAvitoStatus(s AvitoAccountStatus) bool {
	switch s {
	case repository.AvitoAccountStatusACTIVE, repository.AvitoAccountStatusDISABLED, repository.AvitoAccountStatusERROR:
		return true
	default:
		return false
	}
}
