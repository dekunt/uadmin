package uadmin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/dekunt/uadmin/colors"
	"github.com/dekunt/uadmin/helper"
)

// HideInDashboarder used to check if a model should be hidden in
// dashboard
type HideInDashboarder interface {
	HideInDashboard() bool
}

// CustomTranslation !
var CustomTranslation = []string{
	"uadmin/system",
}

// Register is used to register models to uadmin
func Register(langCode string, m ...interface{}) {
	modelList := []interface{}{}

	if models == nil {
		models = map[string]interface{}{}

		// Initialize system models
		modelList = []interface{}{
			DashboardMenu{},
			User{},
			UserGroup{},
			Session{},
			UserPermission{},
			GroupPermission{},
			Language{},
			Log{},
		}
	}

	// System models count
	SMCount := len(modelList)

	// Now add user defined models
	modelList = append(modelList,
		m...,
	)

	// Inialize the Database
	initializeDB(modelList...)

	// Setup languages
	initializeLanguage(langCode)

	// Store models in Model global variable
	// and initialize the dashboard
	dashboardMenus := []DashboardMenu{}
	All(&dashboardMenus)
	modelExists := false
	cat := ""
	Schema = map[string]ModelSchema{}
	for i := range modelList {
		t := reflect.TypeOf(modelList[i])
		name := strings.ToLower(t.Name())
		models[name] = modelList[i]

		// Register Dashboard menu
		// First check if the model is already in dashboard
		for _, val := range dashboardMenus {
			if name == val.URL {
				modelExists = true
				break
			}
		}
		// If not in dashboard, then add it
		if !modelExists {
			hideItem := false
			if _, ok := t.MethodByName("HideInDashboard"); ok {
				hider := modelList[i].(HideInDashboarder)
				hideItem = hider.HideInDashboard()
			}

			// Check if the model is a system model
			if i < SMCount {
				cat = "System"
			} else {
				cat = ""
			}
			// TODO: Make the name a plural properly
			dashboard := DashboardMenu{
				MenuName: strings.Join(helper.SplitCamelCase(t.Name()), " ") + "s",
				URL:      name,
				Hidden:   hideItem,
				Cat:      cat,
			}
			Save(&dashboard)
		}
		modelExists = false
	}

	// Check if encrypt key is there or generate it
	if _, err := os.Stat(".key"); os.IsNotExist(err) {
		EncryptKey = generateByteArray(32)
		ioutil.WriteFile(".key", EncryptKey, 0600)
	} else {
		EncryptKey, _ = ioutil.ReadFile(".key")
	}

	// Check if salt is there or generate it
	users := []User{}
	if _, err := os.Stat(".salt"); os.IsNotExist(err) {
		Salt = GenerateBase64(72)
		ioutil.WriteFile(".salt", []byte(Salt), 0600)
		if Count(&users, "") != 0 {
			recoveryPass := GenerateBase64(24)
			recoverUsername := GenerateBase64(8)
			for Count(&users, "username = ?", recoverUsername) != 0 {
				recoverUsername = GenerateBase64(8)
			}
			admin := User{
				FirstName:    "System",
				LastName:     "Recovery Admin",
				Username:     recoverUsername,
				Password:     hashPass(recoveryPass),
				Admin:        true,
				RemoteAccess: false,
				Active:       true,
			}
			admin.Save()
			Trail(WARNING, "Your salt file was missing, and a new one was generated NO USERS CAN LOGIN UNTIL PASSWORDS ARE RESET.")
			Trail(INFO, "uAdmin generated a recovery user for you. Username:%s Password:%s", admin.Username, recoveryPass)
		}
	} else {
		saltBytes, _ := ioutil.ReadFile(".salt")
		Salt = string(saltBytes)
	}

	// Create an admin user if there is no user in the system
	if Count(&users, "") == 0 {
		admin := User{
			FirstName:    "System",
			LastName:     "Admin",
			Username:     "admin",
			Password:     hashPass("admin"),
			Admin:        true,
			RemoteAccess: true,
			Active:       true,
		}
		admin.Save()
		Trail(INFO, "Auto generated admin user. Username:admin, Password:admin.")
	}

	// Register admin inlines
	RegisterInlines(UserGroup{}, map[string]string{
		"GroupPermission": "UserGroupID",
	})

	RegisterInlines(User{}, map[string]string{
		"UserPermission": "UserID",
	})

	// Get Global Schema
	stat := map[string]int{}
	for _, v := range CustomTranslation {
		tempStat := syncCustomTranslation(v, langCode)
		for k, v := range tempStat {
			stat[k] += v
		}
	}
	for k, v := range models {
		//t := reflect.TypeOf(v)
		//Schema[t.Name()], _ = getSchema(v)
		Schema[k], _ = getSchema(v)
		tempStat := syncModelTranslation(Schema[k], langCode)
		for k, v := range tempStat {
			stat[k] += v
		}
	}
	for k, v := range stat {
		complete := float64(v) / float64(stat[langCode])
		if complete != 1 {
			Trail(WARNING, "Translation of %s at %.0f%% [%d/%d]", k, complete*100, v, stat[langCode])
		}

	}

	// Mark registered as true to prevent auto registeration
	registered = true
}

