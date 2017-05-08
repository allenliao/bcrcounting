//百家樂第一種計算方法
//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響

package countingmethod

import (
	"bcrcounting/models"
	"fmt"
	"math"
	"sort"

	"github.com/astaxie/beego"
)

var (
	maxBanker, maxTie, maxPlayer float64 = -1, -1, -1
)

func CreateCountingResult(BUCode string, tableNo uint8) models.CountingResult {
	var betSuggestionMap = make(map[int]*models.BetSuggestion)
	betSuggestionMap[models.Bcr_BETTYPE_BANKER] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_BANKER,
		HouseEdge:    models.Bcr_BankerHouseEdgeDefault,
		IsSuggestBet: false}
	betSuggestionMap[models.Bcr_BETTYPE_PLAYER] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_PLAYER,
		HouseEdge:    models.Bcr_PlayerHouseEdgeDefault,
		IsSuggestBet: false}
	betSuggestionMap[models.Bcr_BETTYPE_TIE] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_TIE,
		HouseEdge:    models.Bcr_TieHouseEdgeDefault,
		IsSuggestBet: false}

	currentCountingResult := models.CountingResult{
		BUCode:              BUCode,
		TableNo:             tableNo,
		BetSuggestionMap:    betSuggestionMap,
		SuggestionBet:       "",
		SuggestionBetAmount: 100,
		Result:              "",
		GuessResult:         false}

	currentCountingResult.BetSuggestionSliceForSort = make(models.BetSuggestionSlice, 0, len(betSuggestionMap))

	for _, betSuggestion_adr := range betSuggestionMap {
		currentCountingResult.BetSuggestionSliceForSort = append(currentCountingResult.BetSuggestionSliceForSort, betSuggestion_adr)
	}

	return currentCountingResult
}

//換靴時要呼叫
func resetCountingResult(countingResult models.CountingResult) {
	//countingResult
}

//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響
//
func Bcr_CountingMethod1(cardList [6]int, currentCountingResult *models.CountingResult) bool {

	for _, point := range cardList { //idx, card point
		if point == -1 {
			continue
		}
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_PLAYER].HouseEdge += models.Bcr_PlayerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_BANKER].HouseEdge += models.Bcr_BankerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_TIE].HouseEdge += models.Bcr_TieHouseEdgeEffectList[point]
	}
	/*
	   //for test
	   	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_PLAYER].HouseEdge = -5
	   	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_BANKER].HouseEdge = 2
	   	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_TIE].HouseEdge = -1
	*/

	//HouseEdge排序算出結果(越大越好)
	sort.Sort(currentCountingResult.BetSuggestionSliceForSort)
	betSuggestion := currentCountingResult.BetSuggestionSliceForSort[0] //最大的

	maxBanker = math.Max(float64(currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_BANKER].HouseEdge), maxBanker)
	maxPlayer = math.Max(float64(currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_PLAYER].HouseEdge), maxPlayer)
	maxTie = math.Max(float64(currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_TIE].HouseEdge), maxTie)
	beego.Info("maxBanker:" + fmt.Sprint(maxBanker) + " maxPlayer:" + fmt.Sprint(maxPlayer) + " maxTie:" + fmt.Sprint(maxTie))
	for idx, betSuggestion := range currentCountingResult.BetSuggestionSliceForSort {
		beego.Info("[" + fmt.Sprint(idx) + "]betSuggestion BetType:" + fmt.Sprint(betSuggestion.BetType) + " HouseEdge:" + fmt.Sprint(betSuggestion.HouseEdge))
	}

	if betSuggestion.HouseEdge > 0 { //擊敗賭場優勢 //除非有退庸，不然不可能>0
		//TODO:改成個注別大於某一個統計數字就公佈，以統計勝率當作權重
		betSuggestion.IsSuggestBet = true
		currentCountingResult.SuggestionBet = models.TransBetTypeToStr(betSuggestion.BetType) //建議下一局買甚麼
		return true
	}

	return false

}

/*
//長龍
func Bcr_CountingMethod_Trand1(cardList [6]int, currentCountingResult *models.CountingResult) *models.BetSuggestion {

}
*/
