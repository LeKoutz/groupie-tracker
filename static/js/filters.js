// filters.js handles the frontend logic for filtering artists on the homepage.
// It captures input from range sliders, checkboxes, and text inputs,
// constructs a query string, and fetches filtered results from the backend.

// --- Event Listeners ---
document.addEventListener("DOMContentLoaded", () => {
    // Elements
    const minCreationDateInput = document.getElementById("min-creation-date");
    const maxCreationDateInput = document.getElementById("max-creation-date");
    const creationDateVal = document.getElementById("creation-date-val");

    const minFirstAlbumInput = document.getElementById("min-first-album");
    const maxFirstAlbumInput = document.getElementById("max-first-album");
    const firstAlbumVal = document.getElementById("first-album-val");

    const memberCheckboxes = document.querySelectorAll(".member-checkbox");
    const locationInput = document.getElementById("location-search");

    const resultsBox = document.querySelector(".search-results ul");
    const noResultsBox = document.querySelector(".no-results");
    const artistGrid = document.getElementById("artist-list");

    // State
    let debounceTimer;

    // --- Helpers ---

    /**
     * Updates the text display for range sliders (e.g., "1990 - 2000").
     */
    const updateDisplayValues = () => {
        creationDateVal.textContent = `${minCreationDateInput.value} - ${maxCreationDateInput.value}`;
        firstAlbumVal.textContent = `${minFirstAlbumInput.value} - ${maxFirstAlbumInput.value}`;
    };
    updateDisplayValues();

    // --- Core Logic ---

    /**
     * Collects all filter values, debounces the request, and fetches data from the API.
     * Endpoints:
     * - /api/filter: Returns full artist objects for the grid view.
     */
    const fetchFilteredResults = () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(async () => {
            const params = new URLSearchParams();

            // Date Ranges
            params.append("min_creation_date", minCreationDateInput.value);
            params.append("max_creation_date", maxCreationDateInput.value);
            params.append("min_first_album_year", minFirstAlbumInput.value);
            params.append("max_first_album_year", maxFirstAlbumInput.value);

            // Members: Calculate min and max based on checked boxes to support backend range logic.
            const checkedMembers = Array.from(memberCheckboxes)
                .filter(cb => cb.checked)
                .map(cb => cb.value);

            if (checkedMembers.length > 0) {
                const values = checkedMembers.map(v => parseInt(v));
                params.append("min_members", Math.min(...values));
                params.append("max_members", Math.max(...values));
            }

            // Location
            if (locationInput.value) {
                params.append("locations", locationInput.value);
            }

            // Search Query: Include main search bar text if present
            const searchInput = document.querySelector("input[name='search']");
            if (searchInput && searchInput.value) {
                params.append("search", searchInput.value);
            }

            try {
                // Use /api/filter to get full artist objects
                const res = await fetch(`/api/filter?${params.toString()}`);
                const data = await res.json();
                renderResults(data);
            } catch (err) {
                console.error("Error fetching filters:", err);
            }
        }, 300); // 300ms debounce
    };

    /**
     * Renders the fetched artist data into the grid.
     * @param {Array} data - List of artist objects from the backend.
     */
    const renderResults = (data) => {
        if (!data || data.length === 0) {
            artistGrid.innerHTML = '<p class="no-results" style="width:100%; text-align:center;">No artists found matching criteria.</p>';
            return;
        }

        artistGrid.innerHTML = data.map(artist => {
            // Logic for members display
            const membersCount = artist.members ? artist.members.length : 0;
            const membersDisplay = membersCount === 1 ? "Solo" : membersCount;

            return `
            <div href="/artist/${artist.id}" class="artist-card">
                <a href="/artist/${artist.id}" class="artist-tag">
                <div class="card-image">
                    <img src="${artist.image}" alt="${artist.name}">
                </div>

                <div class="card-content">
                    <section class="band-info">
                        <h3>${artist.name}</h3>
                    </section>
                    <section class="card-info">
                        <p><strong>Members:</strong> ${membersDisplay}</p>
                        <p><strong>Start Year:</strong> ${artist.creationDate}</p>
                        <p><strong>First Release:</strong> ${artist.firstAlbum}</p>
                        <a href="/artist/${artist.id}" class="green-button">More Info →</a>
                    </section>
                </div>
                </a>
            </div>
            `;
        }).join("");
    };

    // Date Range Inputs
    [minCreationDateInput, maxCreationDateInput].forEach(el => {
        el.addEventListener("input", () => {
            updateDisplayValues();
            fetchFilteredResults();
        });
    });

    [minFirstAlbumInput, maxFirstAlbumInput].forEach(el => {
        el.addEventListener("input", () => {
            updateDisplayValues();
            fetchFilteredResults();
        });
    });

    // Checkboxes & Location Input
    memberCheckboxes.forEach(cb => {
        cb.addEventListener("change", fetchFilteredResults);
    });

    locationInput.addEventListener("input", fetchFilteredResults);

    // Reset Button Logic
    const resetButton = document.getElementById("reset-filters");
    if (resetButton) {
        resetButton.addEventListener("click", () => {
            // Reset Range Inputs to min/max attributes
            minCreationDateInput.value = minCreationDateInput.min;
            maxCreationDateInput.value = maxCreationDateInput.max;
            minFirstAlbumInput.value = minFirstAlbumInput.min;
            maxFirstAlbumInput.value = maxFirstAlbumInput.max;

            // Uncheck all member checkboxes
            memberCheckboxes.forEach(cb => cb.checked = false);

            // Clear location input
            locationInput.value = "";

            // Update display text and fetch original results
            updateDisplayValues();
            fetchFilteredResults();
        });
    }
});
