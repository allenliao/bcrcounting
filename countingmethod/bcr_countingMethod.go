package countingmethod

import (
	"bcrcounting/models"
)

//百家樂第一種計算方法
//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響

func CreateCurrentCountingResultList(BUCode string, tableNo uint8) map[string]models.CountingResultInterface {
	//注冊算法
	currentCountingResultList := map[string]models.CountingResultInterface{
		"cardCounting": models.CountingResultMethod1{},
		"longtrend":    models.CountingResultMethod2{}}

	for _, methodObj := range currentCountingResultList {
		methodObj.CreateCountingResult(BUCode, tableNo)
	}

	return currentCountingResultList

}

//換靴時要呼叫
func resetCountingResult(countingResult models.CountingResult) {
	//countingResult
}

/*
//長龍
func Bcr_CountingMethod_Trand1(cardList [6]int, currentCountingResult *models.CountingResult) *models.BetSuggestion {

}
*/
