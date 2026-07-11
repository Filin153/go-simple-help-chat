package app

import (
	"fmt"
	"strings"
)

func (a *App) checkMod() error {
	if a == nil {
		return fmt.Errorf("critical app initialization error: app is nil")
	}

	missing := make([]string, 0)
	if a.Repository == nil {
		missing = append(missing, "Repository")
	}
	if a.MsgCache == nil {
		missing = append(missing, "MsgCache")
	}
	if a.JWTService == nil {
		missing = append(missing, "JWTService")
	}
	if a.EncryptionService == nil {
		missing = append(missing, "EncryptionService")
	}
	if a.AuthUseCase == nil {
		missing = append(missing, "AuthUseCase")
	}
	if a.UserUseCase == nil {
		missing = append(missing, "UserUseCase")
	}
	if a.ScheduleUseCase == nil {
		missing = append(missing, "ScheduleUseCase")
	}
	if a.DepartmentUseCase == nil {
		missing = append(missing, "DepartmentUseCase")
	}
	if a.ChatUseCase == nil {
		missing = append(missing, "ChatUseCase")
	}
	if a.API == nil {
		missing = append(missing, "API")
	}
	if a.cfg == nil {
		missing = append(missing, "CONFIG")
	} else {
		if strings.TrimSpace(a.cfg.Admin.Login) == "" {
			missing = append(missing, "Admin.Login")
		}
		if strings.TrimSpace(a.cfg.Admin.Password) == "" {
			missing = append(missing, "Admin.Password")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("critical app initialization error: missing modules: %s", strings.Join(missing, ", "))
	}

	return nil
}
