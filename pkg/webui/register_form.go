package webui

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-ldap/ldap/v3"
	"github.com/jsimonetti/pwscheme/ssha"
)

var (
	_ http.Handler = registerForm{}
)

type registerForm struct {
	*Server
}

var (
	registerFormHTML =`
				<p>Add user:</p>
				<input name="username" type="text" placeholder="username" pattern="[a-z0-9]*" value="{username}" required autofocus><br>
				<input name="newpassword" type="password" placeholder="new password" oninput="form.confirmpassword.pattern = escapeRegExp(this.value)" required><br>
				<input name="confirmpassword" type="password" placeholder="confirm password" required><br>
				<input type="submit" value="SUBMIT">
`
)

func (form registerForm) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	toRender := replaceValues(registerFormHTML, req)

	switch req.Method {
	case http.MethodPost:
		toRender += "<p style='width:200px'>" + form.processPost(req) + "</p>"
	}

	_, _ = res.Write(renderForm(toRender))
}

func (form registerForm) processPost(req *http.Request) string {
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

	newPasswordHash, err := ssha.Generate(newPassword, 4)
	if err != nil {
		return "error: unable to hash the password: "+err.Error()
	}

	err = conn.Request(func(conn *ldap.Conn) error {
		form.Logger.Tracef("bindDN: %s", form.BindDN)
		err := conn.Bind(form.BindDN, form.BindPassword)
		if err != nil {
			log.Print(err)
			return fmt.Errorf("unable to bind")
		}

		request := &ldap.AddRequest{
			// To prevent injections it is required to check username with
			// "validateUsername".
			DN:         "uid="+username+",ou=People,dc=dx,dc=center",
			Attributes: []ldap.Attribute{
				{
					Type: "objectClass",
					Vals: []string{"inetOrgPerson", "posixAccount", "shadowAccount"},
				},
				{
					Type: "uid",
					Vals: []string{username},
				},
				{
					Type: "sn",
					Vals: []string{"x"},
				},
				{
					Type: "givenName",
					Vals: []string{"x"},
				},
				{
					Type: "cn",
					Vals: []string{username},
				},
				{
					Type: "displayName",
					Vals: []string{username},
				},
				{
					Type: "uidNumber",
					Vals: []string{"29999"},
				},
				{
					Type: "gidNumber",
					Vals: []string{"5001"},
				},
				{
					Type: "userPassword",
					Vals: []string{newPasswordHash},
				},
				{
					Type: "gecos",
					Vals: []string{username},
				},
				{
					Type: "loginShell",
					Vals: []string{"/bin/bash"},
				},
				{
					Type: "homeDirectory",
					Vals: []string{"/home/"+username},
				},
			},
		}
		form.Logger.Tracef("request: %#+v", request)
		err = conn.Add(request)

		if err != nil {
			log.Print(err)
			return fmt.Errorf("unable to add '%s'", username)
		}

		return nil
	})

	if err != nil {
		return "error: unable to change password: " + err.Error()
	}

	return "success"
}
