var socket;

$(document).ready(function () {
    // Create a socket
    socket = new WebSocket('ws://' + window.location.host + '/ws/join?uname=' + $('#uname').text());
    // Message received on the socket
    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);
        console.log(data);
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
            $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + data.Content + "</li>");
            break;
        case 4: // RESULT
            $("#chatbox li").first().before("<li><b>" + data.User + "</b>: " + data.Content + "</li>");
            break;
        case 5: // EVENT_ACCOUNT
            $("#ubalance").html(data.Content);
            break;
        case 6: // EVENT_BET
        var BetType="閒"
            
            var ContentObj=JSON.parse(data.Content)
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