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




// File upload list genrator

document.addEventListener("DOMContentLoaded", () => {
  const uploadStatus = document.getElementById("uploadStatus");
  const placeholder = document.getElementById("finleTrackPlacehoder");

  function togglePlaceholder() {
    const items = uploadStatus.querySelectorAll("li:not(#finleTrackPlacehoder)");
    placeholder.style.display = items.length === 0 ? "flex" : "none";
  }

  togglePlaceholder();
  const observer = new MutationObserver(togglePlaceholder);
  observer.observe(uploadStatus, { childList: true });
});


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





function getAllColumns(id = "columnHolder") {
  const container = document.getElementById(id);
  if (!container) return [];
  return Array.from(container.querySelectorAll('[id^="column-"]'));
}

function getAllFiles(column) {
  return Array.from(column.querySelectorAll('[id^="files-"]'));
}

function getAllPages(file) {
  return Array.from(file.querySelectorAll('[id^="page-"]'));
}

function getCurrentPlacement() {
  const columns = getAllColumns();
  const structure = {};

  columns.forEach(column => {
    const columnId = column.dataset.id || column.id;
    structure[columnId] = {};

    const files = getAllFiles(column);
    files.forEach(file => {
      const filename = file.dataset.filename || file.id;
      const fileMeta = {
        author: file.dataset.author || "",
        creation_date: file.dataset.creationDate || "",
        creator: file.dataset.creator || "",
        filename,
        modification_date: file.dataset.modificationDate || "",
        page_count: Number(file.dataset.pageCount) || 0,
        page_size: file.dataset.pageSize || "",
        pages: [],
        producer: file.dataset.producer || "",
        subject: file.dataset.subject || "",
        title: file.dataset.title || "",
        version: file.dataset.version || ""
      };

      const pages = getAllPages(file);
      fileMeta.page_count = pages.length;

      pages.forEach((page, i) => {
        const pageData = {
          flip: Number(page.dataset.flip) || 0,
          id: page.dataset.id || `page-${i}`,
          pagenumber: Number(page.dataset.pageNumber) || i + 1,
          pageorientation: page.dataset.pageOrientation || "portrait",
          preview_path: page.dataset.previewPath || "",
          rotate: Number(page.dataset.rotate) || 0,
          status: page.dataset.status || "show"
        };
        fileMeta.pages.push(pageData);
      });

      structure[columnId][filename] = fileMeta;
    });
  });

  // console.log(structure);
	showJson(structure)
  // return structure;
}


function showJson(jsonData) {
  const viewer = document.getElementById("jsonViewer");
  if (!viewer) return;

  viewer.textContent = JSON.stringify(jsonData, null, 2);
}





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

