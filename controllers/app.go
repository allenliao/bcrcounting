// Copyright 2013 Beego Samples authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package controllers

import (
	"goutils"
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
)

var langTypes []string // Languages that are supported.

func init() {
	goutils.InitLogs()
	// Initialize language type list.
	langTypes = strings.Split(beego.AppConfig.String("lang_types"), "|")

	// Load locale files according to language types.
	for _, lang := range langTypes {
		beego.Trace("Loading language: " + lang)
		if err := i18n.SetMessage(lang, "conf/"+"locale_"+lang+".ini"); err != nil {
			beego.Error("Fail to set message file:", err)
			return
		}
	}

}

// baseController represents base router for all other app routers.
// It implemented some methods for the same implementation;
// thus, it will be embedded into other routers.
type baseController struct {
	beego.Controller // Embed struct that has stub implementation of the interface.
	i18n.Locale      // For i18n usage when process data and render template.
}

// Prepare implemented Prepare() method for baseController.
// It's used for language option check and setting.
func (this *baseController) Prepare() {
	// Reset language option.
	this.Lang = "" // This field is from i18n.Locale.
	isNeedRedir := false
	hasCookie := false

	// 1. Check URL arguments.
	this.Lang = this.Input().Get("lang")

	// 2. Get language information from cookies.
	if len(this.Lang) == 0 {
		this.Lang = this.Ctx.GetCookie("lang")
		hasCookie = true
	} else {
		isNeedRedir = true
	}

	// Check again in case someone modify by purpose.
	if !i18n.IsExist(this.Lang) {
		this.Lang = ""
		isNeedRedir = false
		hasCookie = false
	}

	// 1. Get language information from 'Accept-Language'.
	al := this.Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters.
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	// 2. Default language is English.
	if len(this.Lang) == 0 {
		this.Lang = "en-US"
		isNeedRedir = false
	}

	// Save language information in cookies.
	if !hasCookie {
		this.Ctx.SetCookie("lang", this.Lang, 1<<31-1, "/")
	}

	// Set template level language option.
	this.Data["Lang"] = this.Lang

	if isNeedRedir {
		this.Redirect("/", 302)
	}
}

// AppController handles the welcome screen that allows user to pick a technology and username.
type AppController struct {
	baseController // Embed to use methods that are implemented in baseController.
}

// Get implemented Get() method for AppController.
func (this *AppController) Get() {
	this.TplName = "welcome.html"
}

// Join method handles POST requests for AppController.
func (this *AppController) Join() {
	// Get form value.
	uname := this.GetString("uname")
	tech := this.GetString("tech")

	// Check valid.
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	switch tech {
	case "longpolling":
		this.Redirect("/lp?uname="+uname, 302)
	case "websocket":
		this.Redirect("/ws?uname="+uname, 302)
	default:
		this.Redirect("/", 302)
	}

	// Usually put return after redirect.
	return
}
