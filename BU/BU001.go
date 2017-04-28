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
	BUCode                string
	tableResult           chan TableInitJsonStr
	currentCountingResult models.CountingResult
	//TODO 要準備一個 [TableCode]來放 當下這一局的 結果

	//A~9
)

type TableInitJsonStr struct {
	TableCode string
	JsonStr   []byte // Only for WebSocket users; otherwise nil.
}

func init() {
	initDefaultValue()
	StartProcess()

}

//初始化變數
func initDefaultValue() {
	BUCode = "BU001"
	tableResult = make(chan TableInitJsonStr, 10)

	currentCountingResult = countingmethod.GetCountingResult(BUCode)

	//PublishCountingResult(countingResult) //決定告知結果
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
		if _tableResult, ok := <-tableResult; ok {

			jsonObj, err := simplejson.NewJson(_tableResult.JsonStr)
			goutils.CheckErr(err)
			shoeID, _ := jsonObj.Get("DCGameVO").Get("shoeID").Int()
			gameIDDisplay, _ := jsonObj.Get("DCGameVO").Get("gameIDDisplay").String()
			gameStatus, _ := jsonObj.Get("gameStatus").String()
			beego.Info("shoeID:" + fmt.Sprint(shoeID) + " gameIDDisplay:" + gameIDDisplay)

			//gameStatus= 1=init 2=bet 3=dealing 4=resulting 5=end
			if gameIDDisplay != currentCountingResult.GameIDDisplay && gameStatus == "4" {
				//紀錄、回寫結果
				//logHistory()

				//算牌
				//currentCountingResult.
				//countingCard()
				//判斷勝率是否夠高，決定是否告知

			}

			//TODO 實作處理結果後回傳

			//PublishCountingResult(countingResult) //決定告知結果
		}
	}
}

func countingCard() {

}

func logHistory() {
	//currentCountingResult.Result=
}

//取得 BU001 TABLE 的 資料 (目前只先拿第二桌)
func fetchTableData() {
	timestamp := time.Now().Local()
	var duration time.Duration = 1 //1 秒取一次
	for _ = range time.Tick(duration * time.Second) {
		tableCode := "0001002"
		str := "fetchTableData table:" + tableCode + "> at " + timestamp.String()
		fmt.Println(str)

		connectTable(tableCode)
	}
}

//取得 BU001 TABLE 的 資料
func connectTable(tableCode string) {

	millisecond := fmt.Sprint((time.Now().UnixNano()))
	beego.Info("connectTable time.Millisecond:" + millisecond)
	resp, err := http.Get("http://spi.mld.v9vnb.org/GetData.ashx?tablecode=" + tableCode + "&valuetype=INIT&t=" + millisecond)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	goutils.CheckErr(err)
	//beego.Info("body:" + string(body))

	tableResult <- TableInitJsonStr{TableCode: tableCode, JsonStr: body} //傳資料出去

}
