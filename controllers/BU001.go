//取得 BU001 TABLE 的 資料
//儲存結果
//計算結果
//決定告知結果

package controllers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"fmt"

	"bcrcounting/models"

	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
)

var (
	BUCode       string
	tableInfoMap map[string]*tableInfo
	tableAmount  uint8
	tableResult  chan TableInitJsonStr
)

type tableInfo struct {
	TableCode                 string
	TableNo                   uint8
	CurrentCountingResultList map[string]models.CountingResultInterface //紀錄賽局結果
	//CurrentCountingResultMethod1 *models.CountingResultMethod1 //紀錄方法1的決策結果
	//CurrentCountingResultMethod2 *models.CountingResultMethod2 //紀錄方法2的決策結果
}

type TableInitJsonStr struct {
	TableCode string
	JsonStr   []byte // Only for WebSocket users; otherwise nil.
}

func InitBU() {
	initDefaultValue()
	StartProcess()
}

//初始化變數 create Table Info
func initDefaultValue() {
	BUCode = "BU001"
	tableResult = make(chan TableInitJsonStr, 10)
	tableInfoMap = make(map[string]*tableInfo)
	tableCodeList := []string{"0001001", "0001002", "0001003", "0001004", "0001005", "0001006", "0001007", "0001008", "0001009", "0001010", "0001011", "0001012", "0001013", "0001014"}
	tableNoList := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	for idx, tableCode := range tableCodeList {
		tableNo := tableNoList[idx]
		currentCountingResultList := models.CreateCurrentCountingResultList(BUCode, tableNo) //map[string]models.CountingResultInterface
		tableInfoMap[tableCode] = &tableInfo{TableCode: tableCode, TableNo: tableNo, CurrentCountingResultList: currentCountingResultList}
	}

	for _, currentCountingResultInterface := range tableInfoMap["0001002"].CurrentCountingResultList {
		currentCountingResult := currentCountingResultInterface.GetCountingResult()
		beego.Info("initDefaultValue currentCountingResult.BUCode:" + currentCountingResult.BUCode)

	}

}

func StartProcess() {
	go processData()
	for _, tableInfo := range tableInfoMap {
		go fetchTableData(tableInfo) //TODO: HARD CODE
	}
}

func jsonBeadRoadCode2BetTypeStr(beadRoadCode string) string {
	if beadRoadCode == "2" || beadRoadCode == "6" || beadRoadCode == "7" {
		return fmt.Sprint(models.Bcr_BETTYPE_PLAYER)
	}
	if beadRoadCode == "1" || beadRoadCode == "4" || beadRoadCode == "5" || beadRoadCode == "8" {
		return fmt.Sprint(models.Bcr_BETTYPE_BANKER)
	}
	if beadRoadCode == "3" || beadRoadCode == "10" || beadRoadCode == "11" {
		return fmt.Sprint(models.Bcr_BETTYPE_TIE)
	}
	return fmt.Sprint(models.Bcr_BETTYPE_TIE)
}

func barcode2point(barcode int) int {
	point := barcode % 13
	if point == 0 || point >= 10 {
		point = 0
	}
	return point
}
func jsonGameResult2BetType(result, betType string) uint8 {
	if betType == "1" {
		if result == "B" {
			return models.Bcr_BETTYPE_BANKER
		}
		if result == "P" {
			return models.Bcr_BETTYPE_PLAYER
		}
		if result == "T" {
			return models.Bcr_BETTYPE_TIE
		}
	}

	return models.Bcr_BETTYPE_NONE
}

