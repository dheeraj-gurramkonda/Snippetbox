// Mark the navigation link for the current path so visitors can identify the active page.
var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		// Navigation paths are unique here, so no later link needs checking.
		break;
	}
}