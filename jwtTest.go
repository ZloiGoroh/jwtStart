package main

import (
	//"errors"

	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type post struct {
	Theme string
	Text  string
}

type personInfo struct {
	Name  string
	Posts []post
}

type customClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

/*var data = personInfo{
	Name: "David",
	Posts: []post{
		{
			Theme: "Watch list",
			Text:  "Some movies, serials, youtube clips",
		},
		{
			Theme: "My day",
			Text:  "Wrote some code, watched a movie, enjoyed every moment",
		},
	},
}*/
var templates = template.Must(template.ParseFiles("startPage.html", "localPage.html"))

func main() {
	http.HandleFunc("/", startPageHandler)
	http.HandleFunc("/page/", profileHandler)
	http.HandleFunc("/login/", setJWTHandler)
	http.HandleFunc("/make-post/", makePostHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	postsArray := []post{}
	cookieThemes, err1 := r.Cookie("Theme")
	cookieTexts, _ := r.Cookie("Text")
	fmt.Println(cookieThemes)
	if err1 == nil {
		themesArray := strings.Split(cookieThemes.Value, "&")
		textsArray := strings.Split(cookieTexts.Value, "&")
		j := 0
		for range themesArray {
			postsArray[j] = post{Theme: themesArray[j], Text: textsArray[j]}
			j++
		}
	}

	//var p = []string{"one theme", "two theme", "three theme"}
	//newP := strings.Join(p[:], ",")
	// cok := http.Cookie{
	// 	Name:  "Themes",
	// 	Value: newP,
	// }
	//oldP := strings.Split(newP, ",")
	//http.SetCookie(w, &cok)
	c, _ := r.Cookie("JWTTest")
	user := decodeToken(c)
	userData := personInfo{
		Name:  user.Username,
		Posts: postsArray,
	}
	err := templates.ExecuteTemplate(w, "localPage.html", userData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func startPageHandler(w http.ResponseWriter, r *http.Request) {
	if c, _ := r.Cookie("JWTTest"); c != nil {
		claim := decodeToken(c)
		if claim != nil {
			http.Redirect(w, r, "/page/", http.StatusFound)
		}
	}
	err := templates.ExecuteTemplate(w, "startPage.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setJWTHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	makeToken(w, name, "/page", "JWTTest")
	makeToken(w, name, "/", "JWTTest")
	http.Redirect(w, r, "/page/", http.StatusFound)
}

func makePostHandler(w http.ResponseWriter, r *http.Request) {
	sendingThemes := ""
	sendingTexts := ""
	allThemesCookie, err := r.Cookie("Theme")
	allTextsCookie, _ := r.Cookie("Text")
	if err == nil {
		themes := strings.Split(allThemesCookie.Value, "&")
		newThemes := make([]string, len(themes))
		texts := strings.Split(allTextsCookie.Value, "&")
		newTexts := make([]string, len(texts))
		newThemes = themes
		newTexts = texts
		newThemes[len(themes)] = r.FormValue("theme")
		sendingThemes = strings.Join(newThemes, "&")
		newTexts[len(texts)] = r.FormValue("text")
		sendingTexts = strings.Join(newTexts, "&")
	} else {
		sendingThemes = r.FormValue("theme")
		sendingTexts = r.FormValue("text")
	}

	newThemeCookie := http.Cookie{
		Name:  "Theme",
		Value: sendingThemes,
		Path:  "/page",
	}
	http.SetCookie(w, &newThemeCookie)
	newThemeCookie2 := http.Cookie{
		Name:  "Theme",
		Value: sendingThemes,
		Path:  "/make-post",
	}
	http.SetCookie(w, &newThemeCookie2)

	newTextCookie := http.Cookie{
		Name:  "Text",
		Value: sendingTexts,
		Path:  "/page",
	}
	http.SetCookie(w, &newTextCookie)
	newTextCookie2 := http.Cookie{
		Name:  "Text",
		Value: sendingTexts,
		Path:  "/make-post",
	}
	http.SetCookie(w, &newTextCookie2)
	http.Redirect(w, r, "/page", http.StatusFound)
}

func decodeToken(c *http.Cookie) *customClaims {
	gotToken, _ := jwt.ParseWithClaims(
		c.Value,
		&customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("someTextToSecure"), nil
		},
	)
	newClaims, err := gotToken.Claims.(*customClaims)
	if !err {
		return nil
	}
	return newClaims
}

func makeToken(w http.ResponseWriter, name string, path string, cookieName string) {
	claims := customClaims{
		Username: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "jwtTest",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte("someTextToSecure"))
	cookie := http.Cookie{
		Name:  cookieName,
		Value: signedToken,
		Path:  path,
	}
	http.SetCookie(w, &cookie)
}
