//取得 BU001 TABLE 的 資料
//儲存結果
//計算結果
//決定告知結果

package controllers

import (
	"bytes"
	"goutils"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"fmt"

	"bcrcounting/models"

	"github.com/bitly/go-simplejson"
)

var (
	BUCode              string
	tableInfoMap        map[string]*tableInfo
	tableAmount         uint8
	tableResult         chan TableInitJsonStr
	Odds                map[uint8]float64                //賠率
	connectTableTimeout = time.Duration(2 * time.Second) //超過2秒沒回應就TIMEOUT
	BetAccount          *models.SimBetAccount
)

type tableInfo struct {
	TableCode                 string
	TableNo                   uint8
	CurrentCountingResultList map[string]models.CountingResultInterface //紀錄賽局結果
	bankerPayout              int8
	bodyStr                   string //從HTTP.GET 取得的 內容
	client                    *http.Client
	//CurrentCountingResultMethod1 *models.CountingResultMethod1 //紀錄方法1的決策結果
	//CurrentCountingResultMethod2 *models.CountingResultMethod2 //紀錄方法2的決策結果
}

type TableInitJsonStr struct {
	TableCode string
	JsonStr   []byte // Only for WebSocket users; otherwise nil.
}

func InitBU() {
	LoginBetAccount()
	InitTableInfo()
	StartProcess()
}

func LoginBetAccount() {
	BetAccount = &models.SimBetAccount{Balance: 10000}
	BetAccount.LoginTime = time.Now()
	BetAccount.BetRecordList = make(map[string]models.BetRecord)
	BetAccount.TotalBetStatistic = &models.BetStatistic{StartTime: time.Now()}
	BetAccount.SubBetStatistic = &models.BetStatistic{StartTime: time.Now()}
	NotifyCurrentBalance()
}

//初始化變數 create Table Info
func InitTableInfo() {
	BUCode = "BU001"
	tableResult = make(chan TableInitJsonStr, 10)
	tableInfoMap = make(map[string]*tableInfo)
	Odds = make(map[uint8]float64) //每一桌都一樣
	Odds[models.Bcr_BETTYPE_BANKER] = 1.95
	Odds[models.Bcr_BETTYPE_PLAYER] = 2
	Odds[models.Bcr_BETTYPE_TIE] = 8
	tableCodeList := []string{"0001001", "0001002", "0001003", "0001004", "0001005", "0001006", "0001007", "0001008", "0001009", "0001010", "0001011", "0001012", "0001013", "0001014"}
	tableNoList := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	for idx, tableCode := range tableCodeList {
		tableNo := tableNoList[idx]
		currentCountingResultList := models.CreateCurrentCountingResultList(BUCode, tableNo) //map[string]models.CountingResultInterface
		tableInfoMap[tableCode] = &tableInfo{TableCode: tableCode, TableNo: tableNo, CurrentCountingResultList: currentCountingResultList}
	}
}

func StartProcess() {
	go runBetStatisticTimeTick() //不起執行緒會佔住執行緒
	go processData()
	for _, tableInfo := range tableInfoMap {
		go fetchTableData(tableInfo) //TODO: HARD CODE
	}

}

