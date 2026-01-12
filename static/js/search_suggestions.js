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

      if (!results || results.length === 0) {
        resultsBox.style.display = "none";
        return;
      }

      resultsBox.innerHTML = `
        <ul>
          ${results.map(r => `
            <li>
              <a href="/artist/${r.ID}">${r.Label}</a>
            </li>
          `).join("")}
        </ul>
      `;

      resultsBox.style.display = "block";
    }, 300);
  });
});
