/*
 * Copyright (c) 2018, Ryan Westlund.
 * This code is under the BSD 3-Clause license.
 */

var newMsg = ''; // Holds new messages to be sent to the server
var chatContent = ''; // A running list of chat messages displayed on the screen
var username = null; // Our username
var socket = new WebSocket('ws://' + window.location.host + '/ws');
// This variable is set later, in the battle() function. It has to be initialized here so that other functions can have access to it.
var inputter;

socket.onmessage = function(e) {
  var msg = JSON.parse(e.data);
  if (msg.hasOwnProperty('message')) {
    handleChatMessage(msg)
  } else {
    handleBattleUpdate(msg)
  }
};

function handleChatMessage(msg) {
  if (msg.command == "START GAME") {
    document.getElementById('ownName').innerHTML=username
    document.getElementById('enemyName').innerHTML=msg.message
    document.getElementById("matchSound").play()
    document.getElementById("readyButton").innerHTML="Ready for game"
    document.getElementById('chat').style.display="none"
    document.getElementById('battleUI').style.display="block"
    document.getElementById("getReadyText").style.display="block"
    setTimeout(function(){
      setTimeout(function(){
        setTimeout(function(){
          setTimeout(function(){
            document.getElementById("getReadyText").innerHTML="Get ready!"
            document.getElementById("getReadyText").style.display="none"
            document.getElementById("startSound").play()
            battle()
          },1000)
          document.getElementById("getReadyText").innerHTML="1..."
          document.getElementById("countdownSound").fastSeek(0)
          document.getElementById("countdownSound").play()
        },1000)
        document.getElementById("getReadyText").innerHTML="2..."
        document.getElementById("countdownSound").fastSeek(0)
        document.getElementById("countdownSound").play()
      },1000)
      document.getElementById("getReadyText").innerHTML="3..."
      document.getElementById("countdownSound").play()
    },1000)
    return
  }
  chatContent += '<div class="chip">'
   + msg.username
   + '</div>'
   + (msg.message) + '<br/>';
  var element = document.getElementById('chat-messages');
  element.innerHTML=chatContent;
  element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
};

function send () {
    newMsg=document.getElementById("msgbox").value;
    console.log("newmsg is"+newMsg);
    if (newMsg != '') {
        socket.send(
            JSON.stringify({
                username: username,
                message: newMsg,
		command: ""
            }
        ));
        document.getElementById("msgbox").value=""; // Reset the message box
    }
}

function join () {
    username = document.getElementById("usernamebox").value;
    console.log("username is "+username);
    if (!username) {
        Materialize.toast('You must choose a username', 2000);
        return;
    }
    document.getElementById("afterjoin").style.display = "block";
    document.getElementById("beforejoin").style.display = "none";
    socket.send(
      JSON.stringify({
        username:username,
        message: "",
        command: "SETNAME"
    }));
}

function toggleReady () {
    readyStatus = document.getElementById("readyButton").innerHTML;
    console.log("(Un)readying for game...");
    if (readyStatus.search("Unready for game")==-1) {
        var command = "READY";
        document.getElementById("readyButton").innerHTML="Unready for game";
    } else {
        var command = "UNREADY";
        document.getElementById("readyButton").innerHTML="Ready for game";
    }
    socket.send(
        JSON.stringify({
            username:username,
            message: "",
            command: command
	}
    ));
}

function enter (event) {
    if (event.keyCode == 13) {
        send()
    }
}

function fightBot() {
	socket.send(
	  JSON.stringify({
	    username:username,
	    message: document.getElementById("botMenu").value,
	    command: "BOT MATCH"
	  })
	);
}

function toggleInstructions () {
	element=document.getElementById("instructions")
	console.log(element.style.display)
	if (element.style.display=="none") {
		element.style.display="block"
	} else {
		element.style.display="none"
	}
	element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
}

