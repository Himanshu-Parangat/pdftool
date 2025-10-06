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



function createFileColumnsFromJSON(data) {
    const holder = document.getElementById('columnHolder');
    if (!holder) return;

    Object.entries(data).forEach(([filename, fileData]) => {
        const fileId = filename.slice(0, 15);
        const fileDisplayName = filename.slice(16);

        const columnDiv = document.createElement('div');
        columnDiv.id = `column-${fileId}`;
        columnDiv.className = 'min-w-[20%] flex flex-col  border-1 border-dotted hover:border-2 border-hover rounded-tl-2xl rounded-tr-2xl space-y-3 text-center';
        columnDiv.dataset.id = fileId;
        columnDiv.dataset.fileCount = fileData.page_count || 0;

        const fileDiv = document.createElement('div');
        fileDiv.id = `files-${fileId}`;
        fileDiv.className = 'bg-primary p-1 rounded-t-2xl';
        Object.assign(fileDiv.dataset, {
            id: fileId,
            filename: fileData.filename || '',
            author: fileData.author || '',
            creationDate: fileData.creation_date || '',
            creator: fileData.creator || '',
            modificationDate: fileData.modification_date || '',
            pageCount: fileData.page_count || '',
            pageSize: fileData.page_size || '',
            producer: fileData.producer || '',
            subject: fileData.subject || '',
            title: fileData.title || '',
            version: fileData.version || '',
        });

        const infoDiv = document.createElement('div');
        infoDiv.textContent = fileDisplayName;
				infoDiv.className = "truncate text-sm pl-2 pr-2 border-b-1 pb-2"
        fileDiv.appendChild(infoDiv);

        const pageHolder = document.createElement('div');
        pageHolder.id = `pageHolder-${fileId}`;
        pageHolder.className = 'm-2 space-y-3 overflow-y-auto';

        (fileData.pages || []).forEach(page => {
            const pageDiv = document.createElement('div');
            pageDiv.id = `page-${page.id}`;
            pageDiv.className = 'border-1 border-primary rounded-xl flex flex-row justify-center overflow-hidden';
            Object.assign(pageDiv.dataset, {
                pageNumber: page.pagenumber || '',
                flip: page.flip || '',
                id: page.id || '',
                pageOrientation: page.pageorientation || '',
                previewPath: page.preview_path || '',
                rotate: page.rotate || '',
                status: page.status || '',
            });

            const img = document.createElement('img');
            img.src = page.preview_path || '';
            img.alt = `preview page ${page.pagenumber || ''}`;
            img.className = 'max-h-75 object-contain';

            pageDiv.appendChild(img);
            pageHolder.appendChild(pageDiv);
        });

        fileDiv.appendChild(pageHolder);
        columnDiv.appendChild(fileDiv);

        holder.prepend(columnDiv);
    });

    initSortables();
}


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
	document.querySelectorAll('[id^="pageHolder-"]').forEach(element => {
		new Sortable(element, {
			group: 'page',
			delay: 0,
			animation: 0,
			swapThreshold: 0.65,
			ghostClass: 'opacity-40',
			selectedClass: "selectedPage",
		  forceFallback: true,
			onEnd: () => columncallback()
		});
	});
	document.querySelectorAll('[id^="column-"]').forEach(element => {
		new Sortable(element, {
			group: 'section',
			delay: 0,
			animation: 0,
			swapThreshold: 0.65,
			ghostClass: 'opacity-40',
			selectedClass: "selectedFile",
		  forceFallback: true,
			onEnd: () => columncallback()
		});
	});
}

initSortables();
