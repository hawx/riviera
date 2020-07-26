const toggles = document.querySelectorAll('[data-toggle]');

for (let toggle of toggles) {
    const attr = toggle.getAttribute('data-toggle');
    
    toggle.addEventListener('click', () => {
        const el = document.querySelector("[data-toggled='" + attr + "']");

        el.classList.toggle('open');
    });
}
