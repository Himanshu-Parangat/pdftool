console.log("hi")


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





