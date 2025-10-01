console.log("initialize setup")


// Theme swicher setup
const themes = ['theme-light', 'theme-dark' ];

document.addEventListener('DOMContentLoaded', () => {
	const select = document.getElementById('theme-dropdown');
	if (!select) return;

	const savedTheme = localStorage.getItem('theme') || 'theme-light';

	document.body.classList.remove(...Array.from(document.body.classList).filter(c => c.startsWith("theme-")));
	document.body.classList.add(savedTheme);
	select.value = savedTheme;

	select.addEventListener("change", function () {
		const selectedValue = select.value;

		document.body.classList.remove(...Array.from(document.body.classList).filter(c => c.startsWith("theme-")));
		document.body.classList.add(selectedValue);

		localStorage.setItem('theme', selectedValue);
	});
});



// Server side event setup

const evtSource = new EventSource("http://localhost:8080/events");
const updatesDiv = document.getElementById("columnHolder");

evtSource.onmessage = function(event) {
	updatesDiv.insertAdjacentHTML("beforeend", event.data);
};

evtSource.onerror = function() {
	updatesDiv.insertAdjacentHTML("beforeend", "<p style='color:red;'>Connection lost...</p>");

};




// File upload list genrator

function showFiles() {
 const input = document.getElementById("pdfs");
 const list = document.getElementById("fileList");
 list.innerHTML = "";
 for (let file of input.files) {
				 let li = document.createElement("li");
				 li.textContent = file.name;
				 list.appendChild(li);
 }
}


function cleanColumn(columnId) {
	const col = document.getElementById(columnId);
	if (!col) return;

	const cols = Array.from(holder.querySelectorAll("[id^='col-']"));
	cols.forEach((col, index) => {
		const isLast = index === cols.length - 1;
		const isEmpty =
			col.children.length === 0 && col.innerText.trim() === "";

		if (isEmpty && !isLast) {
			col.remove();
		}
	});

const updatedChildren = Array.from(col.children);
const last = updatedChildren[updatedChildren.length - 1];

if (!last || last.children.length > 0 || last.innerText.trim() !== "") {
		const emptyDiv = document.createElement("div");
		emptyDiv.className = "w-[25%] bg-gray-400  flex flex-col space-y-2 hover:border-2 border-hover rounded-tl-2xl rounded-tr-2xl";
		emptyDiv.id="col-empty"
		col.appendChild(emptyDiv);
}
}

function mergeSections() {
	const mergeDiv = event.currentTarget.parentElement;
	const prevSlot = mergeDiv.previousElementSibling;
	const nextSlot = mergeDiv.nextElementSibling;
	if (!prevSlot?.id.startsWith("slot-") || !nextSlot?.id.startsWith("slot-")) return;
		prevSlot.innerHTML += nextSlot.innerHTML;
		nextSlot.remove();
		mergeDiv.remove();
}

function updateMergeDivs() {
	observer.disconnect();

	document.querySelectorAll("[id^='col-']").forEach(col => {
		cleanColumn(col);

		col.querySelectorAll(".filtered").forEach(el => el.remove());

		const slots = Array.from(col.querySelectorAll("[id^='slot-']"));
		if (slots.length < 2) return;

		for (let i = 0; i < slots.length - 1; i++) {
			const mergeDiv = document.createElement("div");
			mergeDiv.className = "filtered flex items-center justify-between px-2 py-1";
			mergeDiv.id = "optionMenu";
			mergeDiv.style.transition = "opacity 0.3s";

		mergeDiv.innerHTML = `
			<button 
				class="px-3 py-1 rounded-lg text-gray-50 button-primary button-primary-hover"
				onclick="mergeSections()">
				Merge
			</button>
			<button 
				class="px-3 py-1 rounded-lg bg-gray-300 text-primary hover:bg-gray-400"
				onclick="this.parentElement.style.display='none'">
				Dismiss
			</button>
		`;
		// slots[i].insertAdjacentElement("afterend", mergeDiv);
		slots[i].parentElement.insertBefore(mergeDiv, slots[i].nextSibling);

		}
	});
observeColumns();
}

function observeColumns() {
document.querySelectorAll("[id^='col-']").forEach(col => {
		observer.observe(col, { childList: true });
});
}

const observer = new MutationObserver(updateMergeDivs);
updateMergeDivs();




function initSortables() {
	document.querySelectorAll('[id^="col-"]').forEach(element => {
		new Sortable(element, {
			filter: '.filtered',
			group: 'column',
			animation: 150,
			swapThreshold: 0.65,
			ghostClass: 'opacity-50',
		});
	});

	document.querySelectorAll('[id^="slot-"]').forEach(element => {
		new Sortable(element, {
			group: 'slot',
			filter: 'slothandle',
			animation: 150,
			swapThreshold: 0.65,
			ghostClass: 'opacity-40',
			multiDrag: true,
			selectedClass: "selected"

			// scroll: true,
			// forceAutoScrollFallback: false,
			// scrollFn: function(offsetX, offsetY, originalEvent, touchEvt, hoverTargetEl) { ... },
			// scrollSensitivity: 30,
			// scrollSpeed: 10,
			// bubbleScroll: true

		});
	});
}

initSortables();

