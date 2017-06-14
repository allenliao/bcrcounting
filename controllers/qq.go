package controllers

import (
	"bcrcounting/models"
	"fmt"
	"goutils"
	"io/ioutil"
	"net/http"
	"net/url"

	simplejson "github.com/bitly/go-simplejson"
)

//QQ發佈長龍通知
func PublishChanceResultToQQ(_countingSuggest *models.CountingResult, _betAccount *models.SimBetAccount) {
	//Message := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " (" + _countingSuggest.TrendName + ")"
	SuggestionBetStr := models.TransBetTypeToStr(_countingSuggest.SuggestionBet)
	//SuggestionBetAmountMultipleStr := "$" + fmt.Sprint(_countingSuggest.SuggestionBetAmount/_countingSuggest.DefaultBetAmount)
	SuggestionBetAmountStr := "$" + fmt.Sprint(_countingSuggest.SuggestionBetAmount)
	Message := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " 下一局建議買 " + SuggestionBetStr + " " + SuggestionBetAmountStr + " 帳戶餘額:" + fmt.Sprint(_betAccount.Balance)
	sendMsgToQQGroup(Message)
}

func sendMsgToQQGroup(Message string) {
	host := "127.0.0.1"
	port := "1236"
	RobotQQ := "3378333039"
	Key := "a123456b"
	GroupId := "254998530"

	/*
		resp, err := http.PostForm("http://"+host+":"+port+"/SendClusterIM.do",
			url.Values{"RobotQQ": {RobotQQ},
				"Key":     {Key},
				"GroupId": {GroupId},
				"Message": {Message}})
	*/
	encodeMsg := url.QueryEscape(Message)
	urlStr := "http://" + host + ":" + port + "/SendClusterIM.do?RobotQQ=" + RobotQQ + "&Key=" + Key + "&GroupId=" + GroupId + "&Message=" + encodeMsg

	resp, err := http.Get(urlStr)

	goutils.Logger.Info("sendMsgToQQGroup 發送消息:", urlStr)
	if err != nil {
		goutils.Logger.Error("sendMsgToQQGroup Get:"+urlStr+" Error:", err.Error())
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			goutils.Logger.Error("sendMsgToQQGroup ReadAll:"+urlStr+" Error:", err.Error())
		} else {
			jsonStr := body
			jsonObj, err := simplejson.NewJson(jsonStr)
			if err != nil {
				goutils.Logger.Error("sendMsgToQQGroup simplejson.NewJson Error:", err.Error())
			}
			//goutils.CheckErr(err)
			nFlag, _ := jsonObj.Get("nFlag").Int()
			Info, _ := jsonObj.Get("Info").String()
			goutils.Logger.Info("sendMsgToQQGroup 發送消息 Response:", string(jsonStr))

			if nFlag == 1 {
				goutils.Logger.Info("sendMsgToQQGroup 發送消息成功:", Info)
			} else {
				goutils.Logger.Info("sendMsgToQQGroup !發送消息失敗:", Info)
			}

		}
		//goutils.CheckErr(err)
	}
}
