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
    draggedElement.innerHTML = this.innerHTML;
    this.innerHTML = toExchange;
    //e.target.classList.remove("over");
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
  //creates a popup that asks for the name of the new item and validates it
  let newLink = prompt("Enter the link to the new feed");
  if (newLink === null || newLink === "") {
    return;
  }
  const data = { newLink }; 
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

  const addItemButton = document.getElementById('add-item');
  addItemButton.addEventListener('click', handleAddItem, false);
}


document.addEventListener('DOMContentLoaded', function() {addEventListeners(); });