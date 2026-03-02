document.addEventListener("DOMContentLoaded", () => {
  const inputs = {
    minCD: document.getElementById("min-creation-date"),
    maxCD: document.getElementById("max-creation-date"),
    minFA: document.getElementById("min-first-album"),
    maxFA: document.getElementById("max-first-album"),
    members: document.querySelectorAll(".member-checkbox"),
    location: document.getElementById("location-search"),
    search: document.querySelector("input[name='search']"),
    reset: document.getElementById("reset-filters"),
  };
  const updateDisplayValues = () => {
    document.getElementById("creation-date-val").textContent =
      inputs.minCD.value + " - " + inputs.maxCD.value;
    document.getElementById("first-album-val").textContent =
      inputs.minFA.value + " - " + inputs.maxFA.value;
  };
  updateDisplayValues();
  const grid = document.getElementById("artist-list");

  let debounce;
  const sidebar = document.querySelector(".sidebar");
  document.getElementById("toggle-filters").addEventListener("click", () => {
    sidebar.classList.toggle("visible");
  });
  const buildParams = () => {
    const params = new URLSearchParams();
    params.set("min_creation_date", inputs.minCD.value);
    params.set("max_creation_date", inputs.maxCD.value);
    params.set("min_first_album_year", inputs.minFA.value);
    params.set("max_first_album_year", inputs.maxFA.value);
    const checked = [...inputs.members]
      .filter((c) => c.checked)
      .map((c) => +c.value);
    if (checked.length) params.set("min_members", Math.min(...checked));
    if (checked.length) params.set("max_members", Math.max(...checked));
    if (inputs.location.value) params.set("locations", inputs.location.value);
    if (inputs.search?.value) params.set("search", inputs.search.value);
    return params;
  };
  const fetchResults = async () => {
    clearTimeout(debounce);
    debounce = setTimeout(async () => {
      try {
        const res = await fetch(`/api/filter?${buildParams()}`);
        if (!res.ok) throw new Error("Network error");
        const artists = await res.json();
        grid.innerHTML =
          artists && artists.length
            ? artists.map(renderCard).join("")
            : '<p class="no-results">No artists found</p>';
      } catch (err) {
        console.error("Error:", err);
      }
    }, 300);
  };
  const renderCard = (a) => `
        <div class="artist-card">
            <a href="/artist/${a.id}">
                <img src="${a.image}" alt="${a.name}">
                <h3>${a.name}</h3>
                <p>Members: ${a.members?.length || 0}</p>
                <p>Start: ${a.creationDate}</p>
            </a>
        </div>`;
  [
    inputs.minCD,
    inputs.maxCD,
    inputs.minFA,
    inputs.maxFA,
    inputs.location,
  ].forEach((el) => {
    el?.addEventListener("input", () => {
      updateDisplayValues();
      fetchResults();
    });
  });
  inputs.members.forEach((cb) => cb.addEventListener("change", fetchResults));
  inputs.reset?.addEventListener("click", () => {
    inputs.minCD.value = inputs.minCD.min;
    inputs.maxCD.value = inputs.maxCD.max;
    inputs.minFA.value = inputs.minFA.min;
    inputs.maxFA.value = inputs.maxFA.max;
    updateDisplayValues();
    inputs.members.forEach((cb) => (cb.checked = false));
    inputs.location.value = "";
    fetchResults();
  });
});
