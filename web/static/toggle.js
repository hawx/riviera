const toggles = document.querySelectorAll('[data-toggle]');

for (let toggle of toggles) {
    const attr = toggle.getAttribute('data-toggle');
    
    toggle.addEventListener('click', () => {
        const els = document.querySelectorAll("[data-toggled='" + attr + "']");

        els.forEach(el => el.classList.toggle('open'));
    });
}
