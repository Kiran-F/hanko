package credential_onboarding

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnVerifyAttestationResponse struct {
	shared.Action
}

func (a WebauthnVerifyAttestationResponse) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnVerifyAttestationResponse
}

func (a WebauthnVerifyAttestationResponse) GetDescription() string {
	return "Send the result which was generated by creating a webauthn credential."
}

func (a WebauthnVerifyAttestationResponse) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("webauthn_available").Bool() {
		c.SuspendAction()
	}

	c.AddInputs(flowpilot.JSONInput("public_key").Required(true).Persist(false))
}

func (a WebauthnVerifyAttestationResponse) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get("webauthn_session_data_id").Exists() {
		return errors.New("webauthn_session_data_id does not exist in the stash")
	}

	sessionDataID, err := uuid.FromString(c.Stash().Get("webauthn_session_data_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse webauthn_session_data_id: %w", err)
	}

	userID, err := uuid.FromString(c.Stash().Get("user_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse user_id into a uuid: %w", err)
	}

	params := services.VerifyAttestationResponseParams{
		Tx:            deps.Tx,
		SessionDataID: sessionDataID,
		PublicKey:     c.Input().Get("public_key").String(),
		UserID:        userID,
		Email:         c.Stash().Get("email").String(),
		Username:      c.Stash().Get("username").String(),
	}

	credential, err := deps.WebauthnService.VerifyAttestationResponse(params)
	if err != nil {
		if errors.Is(err, services.ErrInvalidWebauthnCredential) {
			return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorPasskeyInvalid.Wrap(err))
		}

		return fmt.Errorf("failed to verify attestation response: %w", err)
	}

	err = c.Stash().Set("webauthn_credential", credential)
	if err != nil {
		return fmt.Errorf("failed to set webauthn_credential to the stash: %w", err)
	}

	err = c.Stash().Set("user_has_webauthn_credential", true)
	if err != nil {
		return fmt.Errorf("failed to set user_has_webauthn_credential to the stash: %w", err)
	}

	return c.EndSubFlow()
}
