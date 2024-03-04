let draggedElement = null;

function handleDragStart(e) {
  draggedElement = e.target;
}

function handleDragOver(e) {
  e.preventDefault();
  return false;
}

function handleDragEnter(e) {
  if (e.target.classList.contains("draggable-item")) {
    e.target.classList.add("over");
  }
}

function handleDragLeave(e) {
  e.target.classList.remove("over");
}

function handleDrop(e) {
  e.preventDefault();
  if (e.target === draggedElement) {
    return;
  }

  if (e.target.classList.contains("draggable-item")) {
    var toExchange = draggedElement.innerHTML
    var toExchangeLink1 = draggedElement.querySelector('a').href;
    var toExchangeLink2 = this.querySelector('a').href;
    console.log(toExchangeLink1, toExchangeLink2);
    var toSend = toExchangeLink1 + toExchangeLink2;
    console.log(toSend);
    fetch('http://localhost:3000/changeorder', {
    method: 'POST',
    headers: {
      'Content-Type': 'text/plain',
    },
    body: toSend
    })
    draggedElement.innerHTML = this.innerHTML;
    this.innerHTML = toExchange;

  }
}

function handleDragEnd(e) {
  let items = document.querySelectorAll('.draggable-item');
  items.forEach(function(item) {
    item.classList.remove("over");
  });
}

function handleDragExit(e) {
  let items = document.querySelectorAll('.draggable-item');
  items.forEach(function(item) {
    item.classList.remove("over");
  });
}

function handleAddItem(e) {
  let newLink = prompt("Enter the link to the new feed");
  if (newLink === null || newLink === "") {
    return;
  }
  fetch('http://localhost:3000/addlink', {
  method: 'POST', // or 'PUT'
  headers: {
    'Content-Type': 'text/plain',
  },
  body: newLink
  }).then(response => {
    if (!response.ok) { // If response is not ok, throw an error
      alert("Link not valid or already exists");
    }})
}


function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
}


function addEventListeners() {
  let items = document.querySelectorAll('.draggable-item');
  items.forEach(function(item) {
    item.addEventListener('dragstart', handleDragStart, false);
    item.addEventListener('dragenter', handleDragEnter, false);
    item.addEventListener('dragover', handleDragOver, false);
    item.addEventListener('dragleave', handleDragLeave, false);
    item.addEventListener('drop', handleDrop, false);
    item.addEventListener('dragend', handleDragEnd, false);
    item.addEventListener('dragend', handleDragExit, false);
  });

  deleteZone = document.getElementById('delete-zone');

  if (deleteZone) {
    console.log('Element found:', deleteZone);
  } else {
    console.log('Element not found');
  }
  deleteZone.addEventListener('dragover', (event) => {
      event.preventDefault();
      deleteZone.style.backgroundColor = '#FFCCCB';
  });

  deleteZone.addEventListener('dragleave', (event) => {
    deleteZone.style.backgroundColor = ''; // Reset visual feedback
  });

  deleteZone.addEventListener('drop', (event) => {
    event.preventDefault(); // Prevent default behavior
    deleteZone.style.backgroundColor = ''; // Reset visual feedback
    var link = draggedElement.querySelector('a').href;
    draggedElement.parentNode.removeChild(draggedElement);

    console.log(getCookie("feeds"));
    fetch('http://localhost:3000/removelink', {
      method: 'POST',
      headers: {
        'Content-Type': 'text/plain',
      },
      body: link
      }).then(response => {
        if (!response.ok) { // If response is not ok, throw an error
          alert("Link not valid or already exists");
        }})
    //dropZone.appendChild(document.getElementById(data));
  });
  const addItemButton = document.getElementById('add-item');
  addItemButton.addEventListener('click', handleAddItem, false);
}


document.addEventListener('DOMContentLoaded', function() {addEventListeners(); });