function handleBattleUpdate(update) {
  if (update.self.life<=0 || update.enemy.life<=0) {
    document.getElementById('battleUI').style.display="none"
    document.getElementById('chat').style.display="block"
    // Display a message telling the result of the battle.
    chatContent += '<div class="chip">'
     + "server"
     + "</div>"
     + "Result of battle: you had "+update.self.life.toString()+" life and the enemy had "+update.enemy.life.toString()+"<br>";
    var element = document.getElementById('chat-messages');
    element.innerHTML=chatContent;
    element.scrollTop = element.scrollHeight;

    clearInterval(inputter)
    socket.send(JSON.stringify({
      "username":username,
      "message":"",
      "command":"END MATCH"
    }));
  }
  document.getElementById('ownLife').style.width=update.self.life.toString()+"%"
  document.getElementById('ownStam').style.width=update.self.stamina.toString()+"%"
  document.getElementById('ownDuration').style.width=update.self.stateDur.toString()+"%"
  document.getElementById('enemyLife').style.width=update.enemy.life.toString()+"%"
  document.getElementById('enemyStam').style.width=update.enemy.stamina.toString()+"%"
  document.getElementById('enemyDuration').style.width=update.enemy.stateDur.toString()+"%"
  var ownState=update.self.state
  var enemyState=update.enemy.state
  document.getElementById('ownBlockSymbol').style.display="none"
  document.getElementById('ownLightSymbol').style.display="none"
  document.getElementById('ownLeftLightSymbol').style.display="none"
  document.getElementById('ownHeavySymbol').style.display="none"
  document.getElementById('enemyBlockSymbol').style.display="none"
  document.getElementById('enemyLightSymbol').style.display="none"
  document.getElementById('enemyHeavySymbol').style.display="none"
  document.getElementById('enemyRightLightSymbol').style.display="none"
  document.getElementById('leftArrowSymbol').style.display="none"
  document.getElementById('upArrowSymbol').style.display="none"
  document.getElementById('rightArrowSymbol').style.display="none"
  document.getElementById('downArrowSymbol').style.display="none"
  switch (ownState) {
    case "standing":
      break
    case "blocking":
      document.getElementById('ownBlockSymbol').style.display="inline-block"
      break;
    case "light attack":
      document.getElementById('ownLightSymbol').style.display="inline-block"
      break;
    case "heavy attack":
      document.getElementById('ownHeavySymbol').style.display="inline-block"
      break;
    case "counterattack":
      document.getElementById('ownBlockSymbol').style.display="inline-block"
      document.getElementById('ownLightSymbol').style.display="inline-block"
      break;
    case "countered":
      document.getElementById('ownLeftLightSymbol').style.display="inline-block"
      document.getElementById('ownBlockSymbol').style.display="inline-block"
      break;
    default:
      // If we're the heavy attack player
      if (ownState.search("interrupted")==0) {
        arrow=ownState.slice(ownState.indexOf("_")+1,ownState.length)
        document.getElementById('ownHeavySymbol').style.display="inline-block"
        document.getElementById(arrow+'ArrowSymbol').style.display="inline-block"
      } else {
        arrow=ownState.slice(ownState.indexOf("_")+1,ownState.length)
        document.getElementById('ownLightSymbol').style.display="inline-block"
        document.getElementById(arrow+'ArrowSymbol').style.display="inline-block"
      }

  }
  switch (enemyState) {
    case "standing":
      break
    case "blocking":
      document.getElementById('enemyBlockSymbol').style.display="inline-block"
      break;
    case "light attack":
      document.getElementById('enemyLightSymbol').style.display="inline-block"
      break;
    case "heavy attack":
      document.getElementById('enemyHeavySymbol').style.display="inline-block"
      break;
    case "counterattack":
      document.getElementById('enemyBlockSymbol').style.display="inline-block"
      document.getElementById('enemyLightSymbol').style.display="inline-block"
      break;
    case "countered":
      document.getElementById('enemyRightLightSymbol').style.display="inline-block"
      document.getElementById('enemyBlockSymbol').style.display="inline-block"
      break;
    default:
      if (ownState.search("interrupted")==0) {
        document.getElementById('enemyLightSymbol').style.display="inline-block"
      } else {
        document.getElementById('enemyHeavySymbol').style.display="inline-block"
      }
  }
};


function battle () {
  var input = "NONE"
  document.addEventListener('keyup', function(e) {
    if (e.keyCode == 32) {
      input = "NONE"
      }
    return
  });
  document.addEventListener('keydown', function(e) {
    switch (e.keyCode) {
      case 32:
        input = "BLOCK"
        return
      case 81:
        input = "LIGHT"
        return
      case 87:
        input = "HEAVY"
        return
      case 37:
        input = "INTERRUPT_LEFT"
        return
      case 38:
        input = "INTERRUPT_UP"
        return
      case 39:
        input = "INTERRUPT_RIGHT"
        return
      case 40:
        input = "INTERRUPT_DOWN"
        return
    }
    if (e.shiftKey) {
      input = "DODGE"
      return
    }
    if (e.ctrlKey) {
      input = "SAVE"
      return
    }
  });

  function sendUpdate () {
    socket.send(JSON.stringify({
    "username":username,
    "message":input,
    "command":""}))
    if (input!="BLOCK"){
      input = "NONE"
    }
  }
  // This interval has to be longer than the server's mainloop interval, otherwise updates will get lost.
  inputter = setInterval(sendUpdate, 20)
}
