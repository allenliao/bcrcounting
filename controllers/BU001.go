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

//取得 BU001 TABLE 的 資料
//儲存結果
//計算結果
//決定告知結果

package controllers

import (
	"goutils"
	"io/ioutil"
	"net/http"
	"time"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
)

type TableInitJsonStr struct {
	TableCode string
	JsonStr   []byte // Only for WebSocket users; otherwise nil.
}

type CountingResult struct {
	GameIDDisplay       string
	BU                  string
	TableNo             uint8
	BetSuggestionData   *[]BetSuggestion
	SuggestionBet       string
	SuggestionBetAmount int16
	Result              string
	GuessResult         bool
}

type BetSuggestion struct {
	BetType     string
	Probability float32
	SuggestBet  bool
}

var (
	BUCode                string         = "BU001"
	tableResultChan                      = make(chan TableInitJsonStr, 10) //TODO 用ARR 存
	currentCountingResult CountingResult                                   //TODO 用ARR 存

)

func init() {
	/*
		// Initialize language type list.

	*/
	StartProcess()

}

func StartProcess() {
	go processData()
	go fetchTableData()
}

//處理資料
//儲存結果
//計算結果
func processData() {
	for {
		if _tableResult, ok := <-tableResultChan; ok {

			jsonObj, err := simplejson.NewJson(_tableResult.JsonStr)
			goutils.CheckErr(err)
			shoeID, _ := jsonObj.Get("DCGameVO").Get("shoeID").Int()
			gameIDDisplay, _ := jsonObj.Get("DCGameVO").Get("gameIDDisplay").String()
			tableCodeSimple, _ := jsonObj.Get("DCGameVO").Get("tableCodeSimple").String()

			millisecond := fmt.Sprint((time.Now().UnixNano()))
			beego.Info("processData time.Millisecond:" + millisecond)
			beego.Info("shoeID:" + fmt.Sprint(shoeID) + " gameIDDisplay:" + gameIDDisplay)

			if currentCountingResult.GameIDDisplay != "" {
				//第一局不計算
				currentCountingResult = CountingResult{
					GameIDDisplay:       gameIDDisplay,
					BU:                  BUCode,
					TableNo:             tableCodeSimple,
					BetSuggestionData:   &betSuggestionData,
					SuggestionBet:       "莊",
					SuggestionBetAmount: 100,
					Result:              "莊",
					GuessResult:         true}
			}

			//TODO 實作處理結果後回傳
			betSuggestionData := make([]BetSuggestion, 5)

			PublishCountingResult(countingResult) //決定告知結果
		}
	}
}

func fetchTableData() {
	timestamp := time.Now().Local()
	var duration time.Duration = 10 //10 秒一次
	for _ = range time.Tick(duration * time.Second) {

		str := "Polling remote terminal data at <some remote terminal name> at " + timestamp.String()
		fmt.Println(str)

		connectTable("0001002")
	}

}

func connectTable(tableCode string) {

	millisecond := fmt.Sprint((time.Now().UnixNano()))
	beego.Info("connectTable time.Millisecond:" + millisecond)
	resp, err := http.Get("http://spi.mld.v9vnb.org/GetData.ashx?tablecode=" + tableCode + "&valuetype=INIT&t=" + millisecond)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	goutils.CheckErr(err)
	//beego.Info("body:" + string(body))

	tableResultChan <- TableInitJsonStr{TableCode: tableCode, JsonStr: body} //傳資料出去

}
