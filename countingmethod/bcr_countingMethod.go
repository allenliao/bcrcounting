//百家樂第一種計算方法
//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響

package countingmethod

import "bcrcounting/models"

var ()

func GetCountingResult(BUCode string, tableNo uint8) models.CountingResult {
	var betSuggestionData [3]models.BetSuggestion
	betSuggestionData[0] = models.BetSuggestion{
		BetType:    models.Bcr_BETTYPE_BANKER,
		HouseEdge:  models.Bcr_BankerHouseEdgeDefault,
		SuggestBet: false}
	betSuggestionData[0] = models.BetSuggestion{
		BetType:    models.Bcr_BETTYPE_PLAYER,
		HouseEdge:  models.Bcr_PlayerHouseEdgeDefault,
		SuggestBet: false}
	betSuggestionData[0] = models.BetSuggestion{
		BetType:    models.Bcr_BETTYPE_BANKER,
		HouseEdge:  models.Bcr_PlayerHouseEdgeDefault,
		SuggestBet: false}

	currentCountingResult := models.CountingResult{
		BUCode:              BUCode,
		TableNo:             tableNo,
		BetSuggestionData:   betSuggestionData,
		SuggestionBet:       "",
		SuggestionBetAmount: 100,
		Result:              "",
		GuessResult:         true}

	return currentCountingResult
}

//換靴時要呼叫
func resetCountingResult(countingResult models.CountingResult) {
	//countingResult
}
