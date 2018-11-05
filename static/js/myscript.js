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