func runBetStatisticTimeTick() {
	ticker := time.NewTicker(time.Hour * 4) //每四小時統計一次
	NotifyCurrentBetStatistic()
	for _ = range ticker.C {
		NotifyCurrentBetStatistic()
		ResetBetStatistic()
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

func CalPlaceBetStatistic(betAmount float64) {
	BetAccount.SubBetStatistic.BetCount++
	BetAccount.SubBetStatistic.AccumulateBetAmount += betAmount
	BetAccount.SubBetStatistic.TotalWinAmount -= betAmount

	BetAccount.TotalBetStatistic.BetCount++
	BetAccount.TotalBetStatistic.AccumulateBetAmount += betAmount
	BetAccount.TotalBetStatistic.TotalWinAmount -= betAmount
}

func CalResultStatistic(countingResult *models.CountingResult) {
	if countingResult.Result == models.Bcr_BETTYPE_TIE {
		BetAccount.SubBetStatistic.TieBetCount++
		BetAccount.SubBetStatistic.TotalWinAmount += countingResult.WinAmmount

		BetAccount.TotalBetStatistic.TieBetCount++
		BetAccount.TotalBetStatistic.TotalWinAmount += countingResult.WinAmmount
	} else {
		if countingResult.GuessResult {
			BetAccount.SubBetStatistic.WinBetCount++
			BetAccount.SubBetStatistic.TotalWinAmount += countingResult.WinAmmount

			BetAccount.TotalBetStatistic.WinBetCount++
			BetAccount.TotalBetStatistic.TotalWinAmount += countingResult.WinAmmount
		} else {
			BetAccount.SubBetStatistic.LoseBetCount++

			BetAccount.TotalBetStatistic.LoseBetCount++
		}
	}

}

func ResetBetStatistic() {
	BetAccount.SubBetStatistic.BetCount = 0
	BetAccount.SubBetStatistic.AccumulateBetAmount = 0
	BetAccount.SubBetStatistic.TotalWinAmount = 0
	BetAccount.SubBetStatistic.TieBetCount = 0
	BetAccount.SubBetStatistic.WinBetCount = 0
	BetAccount.SubBetStatistic.LoseBetCount = 0
	BetAccount.SubBetStatistic.StartTime = time.Now()
}

//下注
func PlaceBet(countingResult *models.CountingResult, GameIDDisplay string) {
	//走到這裡都是確定要下注了
	betamount := countingResult.SuggestionBetAmount

	BetAccount.Balance -= betamount

	betRecord := models.BetRecord{BUCode: countingResult.BUCode, TableNo: countingResult.TableNo, Settled: false}
	betRecord.BetAmmount = betamount
	betRecord.BetType = countingResult.SuggestionBet
	betRecord.BetTypeStr = models.TransBetTypeToStr(countingResult.SuggestionBet)
	betRecord.GameIDDisplay = GameIDDisplay
	betRecord.BetTime = time.Now()
	BetAccount.BetRecordList[GameIDDisplay] = betRecord
	betRecord.CurrentBalance = BetAccount.Balance
	goutils.Logger.Info("TableNo:" + fmt.Sprint(countingResult.TableNo) + " PlaceBet Balance:" + fmt.Sprint(BetAccount.Balance) + " BetAmmount:" + fmt.Sprint(betRecord.BetAmmount))
	/*
		BetTime        time.Time
			BUCode         string
			TableNo        uint8
			GameIDDisplay  string //局號
			GameResultType uint8
			BetType        uint8
			BetAmmount     float64
	*/
	CalPlaceBetStatistic(betamount)

	NotifyBetRecord(betRecord)

}

//派彩
func SettleBet(countingResult *models.CountingResult) {
	if countingResult.HasBeted {
		betRecord := BetAccount.BetRecordList[countingResult.GameIDDisplay]
		betRecord.GameResultType = countingResult.Result
		betRecord.GameResultTypeStr = models.TransBetTypeToStr(betRecord.GameResultType)
		//贏錢
		if betRecord.BetType == countingResult.Result || countingResult.Result == models.Bcr_BETTYPE_TIE {
			odd := Odds[countingResult.Result]
			if countingResult.Result == models.Bcr_BETTYPE_TIE && betRecord.BetType != countingResult.Result {
				odd = 1 //下莊閒 開和要退錢
			}
			betRecord.Settled = true
			betRecord.WinAmmount = odd * betRecord.BetAmmount
			countingResult.WinAmmount = betRecord.WinAmmount

			BetAccount.Balance += betRecord.WinAmmount
			betRecord.CurrentBalance = BetAccount.Balance

			goutils.Logger.Info("TableNo:" + fmt.Sprint(countingResult.TableNo) + " SettleBet Balance:" + fmt.Sprint(BetAccount.Balance) + " BetAmmount:" + fmt.Sprint(betRecord.BetAmmount) + " WinAmmount:" + fmt.Sprint(betRecord.WinAmmount) + " odd:" + fmt.Sprint(odd))
			if countingResult.TableNo == 0 {
				panic("TableNo=0")
			}
			NotifyBetRecord(betRecord)

		}

		CalResultStatistic(countingResult)

		countingResult.HasBeted = false
	}

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
				goutils.Logger.Error("simplejson.NewJson Error:", err.Error())
				continue
			}
			//goutils.CheckErr(err)
			//shoeID, _ := jsonObj.Get("DCGameVO").Get("shoeID").Int()
			gameIDDisplay, _ := jsonObj.Get("DCGameVO").Get("gameIDDisplay").String()
			handCount, _ := jsonObj.Get("DCGameVO").Get("handCount").Int()
			gameStatus, _ := jsonObj.Get("gameStatus").Int()
			arrayOfGameResult, _ := jsonObj.Get("arrayOfGameResult").Array() //有時候gameStatus == 5 才有值
			beadRoadDisplayList, _ := jsonObj.Get("allRoadDisplayList").Get("beadRoadDisplayList").Array()

			millisecond := fmt.Sprint((time.Now().UnixNano()))
			goutils.Logger.Info("收到:" + tableCode + " t:" + millisecond)

			//所有算法輪巡
			for _, currentCountingResultInterface := range tableInfoMap[tableCode].CurrentCountingResultList {
				currentCountingResult := currentCountingResultInterface.GetCountingResult()
				//goutils.Logger.Info(string(_tableResult.JsonStr))
				//gameStatus= 1=init 2=bet 3=dealing 4=resulting 5=end
				if handCount == 1 && !currentCountingResult.HasInit {
					//換靴 重算
					currentCountingResultInterface.InitChangShoeField()
				}

				//下注時間 且該桌有預測訊息 還沒被下過注
				if gameStatus == 2 && currentCountingResult.SuggestionBet != models.Bcr_BETTYPE_NONE && !currentCountingResult.HasBeted {
					//決定要不要投注 TODO:移到currentCountingResult中做判斷
					//若已經是第60局 又不是在追倍投 就建議不要下注了
					if !currentCountingResultInterface.IsNeedPlaceBet(handCount, BetAccount.Balance) {
						//if handCount >= 60 && !currentCountingResult.NextBetDubleBet {
						goutils.Logger.Info("tableCode:" + tableCode + " 建議不要下注了 handCount:" + fmt.Sprint(handCount) + " NextBetDubleBet:" + fmt.Sprint(currentCountingResult.NextBetDubleBet))
					} else {
						goutils.Logger.Info("tableCode:" + tableCode + " (PlaceBet) TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)) + " json.gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus) + " currentCountingResult.SuggestionBetStr:" + models.TransBetTypeToStr(currentCountingResult.SuggestionBet))
						goutils.Logger.Info("tableCode:" + tableCode + " (PlaceBet) beadRoadDisplayList.len:" + fmt.Sprint(len(beadRoadDisplayList)) + " handCount:" + fmt.Sprint(handCount))
						PlaceBet(currentCountingResult, gameIDDisplay)
					}

				}
				//FOR TEST
				betRecord := BetAccount.BetRecordList[gameIDDisplay]
				if betRecord.TableNo != 0 {
					goutils.Logger.Info("tableCode:" + tableCode + " (out) TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)) + " json.gameIDDisplay:" + gameIDDisplay + " currentCountingResult.gameIDDisplay:" + currentCountingResult.GameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus) + " currentCountingResult.SuggestionBetStr:" + models.TransBetTypeToStr(currentCountingResult.SuggestionBet))
					goutils.Logger.Info("tableCode:" + tableCode + " (out) beadRoadDisplayList.len:" + fmt.Sprint(len(beadRoadDisplayList)) + " handCount:" + fmt.Sprint(handCount) + " arrayOfGameResult.len:" + fmt.Sprint(len(arrayOfGameResult)))
				}
				if gameIDDisplay != currentCountingResult.GameIDDisplay && gameStatus == 4 {
					goutils.Logger.Info("tableCode:" + tableCode + " TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)) + " GetCard json.gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus) + " currentCountingResult.SuggestionBetStr:" + models.TransBetTypeToStr(currentCountingResult.SuggestionBet))
					goutils.Logger.Info("tableCode:" + tableCode + " beadRoadDisplayList.len:" + fmt.Sprint(len(beadRoadDisplayList)) + " handCount:" + fmt.Sprint(handCount))
					currentCountingResult.HasInit = false
					currentCountingResult.GotCard = true
					currentCountingResult.GotResult = false
					currentCountingResult.GameIDDisplay = gameIDDisplay //標記算過了
					//若上一局有預測結果，要告知這一局的發牌結果 TransBetTypeToStr(roadPatternInfo.SuggestionBetType)

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

					//goutils.Logger.Info("JsonStr:", string(_tableResult.JsonStr))

					goutils.Logger.Info("B1~3,P1~3:", b1, b2, b3, p1, p2, p3)
					//算牌

					cardList := [6]int{b1, b2, b3, p1, p2, p3}
					for idx, barcode := range cardList {
						cardList[idx] = barcode2point(barcode)
					}
					currentCountingResult.CardList = cardList

				}

				//取結果
				//其實就是gameStatus=4/5 才會有機會len(arrayOfGameResult) > 0
				if !currentCountingResult.GotResult && currentCountingResult.GotCard && currentCountingResult.SuggestionBet != models.Bcr_BETTYPE_NONE && len(arrayOfGameResult) > 0 {
					goutils.Logger.Info("tableCode:" + tableCode + " TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)) + " GetResult json.gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus))
					//currentCountingResult.GotCard = false
					//currentCountingResult.GotCard &&
					currentCountingResult.GotResult = true
					for _, resultObj := range arrayOfGameResult {
						resultMap, _ := resultObj.(map[string]interface{}) //要做斷言檢查才能使用
						resultStr := fmt.Sprint(resultMap["result"])
						betTypeStr := fmt.Sprint(resultMap["betType"])

						betType := jsonGameResult2BetType(resultStr, betTypeStr)
						goutils.Logger.Info("tableCode:" + tableCode + " arrayOfGameResult resultStr:" + resultStr + " betTypeStr:" + betTypeStr)
						if betType != models.Bcr_BETTYPE_NONE {
							//取得結果
							currentCountingResult.Result = betType
							currentCountingResult.TieReturn = (currentCountingResult.Result == models.Bcr_BETTYPE_TIE &&
								(currentCountingResult.SuggestionBet == models.Bcr_BETTYPE_BANKER || currentCountingResult.SuggestionBet == models.Bcr_BETTYPE_PLAYER))
							currentCountingResult.FirstHand = (handCount == 1)
							currentCountingResult.GuessResult = currentCountingResult.Result == currentCountingResult.SuggestionBet

							//決定下一注要不要倍投
							currentCountingResultInterface.IsNeedPlaceNextBet() //需呼叫在GuessResult決定之後

							break
						} else {
							currentCountingResult.StopDubleBet()
						}

					}
					if currentCountingResult.FirstHand {
						goutils.Logger.Info("tableCode:" + tableCode + " 公佈預測結果  第一局 預測不算")
					} else {
						goutils.Logger.Info("tableCode:" + tableCode + " 公佈預測結果  currentCountingResult.Result:" + models.TransBetTypeToStr(currentCountingResult.Result) + " currentCountingResult.GuessResult:" + fmt.Sprint(currentCountingResult.GuessResult))
						SettleBet(currentCountingResult)

					}
					NotifyGameResult(currentCountingResult) //公佈預測結果(有沒有猜中)

				}

				//取路紙
				//其實就是gameStatus=4/5 才會有機會len(beadRoadDisplayList) >= handCount
				if currentCountingResult.GotCard {
					gameIDDisplayHand, _ := strconv.ParseInt(GetGameIDDisplayHand(currentCountingResult.GameIDDisplay), 10, 64)
					goutils.Logger.Info("tableCode:" + tableCode + " gameIDDisplayHand:" + fmt.Sprint(gameIDDisplayHand) + "checkGetRoad len(beadRoadDisplayList):" + fmt.Sprint(len(beadRoadDisplayList)))
					if int64(len(beadRoadDisplayList)) == gameIDDisplayHand {
						goutils.Logger.Info("tableCode:" + tableCode + " TypeOf:" + fmt.Sprint(reflect.TypeOf(currentCountingResultInterface)) + " GetRoad json.gameIDDisplay:" + gameIDDisplay + " gameStatus:" + fmt.Sprint(gameStatus))
						currentCountingResult.GotCard = false
						//currentCountingResult.GotResult = false
						//取路紙(珠盤路)
						if beadRoadDisplayList != nil {
							//beadRoadDisplayListLen := len(beadRoadDisplayList)
							//beadRoadStrList := make([]int, beadRoadDisplayListLen)
							var beadRoadBfr bytes.Buffer

							for _, betType := range beadRoadDisplayList {
								beadRoadBfr.WriteString(jsonBeadRoadCode2BetTypeStr(fmt.Sprint(betType)))
								//betType, _ := betType.(map[string]interface{}) //要做斷言檢查才能使用
								//goutils.Logger.Info("tableCode:" + tableCode + " 珠盤路[" + fmt.Sprint(idx) + "]:" + fmt.Sprint(betType))
							}
							currentCountingResult.BeadRoadStr = beadRoadBfr.String()
							goutils.Logger.Info("tableCode:" + tableCode + " 珠盤路:" + currentCountingResult.BeadRoadStr)

						}
						_isKeepPreviousSuggestion := currentCountingResultInterface.IsKeepPreviousSuggestion()
						//餵牌 餵路紙 做計算
						gotResult := currentCountingResultInterface.Counting(currentCountingResult.CardList, currentCountingResult.BeadRoadStr)
						if _isKeepPreviousSuggestion {
							gotResult = _isKeepPreviousSuggestion
						}
						if gotResult {
							//有預測結果了
							goutils.Logger.Info("tableCode:" + tableCode + " 有預測結果了 決定告知預測")
							NotifySuggest(currentCountingResult) //決定告知預測
						} else {
							//這局沒有預測結果，清除上一期預測結果(已公布過的)
							goutils.Logger.Info("tableCode:" + tableCode + " 沒有預測結果 ClearGuessResult")
							currentCountingResultInterface.ClearGuessResult() //這裡會把剛剛要公布的結果也刪掉，所以這裡只清預測結果
						}
					}
				}

			}

		}
	}
}

