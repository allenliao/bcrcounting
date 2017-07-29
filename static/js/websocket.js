var socket;

function TransBetTypeToStr(betType) {
	switch (betType) {
	case 0:
		return "莊"
	case 1:
		return "閒"
	case 2:
		return "和"
	}
	return ""
}

$(document).ready(function () {
    // Create a socket
    socket = new WebSocket('ws://' + window.location.host + '/ws/join?uname=' + $('#uname').text());
    // Message received on the socket
    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);
        console.log(data);
        var ContentObj,msgStr;
        switch (data.Type) {
            /*
        case 0: // JOIN
            if (data.User == $('#uname').text()) {
                $("#chatbox li").first().before("<li>You joined the chat room.</li>");
            } else {
                $("#chatbox li").first().before("<li>" + data.User + " joined the chat room.</li>");
            }
            break;
        case 1: // LEAVE
            $("#chatbox li").first().before("<li>" + data.User + " left the chat room.</li>");
            break;
        case 2: // MESSAGE
            $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + data.Content + "</li>");
            break;
            */
        case 3: // EVENT_SUGGESTION
            ContentObj=JSON.parse(data.Content)
            SuggestionBetStr=TransBetTypeToStr(ContentObj.SuggestionBet)
            msgStr="第 " +ContentObj.TableNo + " 桌 " + ContentObj.GameIDDisplay + " 下一局建議買 " + SuggestionBetStr + " (" + ContentObj.TrendName + ")"
            $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + msgStr + "</li>");
            break;
        case 4: // RESULT
            ContentObj=JSON.parse(data.Content)
            var guessResultStr
			if (ContentObj.TieReturn) {
				guessResultStr = "平"

			} else {

				if (ContentObj.GuessResult ) {
					guessResultStr = "勝"
				} else {
					guessResultStr = "負"
				}

			}
			if (ContentObj.FirstHand) {
				guessResultStr = "第一局預測不記結果"
			}

			msgStr = "第 " + ContentObj.TableNo+ " 桌 " + ContentObj.GameIDDisplay + " 開 " + TransBetTypeToStr(ContentObj.Result) + " 建議結果:" + guessResultStr
            $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + msgStr + "</li>");
            break;
        case 5: // EVENT_ACCOUNT
            $("#ubalance").html(data.Content);
            break;
        case 6: // EVENT_BET
        var BetType="閒"
            
            ContentObj=JSON.parse(data.Content)
            if(ContentObj.Settled){
                $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + ContentObj.BetTime + " 第 "+ContentObj.TableNo+" 桌"+ContentObj.GameIDDisplay+" 開 "+ ContentObj.GameResultTypeStr+" "+ContentObj.WinAmmount+" 帳戶餘額:"+ContentObj.CurrentBalance+"</li>");
            }else{
                $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + ContentObj.BetTime + " 第 "+ContentObj.TableNo+" 桌"+ContentObj.GameIDDisplay+" 買 "+ ContentObj.BetTypeStr+" "+ContentObj.BetAmmount+" 帳戶餘額:"+ContentObj.CurrentBalance+"</li>");
            }
            break;
        }
    };

    // Send messages.
    var postConecnt = function () {
        var uname = $('#uname').text();
        var content = $('#sendbox').val();
        socket.send(content);
        $('#sendbox').val("");
    }

    $('#sendbtn').click(function () {
        postConecnt();
    });
});