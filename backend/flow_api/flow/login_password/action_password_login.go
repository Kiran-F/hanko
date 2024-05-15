package login_password

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordLogin struct {
	shared.Action
}

func (a PasswordLogin) GetName() flowpilot.ActionName {
	return constants.ActionPasswordLogin
}

func (a PasswordLogin) GetDescription() string {
	return "Login with a password."
}

func (a PasswordLogin) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.PasswordInput("password").Required(true))

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}
}

func (a PasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	var userID uuid.UUID

	if c.Stash().Get("email").Exists() {
		emailModel, err := deps.Persister.GetEmailPersister().FindByAddress(c.Stash().Get("email").String())
		if err != nil {
			return fmt.Errorf("failed to find user by email: %w", err)
		}

		if emailModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = *emailModel.UserID
	} else if c.Stash().Get("username").Exists() {
		username := c.Stash().Get("username").String()
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return fmt.Errorf("failed to find user via username: %w", err)
		}

		if userModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = userModel.ID
	} else {
		return a.wrongCredentialsError(c)
	}

	// TODO
	//if h.rateLimiter != nil {
	//	err := rate_limiter.Limit(h.rateLimiter, userId, c)
	//	if err != nil {
	//		return err
	//	}
	//}

	err := deps.PasswordService.VerifyPassword(userID, c.Input().Get("password").String())
	if err != nil {
		if errors.Is(err, services.ErrorPasswordInvalid) {
			err = deps.AuditLogger.CreateWithConnection(
				deps.Tx,
				deps.HttpContext,
				models.AuditLogLoginFailure,
				&models.User{ID: userID},
				err,
				auditlog.Detail("login_method", "password"),
				auditlog.Detail("flow_id", c.GetFlowID()))
			if err != nil {
				return fmt.Errorf("could not create audit log: %w", err)
			}

			return a.wrongCredentialsError(c)
		}

		return fmt.Errorf("failed to verify password: %w", err)
	}

	// Set only for audit logging purposes.
	err = c.Stash().Set("login_method", "password")
	if err != nil {
		return fmt.Errorf("failed to set login_method to the stash: %w", err)
	}

	return c.EndSubFlow()
}

func (a PasswordLogin) wrongCredentialsError(c flowpilot.ExecutionContext) error {
	c.Input().SetError("password", flowpilot.ErrorValueInvalid)
	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(errors.New("wrong credentials")))
}

func (a PasswordLogin) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