func GetGameIDDisplayHand(gameIDDisplay string) string {

	strArr := strings.Split(gameIDDisplay, "-")
	result := "0"
	if len(strArr) == 4 {
		result = strArr[3]
	}
	return result
}

//發佈建議的結果(公布答案) 公佈預測結果(有沒有猜中)
func NotifyGameResult(currentCountingResult *models.CountingResult) {
	PublishGameResult(currentCountingResult) //ws
	//QQ
	PublishGameResultToQQ(currentCountingResult, BetAccount)
}

//決定告知預測 發佈建議
func NotifySuggest(currentCountingResult *models.CountingResult) {
	PublishCountingSuggest(currentCountingResult) //ws
	//QQ
	PublishCountingSuggestToQQ(currentCountingResult, BetAccount)
}

//發佈目前帳戶金額
func NotifyCurrentBalance() {
	PublishAccountBalance(BetAccount) //ws
	//QQ

}

//發佈目前 下注/派彩 行為
func NotifyBetRecord(betRecord models.BetRecord) {
	PublishBet(betRecord) //ws
	//QQ
	if betRecord.Settled {
		PublishSettleBetActionToQQ(betRecord)
	} else {
		PublishPlaceBetActionToQQ(betRecord)
	}

	NotifyCurrentBalance()
}

