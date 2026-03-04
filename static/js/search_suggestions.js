document.addEventListener("DOMContentLoaded", () => {
  const form = document.querySelector(".search-form");
  const input = form.querySelector("input[name='search']");
  const resultsBox = document.querySelector(".search-suggestions");
  const categorySelect = form.querySelector("select[name='category']");
  console.log(resultsBox);

  let debounceTimer;

  const fetchResults =  () => {
    clearTimeout(debounceTimer);

    const query = input.value.trim();
    const category = categorySelect.value;
    if (!query) {
      resultsBox.innerHTML = "";
      resultsBox.style.display = "none";
      return;
    }

    debounceTimer = setTimeout(async () => {
      const res = await fetch(`/api/search?search=${encodeURIComponent(query)}&category=${encodeURIComponent(categorySelect.value)}`);
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
  };
  input.addEventListener("input", fetchResults);
  categorySelect.addEventListener("change", fetchResults);
  // Navigate with arrow keys
  input.addEventListener("keydown", (e) => {
    const items = resultsBox.querySelectorAll("li");
    if (items.length === 0) return;

    let index = Array.from(items).findIndex(item => item.classList.contains("highlight"));
    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (index < items.length - 1) {
        if (index >= 0) items[index].classList.remove("highlight");
        items[++index].classList.add("highlight");
        items[index].scrollIntoView({ block: "nearest" });
      }
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      if (index > 0) {
        items[index].classList.remove("highlight");
        items[--index].classList.add("highlight");
        items[index].scrollIntoView({ block: "nearest" });
      }
    } else if (e.key === "Enter" && index >= 0) {
      e.preventDefault();
      const link = items[index].querySelector("a");
      window.location.href = link.href;
    }
  });
  // Clear input with ESC
  input.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
      input.value = "";
      resultsBox.innerHTML = "";
      resultsBox.style.display = "none";
    }
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
