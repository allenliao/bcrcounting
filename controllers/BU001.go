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

	"bcrcounting/countingmethod"
	"bcrcounting/models"

	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
)

var (
	BUCode       string
	tableInfoMap map[string]*tableInfo
	tableAmount  uint8
	tableResult  chan TableInitJsonStr
	//currentCountingResult models.CountingResult //TODO 要準備一個 [TableCode]來放 當下這一局的 結果currentCountingResult
)

type tableInfo struct {
	TableCode             string
	TableNo               uint8
	CurrentCountingResult *models.CountingResult //TODO 要準備一個 [TableCode]來放 當下這一局的 結果currentCountingResult
}

type TableInitJsonStr struct {
	TableCode string
	JsonStr   []byte // Only for WebSocket users; otherwise nil.
}

func InitBU() {
	initDefaultValue()
	StartProcess()
}

//初始化變數
func initDefaultValue() {
	BUCode = "BU001"
	tableResult = make(chan TableInitJsonStr, 10)
	tableInfoMap = make(map[string]*tableInfo)
	tableCodeList := []string{"0001001", "0001002", "0001003", "0001004", "0001005", "0001006", "0001007", "0001008", "0001009", "0001010", "0001011", "0001012", "0001013", "0001014"}
	tableNoList := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	for idx, tableCode := range tableCodeList {
		tableNo := tableNoList[idx]
		currentCountingResult := countingmethod.CreateCountingResult(BUCode, tableNo)
		tableInfoMap[tableCode] = &tableInfo{TableCode: tableCode, TableNo: tableNo, CurrentCountingResult: &currentCountingResult}
	}

}

func StartProcess() {
	go processData()
	for _, tableInfo := range tableInfoMap {
		go fetchTableData(tableInfo) //TODO: HARD CODE
	}
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
		if _tableResult, ok := <-tableResult; ok {
			tableCode := _tableResult.TableCode
			jsonObj, err := simplejson.NewJson(_tableResult.JsonStr)
			goutils.CheckErr(err)
			shoeID, _ := jsonObj.Get("DCGameVO").Get("shoeID").Int()
			gameIDDisplay, _ := jsonObj.Get("DCGameVO").Get("gameIDDisplay").String()
			handCount, _ := jsonObj.Get("DCGameVO").Get("handCount").Int()
			gameStatus, _ := jsonObj.Get("gameStatus").Int()
			arrayOfGameResult, _ := jsonObj.Get("arrayOfGameResult").Array()
			beego.Info("shoeID:" + fmt.Sprint(shoeID) + " gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus))

			currentCountingResult := tableInfoMap[tableCode].CurrentCountingResult
			//beego.Info(string(_tableResult.JsonStr))
			//gameStatus= 1=init 2=bet 3=dealing 4=resulting 5=end
			if handCount == 1 && !currentCountingResult.HasInit {
				//換靴 重算
				currentCountingResult.InitCountingData()
			}
			beego.Info("gameIDDisplay:" + gameIDDisplay + " currentCountingResult.GameIDDisplay:" + currentCountingResult.GameIDDisplay)
			if gameIDDisplay != currentCountingResult.GameIDDisplay && gameStatus == 4 {
				currentCountingResult.HasInit = false
				currentCountingResult.GameIDDisplay = gameIDDisplay //算過了
				//若上一局有預測結果，要告知這一局的發牌結果
				if currentCountingResult.SuggestionBet != "" {
					for _, resultObj := range arrayOfGameResult {
						resultMap, _ := resultObj.(map[string]interface{}) //要做斷言檢查才能使用
						resultStr := fmt.Sprint(resultMap["result"])
						betTypeStr := fmt.Sprint(resultMap["betType"])

						betType := jsonGameResult2BetType(resultStr, betTypeStr)
						beego.Info("arrayOfGameResult resultStr:" + resultStr + " betTypeStr:" + betTypeStr)
						if betType != models.Bcr_BETTYPE_NONE {
							//取得結果
							currentCountingResult.Result = models.TransBetTypeToStr(betType)
							currentCountingResult.GuessResult = currentCountingResult.Result == currentCountingResult.SuggestionBet
							break
						}

					}

					//currentCountingResult.Result=
					PublishCountingResult(currentCountingResult) //公佈預測結果(有沒有猜中)

					//清除預測結果
					currentCountingResult.ClearGuessResult()

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

				suggestionResult := countingmethod.Bcr_CountingMethod1(cardList, currentCountingResult)
				if suggestionResult != nil {
					//有預測結果了
					PublishCountingResult(currentCountingResult) //決定告知預測
				}

			}

		}
	}
}

//取得 BU001 TABLE 的 資料  tableCode := "0001005"
func fetchTableData(_tableInfo *tableInfo) {
	tableCode := _tableInfo.TableCode
	timestamp := time.Now().Local()
	var duration time.Duration = 1 //1 秒取一次
	for _ = range time.Tick(duration * time.Second) {

		str := "fetchTableData TableCode:" + tableCode + " => at " + timestamp.String()
		fmt.Println(str)

		connectTable(tableCode)
	}
}

//取得 BU001 TABLE 的 資料
func connectTable(tableCode string) {

	millisecond := fmt.Sprint((time.Now().UnixNano()))
	beego.Info("connectTable TableCode:" + tableCode + " time.Millisecond:" + millisecond)
	resp, err := http.Get("http://spi.mld.v9vnb.org/GetData.ashx?tablecode=" + tableCode + "&valuetype=INIT&t=" + millisecond)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	goutils.CheckErr(err)
	//beego.Info("body:" + string(body))

	tableResult <- TableInitJsonStr{TableCode: tableCode, JsonStr: body} //傳資料出去

}
