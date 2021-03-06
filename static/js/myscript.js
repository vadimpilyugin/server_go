var fileToEdit = ""

function deleteLink (s) {
  let isConfirm = confirm('Точно?');
  if (!isConfirm) {
    return;
  }
  console.log("Delete link clicked; url: ", s);
  var request = new XMLHttpRequest();
  request.open('DELETE', s, true);

  request.onload = function() {
    if (request.status >= 200 && request.status < 400) {
      // Success!
      console.log("Successfull DELETE request, redirecting to ---> ", window.location);
      location.reload(true);
    } else {
      // We reached our target server, but it returned an error
      confirm("We reached our target server, but it returned an error");
    }
  };

  request.onerror = function() {
    // There was a connection error of some sort
    confirm("There was a connection error of some sort")
  };

  request.send();
  return;
}

var lastHovered;

function generateQrCode (el, s) {
  let img = document.querySelector("#barcodeImg");
  let data = window.location+s;
  img.src = `https://api.qrserver.com/v1/create-qr-code/?size=500x500&data=${data}`;
  showImage();
  lastHovered = el.parentElement.parentElement;
  lastHovered.classList.remove("hashover");
  console.log(`generated qr code for data='${data}'`);
  return true;
}

function hideImage () {
  if (isImageShown()) {
    let overlay = document.querySelector(".barcode-overlay");
    overlay.classList.remove('show');
    overlay.classList.add('hide');
    lastHovered.classList.add("hashover");
  }
}

function showImage () {
  let overlay = document.querySelector(".barcode-overlay");
  overlay.classList.remove('hide');
  overlay.classList.add('show');
}

function isImageShown () {
  let overlay = document.querySelector(".barcode-overlay");
  return overlay.classList.contains('show');
}

var isShown = false;
var clickImmunity = false;
var isEntered = false;

function makeDraggable (el) {
  // Make the DIV element draggable:
  dragElement(el.parentElement);

  function dragElement(elmnt) {
    var pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;
    if (el) {
      // if present, the header is where you move the DIV from:
      el.onmousedown = dragMouseDown;
    } else {
      // otherwise, move the DIV from anywhere inside the DIV: 
      elmnt.onmousedown = dragMouseDown;
    }

    function dragMouseDown(e) {
      e = e || window.event;
      e.preventDefault();
      // get the mouse cursor position at startup:
      pos3 = e.clientX;
      pos4 = e.clientY;
      document.onmouseup = closeDragElement;
      // call a function whenever the cursor moves:
      document.onmousemove = elementDrag;
    }

    function elementDrag(e) {
      clickImmunity = true;
      e = e || window.event;
      e.preventDefault();
      // calculate the new cursor position:
      pos1 = pos3 - e.clientX;
      pos2 = pos4 - e.clientY;
      pos3 = e.clientX;
      pos4 = e.clientY;
      // set the element's new position:
      elmnt.style.top = (elmnt.offsetTop - pos2) + "px";
      elmnt.style.left = (elmnt.offsetLeft - pos1) + "px";
      let video = document.querySelector("#videoFrame");
      video.removeAttribute("controls");
    }

    function closeDragElement() {
      // stop moving when mouse button is released:
      document.onmouseup = null;
      document.onmousemove = null;
    }
  }
}

function fullScreenCancel() {
  if(document.cancelFullscreen) {
    console.log("Something!");
    document.cancelFullscreen();
  } else if(document.webkitCancelFullScreen ) {
    console.log("Something!");
    document.webkitCancelFullScreen();
  } else if(document.mozCancelFullScreen) {
    console.log("Something!");
    document.mozCancelFullScreen();
  }
}

function isFullscreen() {
  // if(document.isFullscreen) {
  //   console.log("Something!");
  //   return document.isFullscreen;
  // } else if(document.webkitIsFullScreen ) {
  //   console.log("Something!");
  //   return document.webkitIsFullScreen;
  // } else if(document.mozIsFullScreen) {
  //   console.log("Something!");
  //   document.mozIsFullScreen;
  // }
  if (document.webkitIsFullScreen)
    console.log("Something!");
  return document.webkitIsFullScreen;
}

function hideVideo () {
  if (!isShown || clickImmunity || isFullscreen())
    return;
  isShown = false;
  let video = document.querySelector("#videoFrame");
  let overlay = document.querySelector(".video-overlay");
  overlay.classList.remove('show');
  overlay.classList.add('hide');
  video.pause();
}

function showVideo (e, s) {
  e.preventDefault();
  e.stopPropagation();
  isShown = true;
  let overlay = document.querySelector(".video-overlay");
  overlay.classList.remove('hide');
  overlay.classList.add('show');
  let video = document.querySelector("#videoFrame");
  makeDraggable(video);
  let source = document.querySelector("#videoSource");
  video.pause();
  source.setAttribute('src', s);
  video.load();
}

function videoEnter () {
  isEntered = true;
  document.querySelector("body").classList.add('no-scroll');
  console.log("Video enter:", isEntered);
  let video = document.querySelector("#videoFrame");
  if (!video.hasAttribute("controls") && video.offsetWidth >= 400) {
    video.setAttribute("controls","controls")   
  }
}

function videoLeave () {
  isEntered = false;
  clickImmunity = false;
  document.querySelector("body").classList.remove('no-scroll');
  console.log("Video left:", isEntered);
}

function loadData(path) {
  var request = new XMLHttpRequest();
  request.open('GET', path, true);
  request.setRequestHeader("X-Codemirror", "codemirror")

  request.onload = function() {
    if (request.status >= 200 && request.status < 400) {
      // Success!
      var resp = request.responseText;
      editor.setValue(resp);
    } else {
      // We reached our target server, but it returned an error

    }
  };

  request.onerror = function() {
    // There was a connection error of some sort
  };

  request.send();
}

function saveData() {      
  // your CodeMirror textarea ID
  var textToWrite = editor.getValue();

  console.log("Saving: ", textToWrite.length, " characters")

  var formData = new FormData();

  var textFileAsBlob = new Blob([textToWrite], {type:'text/plain'});

  formData.append("file", textFileAsBlob, fileToEdit);

  console.log("fileToEdit:", fileToEdit)

  var request = new XMLHttpRequest();
  request.open("POST", './');
  request.send(formData);
}

function showEditor (thisEl, path) {
  let overlay = document.querySelector("#overlay");
  overlay.classList.remove('hide');
  overlay.classList.add('show');
  loadData(path);
  var lastHovered = thisEl.parentElement.parentElement;
  lastHovered.classList.remove("hashover");
  setTimeout(function() {
    lastHovered.classList.add("hashover");
  }, 100);
  fileToEdit = path;
}

function showEditorByLink (ev, path) {
  ev.preventDefault();
  let overlay = document.querySelector("#overlay");
  overlay.classList.remove('hide');
  overlay.classList.add('show');
  loadData(path);
  fileToEdit = path;
}

function hideEditor () {
  let overlay = document.querySelector(".codemirror-overlay");
  overlay.classList.remove('show');
  overlay.classList.add('hide');
}


function hideOverlay(e) {
  if (!e.target.hasAttribute("closer") && e.currentTarget != e.target) {
    // do nothing
    console.log("click on inner frame")
    console.log("the element:", e.target)
    foo = e.target;
  } else {
    let overlay = document.querySelector("#overlay");
    overlay.classList.remove('show');
    overlay.classList.add('hide');
  }

}