// RegisterInlines is a function to register a model as an inline for another model
// Parameters:
// ===========
//   model (struct instance): Is the model that you want to add inlines to.
//   fk (map[interface{}]string): This is a map of the inlines to be added to the model.
//                                The map's key is the name of the model of the inline
//                                and the value of the map is the foreign key field's name.
//  Example:
//  ========
//  type Person struct {
//    uadmin.Model
//    Name string
//  }
//
//  type Card struct {
//    uadmin.Model
//    PersonID uint
//    Person   Person
//  }
//
// func main() {
//   ...
//   uadmin.RegisterInlines(Person{}, map[string]string{
//     "Card": "PersonID",
//   })
//   ...
// }
func RegisterInlines(model interface{}, fk map[string]string) {
	// TODO: sanity check for the parameters
	// Get the name of the model
	modelName := strings.ToLower(reflect.TypeOf(model).Name())
	if inlines == nil {
		inlines = map[string][]interface{}{}
	}
	if foreignKeys == nil {
		foreignKeys = map[string]map[string]string{}
	}
	inlineList := []interface{}{}
	fkMap := map[string]string{}
	for k, v := range fk {
		kmodel, _ := NewModel(strings.ToLower(k), false)
		t := reflect.TypeOf(kmodel.Interface())
		fkMap[strings.ToLower(t.Name())] = gorm.ToColumnName(v)
		// Check if the field name is in the struct
		if t.Kind() != reflect.Struct {
			fmt.Printf("%sUnable to register inline for (%s) inline %s.%s. Please pass a struct as key.\n", colors.Error, reflect.TypeOf(model).Name(), t.Name(), v)
			continue
		}
		if _, ok := t.FieldByName(v); !ok {
			fmt.Printf("%sUnable to register inline for (%s) inline %s.%s. Field name is not in struct.\n", colors.Error, reflect.TypeOf(model).Name(), t.Name(), v)
			continue
		}
		inlineList = append(inlineList, kmodel.Interface())
	}
	inlines[modelName] = inlineList
	inlines[reflect.TypeOf(model).Name()] = inlineList
	foreignKeys[modelName] = fkMap
	delete(Schema, modelName)
	Schema[modelName], _ = getSchema(model)
}

func registerHandlers() {
	// register static and add parameter
	if !strings.HasSuffix(RootURL, "/") {
		RootURL = RootURL + "/"
	}
	if !strings.HasPrefix(RootURL, "/") {
		RootURL = "/" + RootURL
	}

	// Handleer for uAdmin, static and media
	http.HandleFunc(RootURL, mainHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/media/", mediaHandler)

	// api handler
	http.HandleFunc(RootURL+"api/", apiHandler)
	http.HandleFunc(RootURL+"revertHandler/", revertLogHandler)

	handlersRegistered = true
}