//發佈過去24小時的下注結果
func NotifyCurrentBetStatistic() {
	//
	//PublishAccountBalance(BetAccount) //ws
	//QQ

	PublishBetStatisticToQQ(BetAccount)

}

//取得 BU001 TABLE 的 資料  tableCode := "0001005"
func fetchTableData(_tableInfo *tableInfo) {
	tableCode := _tableInfo.TableCode

	////var duration time.Duration = 1 //1 秒取一次
	////for _ = range time.Tick(duration * time.Second) {
	ticker := time.NewTicker(time.Millisecond * 200)
	tableInfoMap[tableCode].client = &http.Client{Timeout: connectTableTimeout}
	for _ = range ticker.C {
		/*
			timestamp := time.Now().Local()
			str := "fetchTableData TableCode:" + tableCode + " => at " + timestamp.String()
			fmt.Println(str)
		*/

		connectTable(tableCode) //改成自己跑遞回...會吃掉所有記憶體
	}
}

//取得 BU001 TABLE 的 資料
func connectTable(tableCode string) {

	millisecond := fmt.Sprint((time.Now().UnixNano()))
	//goutils.Logger.Info("connectTable TableCode:" + tableCode + " time.Millisecond:" + millisecond)
	urlStr := "http://spi.mld.v9vnb.org/GetData.ashx?tablecode=" + tableCode + "&valuetype=INIT&t=" + millisecond
	client := tableInfoMap[tableCode].client
	resp, err := client.Get(urlStr)
	if err != nil {
		goutils.Logger.Error("connectTable Get:"+tableCode+" Error:", err.Error())
		//connectTable(tableCode) //改成自己跑遞回...會吃掉所有記憶體
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			goutils.Logger.Error("connectTable ReadAll:"+tableCode+" Error:", err.Error())
			//connectTable(tableCode) //改成自己跑遞回...會吃掉所有記憶體
		} else {
			bodyStr := string(body)
			if tableInfoMap[tableCode].bodyStr != bodyStr { //內容變了在處理
				tableInfoMap[tableCode].bodyStr = bodyStr
				goutils.Logger.Info("送出:" + tableCode + " t:" + millisecond)
				tableResult <- TableInitJsonStr{TableCode: tableCode, JsonStr: body} //傳資料出去
			}
			goutils.Logger.Info("connectTable Get:" + tableCode + " t:" + millisecond)
			//connectTable(tableCode) //改成自己跑遞回...會吃掉所有記憶體
		}
		//goutils.CheckErr(err)
	}

	//goutils.Logger.Info("body:" + string(body))

}
