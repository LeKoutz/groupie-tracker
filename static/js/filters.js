document.addEventListener("DOMContentLoaded", () => {
  // Get all elements
  const toggleFiltersButton = document.getElementById("toggle-filters");
  const sidebar = document.querySelector(".sidebar");
  const minCDSlider = document.getElementById("min-creation-date");
  const maxCDSlider = document.getElementById("max-creation-date");
  const minFASlider = document.getElementById("min-album-year");
  const maxFASlider = document.getElementById("max-album-year");
  const minCDInput = document.getElementById("min-creation-input");
  const maxCDInput = document.getElementById("max-creation-input");
  const minFAInput = document.getElementById("min-album-input");
  const maxFAInput = document.getElementById("max-album-input");
  const members = document.querySelectorAll(".member-checkbox");
  const location = document.getElementById("location-search");
  const search = document.querySelector("input[name='search']");
  const reset = document.getElementById("reset-filters");
  const grid = document.getElementById("artist-list");
  if (!grid) return;
  
  let debounce;
  // Toggle sidebar
  if (toggleFiltersButton && sidebar) {
    toggleFiltersButton.addEventListener("click", (e) => {
      e.stopPropagation();
      sidebar.classList.toggle("visible");
    });
    
    document.addEventListener("click", (e) => {
      if (!sidebar.contains(e.target) && e.target !== toggleFiltersButton) {
        sidebar.classList.remove("visible");
      }
    });
  // Close sidebar on Escape key press
  document.addEventListener("keydown", (e) => {
        if (e.key === "Escape" && sidebar.classList.contains("visible")) {
            sidebar.classList.remove("visible");
        }
    });
  }
  // Helper functions
  const updateDisplayValues = () => {
    const cdVal = document.getElementById("creation-date-val");
    const faVal = document.getElementById("first-album-val");
    if (cdVal) cdVal.textContent = `${minCDSlider?.value} - ${maxCDSlider?.value}`;
    if (faVal) faVal.textContent = `${minFASlider?.value} - ${maxFASlider?.value}`;
  };
  const syncSliderToInput = (slider, input) => {
    if (input && slider) input.value = slider.value;
    updateDisplayValues();
  };
  const syncInputToSlider = (input, slider, otherSlider, isMin) => {
    if (!input || !slider || !otherSlider) return;
    
    let val = parseInt(input.value);
    const min = parseInt(slider.min);
    const max = parseInt(slider.max);
    
    val = Math.max(min, Math.min(max, val || min));
    input.value = val;
    
    const otherVal = parseInt(otherSlider.value);
    if (isMin && val > otherVal) {
      otherSlider.value = val;
    } else if (!isMin && val < otherVal) {
      otherSlider.value = val;
    }
    
    slider.value = val;
    updateDisplayValues();
  };
  const buildParams = () => {
    const params = new URLSearchParams();
    params.set("min_creation_date", minCDSlider?.value || 1950);
    params.set("max_creation_date", maxCDSlider?.value || 2026);
    params.set("min_first_album_year", minFASlider?.value || 1950);
    params.set("max_first_album_year", maxFASlider?.value || 2026);
    
    const checked = [...members].filter(c => c.checked).map(c => +c.value);
    if (checked.length) {
      params.set("min_members", Math.min(...checked));
      params.set("max_members", Math.max(...checked));
    }
    if (location?.value) params.set("locations", location.value);
    if (search?.value) params.set("search", search.value);
    return params;
  };
  const renderCard = (a) => {
    const memberCount = a.members?.length === 1 ? "Solo" : a.members?.length || 0;
    return `
      <div class="artist-card">
        <a href="/artist/${a.id}" class="artist-tag">
          <div class="card-image">
            <img src="${a.image}" alt="${a.name}">
          </div>
          <div class="card-content">
            <section class="band-info">
              <h3>${a.name}</h3>
            </section>
            <section class="card-info">
              <p><strong>Members:</strong> ${memberCount}</p>
              <p><strong>Start Year:</strong> ${a.creationDate}</p>
              <p><strong>First Release:</strong> ${a.firstAlbum}</p>
              <a href="/artist/${a.id}" class="green-button">More Info →</a>
            </section>
          </div>
        </a>
      </div>`;
  };
  const fetchResults = async () => {
    clearTimeout(debounce);
    debounce = setTimeout(async () => {
      try {
        const res = await fetch(`/api/filter?${buildParams()}`);
        if (!res.ok) throw new Error("Network error");
        const artists = await res.json();
        grid.innerHTML = artists?.length 
          ? artists.map(renderCard).join("") 
          : '<p class="no-results">No artists found</p>';
      } catch (err) {
        console.error("Filter error:", err);
      }
    }, 300);
  };
  // Setup slider events
  const setupSlider = (slider, otherSlider, input, isMin) => {
    if (!slider || !otherSlider) return;
    
    slider.addEventListener("input", () => {
      const sVal = parseInt(slider.value);
      const oVal = parseInt(otherSlider.value);
      
      if (isMin && sVal > oVal) slider.value = oVal;
      if (!isMin && sVal < oVal) slider.value = oVal;
      
      syncSliderToInput(slider, input);
      fetchResults();
    });
  };
  setupSlider(minCDSlider, maxCDSlider, minCDInput, true);
  setupSlider(maxCDSlider, minCDSlider, maxCDInput, false);
  setupSlider(minFASlider, maxFASlider, minFAInput, true);
  setupSlider(maxFASlider, minFASlider, maxFAInput, false);
  // Number input events
  if (minCDInput) minCDInput.addEventListener("change", () => { syncInputToSlider(minCDInput, minCDSlider, maxCDSlider, true); fetchResults(); });
  if (maxCDInput) maxCDInput.addEventListener("change", () => { syncInputToSlider(maxCDInput, maxCDSlider, minCDSlider, false); fetchResults(); });
  if (minFAInput) minFAInput.addEventListener("change", () => { syncInputToSlider(minFAInput, minFASlider, maxFASlider, true); fetchResults(); });
  if (maxFAInput) maxFAInput.addEventListener("change", () => { syncInputToSlider(maxFAInput, maxFASlider, minFASlider, false); fetchResults(); });
  // Other filters
  if (location) location.addEventListener("input", fetchResults);
  members.forEach(cb => cb.addEventListener("change", fetchResults));
  // Reset
  if (reset) {
    reset.addEventListener("click", () => {
      if (minCDSlider) minCDSlider.value = minCDSlider.min;
      if (maxCDSlider) maxCDSlider.value = maxCDSlider.max;
      if (minFASlider) minFASlider.value = minFASlider.min;
      if (maxFASlider) maxFASlider.value = maxFASlider.max;
      
      if (minCDInput) minCDInput.value = minCDSlider?.min;
      if (maxCDInput) maxCDInput.value = maxCDSlider?.max;
      if (minFAInput) minFAInput.value = minFASlider?.min;
      if (maxFAInput) maxFAInput.value = maxFASlider?.max;
      
      updateDisplayValues();
      members.forEach(cb => cb.checked = false);
      if (location) location.value = "";
      fetchResults();
    });
  }
  // Initial display update
  updateDisplayValues();
});