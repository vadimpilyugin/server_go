function deleteLink (s) {
  let isConfirm = confirm('Точно?');
  if (!isConfirm) {
    return true;
  }
  console.log("Delete link clicked; url: ", s);
  var request = new XMLHttpRequest();
  request.open('DELETE', s, true);

  request.onload = function() {
    if (request.status >= 200 && request.status < 400) {
      // Success!
      console.log("Successfull DELETE request, redirecting to ---> ", window.location);
      location.reload();
    } else {
      // We reached our target server, but it returned an error
      alert("We reached our target server, but it returned an error");
    }
  };

  request.onerror = function() {
    // There was a connection error of some sort
    alert("There was a connection error of some sort")
  };

  request.send();
  return true;
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

function hideVideo () {
  if (!isShown || clickImmunity)
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
  isShown = true;
  event.stopPropagation();
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