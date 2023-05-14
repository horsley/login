package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var appConfig Config

func main() {
	if err := appConfig.Load("config.yaml"); err != nil {
		log.Panicln(err)
	}

	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/auth", authHandler)
	http.ListenAndServe(appConfig.Login.Listen, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { //登录页提交登录
		username := r.FormValue("username")
		password := r.FormValue("password")

		userConfig, ok := appConfig.userMap[username]
		if !ok || password != userConfig.Password {
			log.Println("user", username, "not found, or wrong password")
			fmt.Fprint(w, "用户名或密码错误，请重新输入！")
			return
		}

		// 根据用户权限判断是否可以登录系统
		targetURI := "/"
		if o, err := r.Cookie("origin"); err == nil && o.Value != "" && o.Value != "/login" {
			targetURI = o.Value
		}
		targetURL := r.Header.Get("X-Scheme") + "://" + r.Host + targetURI

		target := appConfig.GetSystemName(targetURL)
		if !ok || !userConfig.CanAccess(target) {
			log.Println("user", username, "password verified, access:", targetURL, "unauthorized")
			fmt.Fprint(w, "您没有访问目标系统的权限！")
			return
		}

		tokenString, err := createToken(username)
		if err != nil {
			log.Println("createToken err:", err)
			http.Error(w, "服务器错误", http.StatusInternalServerError)
			return
		}
		setTokenCookie(w, tokenString)

		http.Redirect(w, r, targetURL, http.StatusFound)
		return
	}

	//第一次无鉴权访问被nginx 302到这里，输出登录页
	if r.Header.Get("X-Original-URI") != "" && r.Header.Get("X-Original-URI") != "/favicon.ico" { //记录原url
		http.SetCookie(w, &http.Cookie{Name: "origin", Value: r.Header.Get("X-Original-URI")})
	}

	tmpl := template.Must(template.ParseFiles("login.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	authSucceed := false
	defer func() {
		if !authSucceed {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}()

	tokenString, err := r.Cookie(cookieName) //第一次无鉴权访问 401
	if err != nil {
		log.Println("no cookie, unauthorized", r.Header.Get("X-Real-IP"))
		return
	}

	claims, err := parseToken(tokenString.Value)
	if err != nil || !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		log.Println("token expired or invalid, unauthorized", r.Header.Get("X-Real-IP"))
		return
	}

	userConfig, ok := appConfig.userMap[claims.Subject]
	if !ok {
		log.Println("user config not found for:", claims.Subject, ", unauthorized")
		return
	}

	targetURL := r.Header.Get("X-Scheme") + "://" + r.Host + r.Header.Get("X-Original-URI")
	target := appConfig.GetSystemName(targetURL)

	if target == "" || !userConfig.CanAccess(target) {
		log.Println("user:", claims.Subject, "access:", targetURL, "target not found/access denied")
		return
	}

	authSucceed = true
	log.Println("user:", claims.Subject, "access:", targetURL, "allow")
}
