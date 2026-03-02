document.addEventListener("DOMContentLoaded", () => {
    const inputs = {
        minCD: document.getElementById("min-creation-date"),
        maxCD: document.getElementById("max-creation-date"),
        minFA: document.getElementById("min-first-album"),
        maxFA: document.getElementById("max-first-album"),
        members: document.querySelectorAll(".member-checkbox"),
        location: document.getElementById("location-search"),
        search: document.querySelector("input[name='search']"),
        reset: document.getElementById("reset-filters")
    };
    const grid = document.getElementById("artist-list");
    let debounce;
    // Build query params from all inputs
    const buildParams = () => {
        const params = new URLSearchParams();
        params.set("min_creation_date", inputs.minCD.value);
        params.set("max_creation_date", inputs.maxCD.value);
        params.set("min_first_album_year", inputs.minFA.value);
        params.set("max_first_album_year", inputs.maxFA.value);
        const checked = [...inputs.members].filter(c => c.checked).map(c => +c.value);
        if (checked.length) params.set("min_members", Math.min(...checked));
        if (checked.length) params.set("max_members", Math.max(...checked));
        if (inputs.location.value) params.set("locations", inputs.location.value);
        if (inputs.search?.value) params.set("search", inputs.search.value);
        return params;
    };
    // Fetch and render
    const fetchResults = async () => {
        clearTimeout(debounce);
        debounce = setTimeout(async () => {
            const res = await fetch(`/api/filter?${buildParams()}`);
            const artists = await res.json();
            grid.innerHTML = artists.length ? artists.map(renderCard).html : 
                '<p class="no-results">No artists found</p>';
        }, 300);
    };
    // Render single card
    const renderCard = a => `
        <div class="artist-card">
            <a href="/artist/${a.id}">
                <img src="${a.image}" alt="${a.name}">
                <h3>${a.name}</h3>
                <p>Members: ${a.members?.length || 0}</p>
                <p>Start: ${a.creationDate}</p>
            </a>
        </div>`;
    // Single event listener for all filter inputs
    document.querySelectorAll("input, select").forEach(el => 
        el.addEventListener("input", fetchResults));
    // Reset handler
    inputs.reset?.addEventListener("click", () => {
        document.querySelectorAll("input[type='range'], input[type='checkbox']").forEach(el => {
            el.value = el.min || "";
            el.checked = false;
        });
        inputs.location.value = "";
        fetchResults();
    });
});
