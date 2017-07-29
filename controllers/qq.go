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

//QQ發佈長龍通知 決定告知預測 發佈建議
func PublishCountingSuggestToQQ(_countingSuggest *models.CountingResult, _betAccount *models.SimBetAccount) {
	//Message := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " (" + _countingSuggest.TrendName + ")"
	SuggestionBetStr := models.TransBetTypeToStr(_countingSuggest.SuggestionBet)
	//SuggestionBetAmountMultipleStr := "$" + fmt.Sprint(_countingSuggest.SuggestionBetAmount/_countingSuggest.DefaultBetAmount)
	SuggestionBetAmountStr := "$" + fmt.Sprint(_countingSuggest.SuggestionBetAmount)

	Message := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " 下一局建议买 " + SuggestionBetStr + " " + SuggestionBetAmountStr
	sendMsgToQQGroup(Message, true)
	sendMsgToQQGroup(Message, false)
}

//發佈建議的結果(公布答案) 公佈預測結果(有沒有猜中)
func PublishGameResultToQQ(_countingResult *models.CountingResult, _betAccount *models.SimBetAccount) {
	var guessResultStr string
	if _countingResult.TieReturn {
		guessResultStr = "平"

	} else {

		if _countingResult.GuessResult {
			guessResultStr = "胜"
		} else {
			guessResultStr = "负"
		}

	}
	if _countingResult.FirstHand {
		guessResultStr = "第一局预测不记结果"
	}

	Message := "第 " + fmt.Sprint(_countingResult.TableNo) + " 桌 " + _countingResult.GameIDDisplay + " 开 " + models.TransBetTypeToStr(_countingResult.Result) + " 建议结果:" + guessResultStr
	sendMsgToQQGroup(Message, true)
	sendMsgToQQGroup(Message, false)
}

//發佈目前 下注 行為
func PublishPlaceBetActionToQQ(betRecord models.BetRecord) {

	Message := "第 " + fmt.Sprint(betRecord.TableNo) + " 桌 " + betRecord.GameIDDisplay + " 下注 " + betRecord.BetTypeStr + " $" + fmt.Sprint(betRecord.BetAmmount) + " 帐户余额:" + fmt.Sprint(betRecord.CurrentBalance)

	sendMsgToQQGroup(Message, false)
}

//發佈目前 派彩 行為
func PublishSettleBetActionToQQ(betRecord models.BetRecord) {

	Message := "第 " + fmt.Sprint(betRecord.TableNo) + " 桌 " + betRecord.GameIDDisplay + " 派彩 " + betRecord.GameResultTypeStr + " $" + fmt.Sprint(betRecord.WinAmmount) + " 帐户余额:" + fmt.Sprint(betRecord.CurrentBalance)

	sendMsgToQQGroup(Message, false)
}

//發佈下注統計
func PublishBetStatisticToQQ(betAccount *models.SimBetAccount) {
	betStatistic := betAccount.SubBetStatistic

	Message := "(小結)由 " + fmt.Sprint(betStatistic.StartTime) + " ~ 目前为止" +
		" \n下注次数:" + fmt.Sprint(betStatistic.BetCount) +
		" \n下注总金额:" + fmt.Sprint(betStatistic.AccumulateBetAmount) +
		" \n输赢总金额:" + fmt.Sprint(betStatistic.TotalWinAmount) +
		" \n胜:" + fmt.Sprint(betStatistic.WinBetCount) +
		" \n负:" + fmt.Sprint(betStatistic.LoseBetCount) +
		" \n平:" + fmt.Sprint(betStatistic.TieBetCount)

	sendMsgToQQGroup(Message, false)
	goutils.Logger.Info("PublishBetStatisticToQQ Message:" + Message)

	betStatistic = betAccount.TotalBetStatistic

	Message = "(總結)由 " + fmt.Sprint(betStatistic.StartTime) + " ~ 目前为止" +
		" \n下注次数:" + fmt.Sprint(betStatistic.BetCount) +
		" \n下注总金额:" + fmt.Sprint(betStatistic.AccumulateBetAmount) +
		" \n输赢总金额:" + fmt.Sprint(betStatistic.TotalWinAmount) +
		" \n胜:" + fmt.Sprint(betStatistic.WinBetCount) +
		" \n负:" + fmt.Sprint(betStatistic.LoseBetCount) +
		" \n平:" + fmt.Sprint(betStatistic.TieBetCount)

	sendMsgToQQGroup(Message, false)
	goutils.Logger.Info("PublishBetStatisticToQQ Message:" + Message)
}

func sendMsgToQQGroup(Message string, goPublic bool) {
	host := "qq.bill.com"
	port := "1236"
	RobotQQ := "3378333039"
	Key := "a123456b"

	GroupId := "254998530" //龍吟88 (公開)
	if !goPublic {
		GroupId = "633301454" //TEST (測試)
	}

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
