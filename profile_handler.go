package uadmin

import (
	"fmt"
	"html/template"
	"net/http"
)

// profileHandler !
func profileHandler(w http.ResponseWriter, r *http.Request, session *Session) {
	/*
		http://domain.com/admin/profile/
	*/
	r.ParseMultipartForm(32 << 20)
	type Context struct {
		User         string
		ID           uint
		Schema       ModelSchema
		Status       bool
		IsUpdated    bool
		Notif        string
		Demo         bool
		SiteName     string
		Language     Language
		RootURL      string
		ProfilePhoto string
		OTPImage     string
		OTPRequired  bool
	}

	c := Context{}
	c.RootURL = RootURL
	c.Language = getLanguage(r)
	c.SiteName = SiteName
	user := session.User
	c.User = user.Username
	c.ProfilePhoto = session.User.Photo
	c.OTPImage = "/media/otp/" + session.User.OTPSeed + ".png"

	// Check if OTP Required has been changed
	Trail(DEBUG, "r.FormValue('otp_required'): %s", r.URL.Query().Get("otp_required"))
	if r.URL.Query().Get("otp_required") != "" {
		if r.URL.Query().Get("otp_required") == "1" {
			user.OTPRequired = true
		} else if r.URL.Query().Get("otp_required") == "0" {
			user.OTPRequired = false
		}
		user.Save()
	}

	c.OTPRequired = user.OTPRequired

	s, _ := getSchema(user)
	r.Form.Set("ModelID", fmt.Sprint(user.ID))
	c.Schema = getFormData(user, r, session, s)

	t := template.New("").Funcs(template.FuncMap{
		"Tf": Tf,
	})
	t, err := t.ParseFiles("./templates/uadmin/" + Theme + "/profile.html")

	if err != nil {
		fmt.Fprint(w, err.Error())
		fmt.Println("ERROR", err.Error())
		return
	}

	if r.Method == cPOST {
		c.IsUpdated = true
		if r.FormValue("save") == "" {
			user.Username = r.FormValue("Username")
			user.FirstName = r.FormValue("FirstName")
			user.LastName = r.FormValue("LastName")
			user.Email = r.FormValue("Email")
			f := c.Schema.FieldByName("Photo")
			user.Photo, c.Schema = processUpload(r, f, "user", session, c.Schema)
			user.Save()
		}
		if r.FormValue("save") == "password" {
			oldPassword := r.FormValue("oldPassword")
			newPassword := r.FormValue("newPassword")
			confirmPassword := r.FormValue("confirmPassword")
			_session := user.Login(oldPassword, "")

			if _session == nil || !user.Active {
				c.Status = true
				c.Notif = "Incorrent old password."
			} else if newPassword != confirmPassword {
				c.Status = true
				c.Notif = "New password and confirm password do not match."
			} else {
				user.Password = hashPass(newPassword)
				user.Save()

				// To logout
				Logout(r)

				return
			}
		}
	}

	err = t.ExecuteTemplate(w, "profile.html", c)
	if err != nil {
		fmt.Println(err.Error())
	}

	return
}
