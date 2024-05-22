import { Fragment } from "preact";
import { useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../contexts/AppProvider";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import ErrorBox from "../components/error/ErrorBox";
import Footer from "../components/wrapper/Footer";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { useFlowState } from "../contexts/FlowState";

interface Props {
  state: State<"login_method_chooser">;
}

const LoginMethodChooserPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { setLoadingAction, stateHandler } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);

  const onPasskeySelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passkey-submit");
    const nextState = await flowState.actions
      .webauthn_generate_request_options(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onPasswordSelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("password-submit");
    const nextState = await flowState.actions
      .continue_to_password_login(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onPasscodeSelectSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("passcode-submit");
    const nextState = await flowState.actions
      .continue_to_passcode_confirmation(null)
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  const onBackClick = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("back");
    const nextState = await flowState.actions.back(null).run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
  };

  return (
    <Fragment>
      <Content>
        <Headline1>{"Choose a method"}</Headline1>
        <ErrorBox flowError={flowState?.error} />
        <Form
          hidden={!flowState.actions.webauthn_generate_request_options?.(null)}
          onSubmit={onPasskeySelectSubmit}
        >
          <Button secondary={true} uiAction={"passkey-submit"} icon={"passkey"}>
            {"Passkey"}
          </Button>
        </Form>
        <Form
          hidden={!flowState.actions.continue_to_passcode_confirmation?.(null)}
          onSubmit={onPasscodeSelectSubmit}
        >
          <Button
            secondary={true}
            uiAction={"passcode-submit"}
            icon={"passkey"}
          >
            {"Passcode"}
          </Button>
        </Form>
        <Form
          hidden={!flowState.actions.continue_to_password_login?.(null)}
          onSubmit={onPasswordSelectSubmit}
        >
          <Button
            secondary={true}
            uiAction={"password-submit"}
            icon={"passkey"}
          >
            {"Password"}
          </Button>
        </Form>
      </Content>
      <Footer>
        <Link
          uiAction={"back"}
          onClick={onBackClick}
          loadingSpinnerPosition={"right"}
        >
          {t("labels.back")}
        </Link>
        <span hidden />
      </Footer>
    </Fragment>
  );
};

export default LoginMethodChooserPage;
