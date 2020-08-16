package webui

import (
	"fmt"
	"net/http"

	"github.com/go-ldap/ldap/v3"
)

var (
	_ http.Handler = changePassForm{}
)

type changePassForm struct {
	*Server
}

var (
	changePassFormHTML =`
				<p>Change password:</p>
				<input name="username" type="text" placeholder="username" pattern="[a-z0-9]*" value="{username}" required autofocus><br>
				<input name="password" type="password" placeholder="old password" required><br>
				<input name="newpassword" type="password" placeholder="new password" oninput="form.confirmpassword.pattern = escapeRegExp(this.value)" required><br>
				<input name="confirmpassword" type="password" placeholder="confirm password" required><br>
				<input type="submit" value="SUBMIT">
`
)

func (form changePassForm) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	toRender := replaceValues(changePassFormHTML, req)

	switch req.Method {
	case http.MethodPost:
		toRender += "<p style='width:200px'>" + form.processPost(req) + "</p>"
	}

	form.Logger.Tracef("toRender: %v", toRender)

	_, _ = res.Write(renderForm(toRender))
}


func (form changePassForm) processPost(req *http.Request) string {
	form.Logger.Tracef("processPost")
	defer form.Logger.Tracef("/processPost")

	conn, err := form.ldapConnector.AcquireConn()
	if err != nil {
		return "error: unable to acquire connection to LDAP: "+ err.Error()
	}
	defer conn.Release()

	username := req.PostFormValue("username")
	if username == "" {
		return "error: username is empty"
	}

	// validateUsername is required to prevent injections below
	if err := validateUsername(username); err != nil {
		return "error: username is invalid: "+err.Error()
	}

	password := req.PostFormValue("password")
	if password == "" {
		return "error: password is empty"
	}
	newPassword := req.PostFormValue("newpassword")
	if newPassword == "" {
		return "error: newpassword is empty"
	}
	confirmPassword := req.PostFormValue("confirmpassword")
	if confirmPassword == "" {
		return "error: confirmpassword is empty"
	}

	if newPassword != confirmPassword {
		return "error: confirmpassword does not match newpassword"
	}

	form.Logger.Tracef("all good, starting the LDAP operations")
	err = conn.Request(func(conn *ldap.Conn) error {
		form.Logger.Tracef("conn.Request()")
		defer form.Logger.Tracef("/conn.Request()")

		// To prevent injections it is required to check username with
		// "validateUsername".
		userID := "uid="+username+",ou=People,dc=dx,dc=center"
		form.Logger.Tracef("userID: %v", userID)

		err := conn.Bind(userID, password)
		if err != nil {
			return fmt.Errorf("unable to bind: %w", err)
		}

		request := &ldap.PasswordModifyRequest{
			UserIdentity: userID,
			OldPassword:  password,
			NewPassword:  newPassword,
		}
		form.Logger.Tracef("request: %#+v", request)
		_, err = conn.PasswordModify(request)
		if err != nil {
			return fmt.Errorf("unable to modify password: %w", err)
		}

		return nil
	})

	if err != nil {
		form.Logger.Warnf("result: %v", err)
		return "error: unable to change password: " + err.Error()
	}

	form.Logger.Tracef("success")
	return "success"
}
