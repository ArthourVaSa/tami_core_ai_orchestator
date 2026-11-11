package repositories

import "time"

// IUsageRepository lleva la cuenta de uso por agente para repartir la carga entre cuentas
type IUsageRepository interface {
	RunsToday(agent string) int
	RecordRun(agent string, success bool, costUSD float64) error
	CooldownUntil(agent string) time.Time
	SetCooldown(agent string, until time.Time) error
}
