document.addEventListener("DOMContentLoaded", () => {
  const form = document.querySelector(".search-form");
  const input = form.querySelector("input[name='search']");
  const resultsBox = document.querySelector(".search-suggestions");
  console.log(resultsBox);

  let debounceTimer;

  input.addEventListener("input", () => {
    clearTimeout(debounceTimer);

    const query = input.value.trim();
    if (!query) {
      resultsBox.innerHTML = "";
      resultsBox.style.display = "none";
      return;
    }

    debounceTimer = setTimeout(async () => {
      const res = await fetch(`/api/search?search=${encodeURIComponent(query)}`);
      const results = await res.json();

      // If no results found show a message, otherwise show a dropdown list
      if (!results || results.length === 0) {
        resultsBox.innerHTML = `<p class="no-results">No results found</p>`;
      } else {
        resultsBox.innerHTML = `
          <ul>
            ${results.map(r => `
              <li>
                <a href="/artist/${r.ID}">${r.Label}</a>
              </li>
            `).join("")}
          </ul>
        `;
            }

      resultsBox.style.display = "block";
    }, 300);
  });
  // close dropdown when clicking outside
  document.addEventListener("click", (e) => {
    if (!form.contains(e.target) && !resultsBox.contains(e.target)) {
      resultsBox.style.display = "none";
    } else {
      if (resultsBox.innerHTML !== "") {
        resultsBox.style.display = "block";
      }
    }
  });
});