//處理資料
//儲存結果
//計算結果
func processData() {
	for {
		//收到桌面資料後
		if _tableResult, ok := <-tableResult; ok {
			tableCode := _tableResult.TableCode
			jsonObj, err := simplejson.NewJson(_tableResult.JsonStr)
			if err != nil {
				beego.Error("simplejson.NewJson Error:", err.Error())
				continue
			}
			//goutils.CheckErr(err)
			//shoeID, _ := jsonObj.Get("DCGameVO").Get("shoeID").Int()
			gameIDDisplay, _ := jsonObj.Get("DCGameVO").Get("gameIDDisplay").String()
			handCount, _ := jsonObj.Get("DCGameVO").Get("handCount").Int()
			gameStatus, _ := jsonObj.Get("gameStatus").Int()
			arrayOfGameResult, _ := jsonObj.Get("arrayOfGameResult").Array()
			beadRoadDisplayList, _ := jsonObj.Get("allRoadDisplayList").Get("beadRoadDisplayList").Array()

			//所有算法輪巡
			for _, currentCountingResultInterface := range tableInfoMap[tableCode].CurrentCountingResultList {
				currentCountingResult := currentCountingResultInterface.GetCountingResult()
				//beego.Info(string(_tableResult.JsonStr))
				//gameStatus= 1=init 2=bet 3=dealing 4=resulting 5=end
				if handCount == 1 && !currentCountingResult.HasInit {
					//換靴 重算
					currentCountingResultInterface.InitChangShoeField()
				}

				if gameIDDisplay != currentCountingResult.GameIDDisplay && gameStatus == 4 && len(arrayOfGameResult) > 0 && len(beadRoadDisplayList) >= handCount {
					beego.Info("tableCode:" + tableCode + " json.gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus) + " currentCountingResult.SuggestionBet:" + currentCountingResult.SuggestionBet)
					beego.Info("tableCode:" + tableCode + " beadRoadDisplayList.len:" + fmt.Sprint(len(beadRoadDisplayList)) + " handCount:" + fmt.Sprint(handCount) + " TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)))
					currentCountingResult.HasInit = false
					currentCountingResult.GameIDDisplay = gameIDDisplay //標記算過了
					//若上一局有預測結果，要告知這一局的發牌結果
					if currentCountingResult.SuggestionBet != "" {
						for _, resultObj := range arrayOfGameResult {
							resultMap, _ := resultObj.(map[string]interface{}) //要做斷言檢查才能使用
							resultStr := fmt.Sprint(resultMap["result"])
							betTypeStr := fmt.Sprint(resultMap["betType"])

							betType := jsonGameResult2BetType(resultStr, betTypeStr)
							beego.Info("tableCode:" + tableCode + " arrayOfGameResult resultStr:" + resultStr + " betTypeStr:" + betTypeStr)
							if betType != models.Bcr_BETTYPE_NONE {
								//取得結果
								currentCountingResult.Result = models.TransBetTypeToStr(betType)
								currentCountingResult.TieReturn = (currentCountingResult.Result == models.TransBetTypeToStr(models.Bcr_BETTYPE_TIE) &&
									(currentCountingResult.SuggestionBet == models.TransBetTypeToStr(models.Bcr_BETTYPE_BANKER) || currentCountingResult.SuggestionBet == models.TransBetTypeToStr(models.Bcr_BETTYPE_PLAYER)))
								currentCountingResult.GuessResult = currentCountingResult.Result == currentCountingResult.SuggestionBet

								break
							}

						}

						beego.Info("tableCode:" + tableCode + " 公佈預測結果  currentCountingResult.Result:" + currentCountingResult.Result + " currentCountingResult.GuessResult:" + fmt.Sprint(currentCountingResult.GuessResult))
						PublishCountingResult(currentCountingResult) //公佈預測結果(有沒有猜中)

					}

					//取牌
					b1, _ := jsonObj.Get("baccaratResultVO").Get("b1").Int()
					b1 = b1 % 13
					b2, _ := jsonObj.Get("baccaratResultVO").Get("b2").Int()
					b3, err := jsonObj.Get("baccaratResultVO").Get("b3").Int()
					if err != nil {
						b3 = -1
					}
					p1, _ := jsonObj.Get("baccaratResultVO").Get("p1").Int()
					p2, _ := jsonObj.Get("baccaratResultVO").Get("p2").Int()
					p3, err := jsonObj.Get("baccaratResultVO").Get("p3").Int()
					if err != nil {
						p3 = -1
					}

					//beego.Info("JsonStr:", string(_tableResult.JsonStr))

					beego.Info("B1~3,P1~3:", b1, b2, b3, p1, p2, p3)
					//算牌
					cardList := [6]int{b1, b2, b3, p1, p2, p3}
					for idx, barcode := range cardList {
						cardList[idx] = barcode2point(barcode)
					}
					var beadRoadStr string
					//取路紙(珠盤路)
					if beadRoadDisplayList != nil {
						//beadRoadDisplayListLen := len(beadRoadDisplayList)
						//beadRoadStrList := make([]int, beadRoadDisplayListLen)
						var beadRoadBfr bytes.Buffer

						for _, betType := range beadRoadDisplayList {
							beadRoadBfr.WriteString(jsonBeadRoadCode2BetTypeStr(fmt.Sprint(betType)))
							//betType, _ := betType.(map[string]interface{}) //要做斷言檢查才能使用
							//beego.Info("tableCode:" + tableCode + " 珠盤路[" + fmt.Sprint(idx) + "]:" + fmt.Sprint(betType))
						}
						beadRoadStr = beadRoadBfr.String()
						beego.Info("tableCode:" + tableCode + " 珠盤路:" + beadRoadStr)

					}

					//餵牌 餵路紙 做計算
					gotResult := currentCountingResultInterface.Counting(cardList, beadRoadStr)
					if gotResult {
						//有預測結果了
						beego.Info("tableCode:" + tableCode + " 有預測結果了 決定告知預測")
						PublishCountingSuggest(currentCountingResult) //決定告知預測
					} else {
						//這局沒勝算，清除上一期預測結果(已公布過的)
						currentCountingResultInterface.ClearGuessResult()
					}
				}

			}

		}
	}
}

//取得 BU001 TABLE 的 資料  tableCode := "0001005"
func fetchTableData(_tableInfo *tableInfo) {
	tableCode := _tableInfo.TableCode

	//var duration time.Duration = 1 //1 秒取一次
	//for _ = range time.Tick(duration * time.Second) {
	ticker := time.NewTicker(time.Millisecond * 200)
	for _ = range ticker.C {
		/*
			timestamp := time.Now().Local()
			str := "fetchTableData TableCode:" + tableCode + " => at " + timestamp.String()
			fmt.Println(str)
		*/

		connectTable(tableCode)
	}
}

//取得 BU001 TABLE 的 資料
func connectTable(tableCode string) {

	millisecond := fmt.Sprint((time.Now().UnixNano()))
	//beego.Info("connectTable TableCode:" + tableCode + " time.Millisecond:" + millisecond)
	resp, err := http.Get("http://spi.mld.v9vnb.org/GetData.ashx?tablecode=" + tableCode + "&valuetype=INIT&t=" + millisecond)
	if err != nil {
		beego.Error("connectTable Get:"+tableCode+" Error:", err.Error())
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			beego.Error("connectTable ReadAll:"+tableCode+" Error:", err.Error())
		} else {
			tableResult <- TableInitJsonStr{TableCode: tableCode, JsonStr: body} //傳資料出去
		}
		//goutils.CheckErr(err)
	}

	//beego.Info("body:" + string(body))

}
