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
	"container/list"
	"encoding/json"
	"goutils"
	"math"
	"time"

	"github.com/gorilla/websocket"

	"bcrcounting/models"
	"fmt"
)

type Subscription struct {
	Archive []models.Event      // All the events from the archive.
	New     <-chan models.Event // New events coming in.
}

func newEvent(ep models.EventType, user, msg string) models.Event {
	return models.Event{ep, user, int(time.Now().Unix()), msg}
}

func Join(user string, ws *websocket.Conn) {
	subscribe <- Subscriber{Name: user, Conn: ws} //send value to channel
}

//發佈目前下注狀況
func PublishBet(_betRecord models.BetRecord) {
	betRecordCh <- _betRecord

}

//發佈目前帳戶金額
func PublishAccountBalance(_betAccount *models.SimBetAccount) {
	betAccountCh <- _betAccount
}

//發佈建議的結果(公布答案)
func PublishGameResult(_countingResult *models.CountingResult) {
	countingResultCh <- _countingResult
}

//發佈建議
func PublishCountingSuggest(_countingResult *models.CountingResult) {
	countingSuggestCh <- _countingResult
}

func Leave(user string) {
	unsubscribe <- user
}

type Subscriber struct {
	Name string
	Conn *websocket.Conn // Only for WebSocket users; otherwise nil.
}

var (
	betAccountCh      = make(chan *models.SimBetAccount, 10)
	betRecordCh       = make(chan models.BetRecord, 10)
	countingResultCh  = make(chan *models.CountingResult, 10)
	countingSuggestCh = make(chan *models.CountingResult, 10)
	// Channel for new join users.
	subscribe = make(chan Subscriber, 10)
	// Channel for exit users.
	unsubscribe = make(chan string, 10)
	// Send events here to publish them.
	publish = make(chan models.Event, 10)
	// Long polling waiting list.
	waitingList = list.New()
	subscribers = list.New()
)

// This function handles all incoming chan messages.
func chatroom() {
	for {
		select {
		case _betRecord := <-betRecordCh:
			contentTyp := "下注:"
			//"下注:"/"派彩:"
			if _betRecord.Settled {
				contentTyp = "派彩:"
			}
			_betRecordStr, err := json.Marshal(_betRecord)
			goutils.CheckErr(err)
			publish <- newEvent(models.EVENT_BET, contentTyp, string(_betRecordStr))

		case _betAccount := <-betAccountCh:
			//更新帳戶
			publish <- newEvent(models.EVENT_ACCOUNT, "目前帳戶:", fmt.Sprint(_betAccount.Balance))

		case _countingSuggest := <-countingSuggestCh:
			//提供建議
			SuggestionBetStr := models.TransBetTypeToStr(_countingSuggest.SuggestionBet)
			if SuggestionBetStr != "" {

				TrendMethodWinRate := math.Abs(math.Ceil(_countingSuggest.TrendMethodWinRate*10000)) / 100
				msg := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " 下一局建議買 " + SuggestionBetStr + " (" + _countingSuggest.TrendName + ") TrendMethodWinRate:" + fmt.Sprint(TrendMethodWinRate)
				goutils.Logger.Info("TableNo:" + fmt.Sprint(_countingSuggest.TableNo) + " *真* 提供建議 msg:" + msg)
				_countingSuggestStr, err := json.Marshal(_countingSuggest)
				goutils.CheckErr(err)
				publish <- newEvent(models.EVENT_SUGGESTION, "建議:", string(_countingSuggestStr))
			}
			/*
				SuggestionBetStr := models.TransBetTypeToStr(_countingSuggest.SuggestionBet)
				if SuggestionBetStr != "" {
					msg := "第 " + fmt.Sprint(_countingSuggest.TableNo) + " 桌 " + _countingSuggest.GameIDDisplay + " 下一局建議買 " + SuggestionBetStr + " (" + _countingSuggest.TrendName + ")"
					publish <- newEvent(models.EVENT_SUGGESTION, "建議:", msg)
					goutils.Logger.Info("TableNo:" + fmt.Sprint(_countingSuggest.TableNo) + " *真* 提供建議")
				}
			*/

		case _countingResult := <-countingResultCh:
			//報告預測結果
			//該局有提供建議時才會填這一局的 Result
			_countingResultStr, err := json.Marshal(_countingResult)
			goutils.CheckErr(err)
			publish <- newEvent(models.EVENT_RESULT, "結果:", string(_countingResultStr))
			/*
				var guessResultStr string
				if _countingResult.TieReturn {
					guessResultStr = "平"

				} else {

					if _countingResult.GuessResult {
						guessResultStr = "勝"
					} else {
						guessResultStr = "負"
					}

				}
				if _countingResult.FirstHand {
					guessResultStr = "第一局預測不記結果"
				}

				msg := "第 " + fmt.Sprint(_countingResult.TableNo) + " 桌 " + _countingResult.GameIDDisplay + " 開 " + models.TransBetTypeToStr(_countingResult.Result) + " 建議結果:" + guessResultStr

				publish <- newEvent(models.EVENT_RESULT, "結果:", msg)
			*/
			goutils.Logger.Info("TableNo:" + fmt.Sprint(_countingResult.TableNo) + " *真* 公佈預測結果")
		case sub := <-subscribe:
			if !isUserExist(subscribers, sub.Name) {
				subscribers.PushBack(sub) // Add user to the end of list.
				// Publish a JOIN event.
				publish <- newEvent(models.EVENT_JOIN, sub.Name, "")
				goutils.Logger.Info("New user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			} else {
				goutils.Logger.Info("Old user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			}
		case event := <-publish:
			// Notify waiting list.
			for ch := waitingList.Back(); ch != nil; ch = ch.Prev() {
				ch.Value.(chan bool) <- true
				waitingList.Remove(ch)
			}
			//TODO:實作不同的call 不同機器人權限連線的  WebSocket
			//注冊時就依權限分類不同的 subscribers(定時去檢查清單，到期就踢掉一些人，或是發佈時檢查其條件還在不在去踢一些人)

			broadcastWebSocket(event) //給所有的 人
			models.NewArchive(event)

			if event.Type == models.EVENT_MESSAGE {
				goutils.Logger.Info("Message from", event.User, ";Content:", event.Content)
			}
		case unsub := <-unsubscribe:
			for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
				if sub.Value.(Subscriber).Name == unsub {
					subscribers.Remove(sub)
					// Clone connection.
					ws := sub.Value.(Subscriber).Conn
					if ws != nil {
						ws.Close()
						goutils.Logger.Error("WebSocket closed:", unsub)
					}
					publish <- newEvent(models.EVENT_LEAVE, unsub, "") // Publish a LEAVE event.
					break
				}
			}
		}
	}
}

func init() {

	go chatroom()
}

func isUserExist(subscribers *list.List, user string) bool {
	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		if sub.Value.(Subscriber).Name == user {
			return true
		}
	}
	return false
}
