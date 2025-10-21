// This tells the browser to wait until the HTML page structure is fully loaded 
document.addEventListener('DOMContentLoaded', () => { 
    
    // Everything we write will go inside this large function.
    
    // 1. Find the parts of the HTML we need to control, using the unique "id"
    const searchButton = document.getElementById('searchButton');
    const searchInput = document.getElementById('searchInput');
    const artistListContainer = document.getElementById('artist-list');
    
    // 2. Add an "ear" (event listener) to the button
    if (searchButton) {
        searchButton.addEventListener('click', handleSearch);
    }
    
    // ... inside document.addEventListener ...

    // 3. Define the main function that handles the search
    function handleSearch() {
        // Get the text the user typed in the box and remove extra spaces
        const query = searchInput.value.trim(); 

        if (query === "") {
            console.log("Search query is empty.");
            return; // Stop the function if the box is empty
        }

        console.log(`User wants to search for: "${query}"`);
        
        // *************************************************************
        // *** THE CORE CLIENT-SERVER COMMUNICATION ***
        // *************************************************************
        
        // The 'fetch' API asks the server for information
        // We build the address: /api/search?q=TEXT_THE_USER_TYPED
        fetch(`/api/search?q=${query}`) 
            .then(response => {
                // First: Check if the server response was successful (status 200, "ok")
                if (!response.ok) {
                    // If the server failed (e.g., status 404 or 500), we stop and trigger the .catch
                    throw new Error(`Server status problem: ${response.status}`);
                }
                // If it was successful, we tell the browser to read the data as JSON
                return response.json();
            })
            .then(data => {
                // SUCCESS! 'data' now holds the filtered list of artists sent by Person 2
                console.log("Search results received:", data);
                
                // Now, send this new data to the function that draws the cards
                updateArtistGrid(data); 
            })
            .catch(error => {
                // This runs if the fetch failed (network issue, server offline, etc.)
                console.error("Error during search:", error);
                // Display a friendly message to the user on the page
                artistListContainer.innerHTML = `<p class="error-message">Sorry, search failed! Server is unreachable or an error occurred.</p>`;
            });
    }

    // ... function updateArtistGrid will go here ...
    
    // 4. Function to draw the artist cards on the page
    function updateArtistGrid(artists) {
        // 1. Clear the screen: erase all the old dummy cards
        artistListContainer.innerHTML = ''; 

        if (artists.length === 0) {
             artistListContainer.innerHTML = `<p>No artists found matching your search. Try a different keyword.</p>`;
             return;
        }

        // 2. Go through each artist the server sent back, one by one
        artists.forEach(artist => {
            // 3. Build the HTML code for ONE card (using backticks ` ` for multi-line text)
            // NOTE: The structure here must exactly match the HTML you designed!
            const cardHTML = `
                <div class="artist-card">
                    <img src="${artist.Image}" alt="${artist.Name} Image">
                    
                    <h3>${artist.Name}</h3>
                    <hr>
                    <p><strong>Formed:</strong> ${artist.CreationDate}</p>
                    <p><strong>First Album:</strong> ${artist.FirstAlbum}</p>
                    
                    <a href="/artist/${artist.ID}" class="details-button">View Tour Details</a>
                </div>
            `;
            // 4. Add the new card's HTML code to the main container
            artistListContainer.innerHTML += cardHTML;
        });
    }

}); // Closes the DOMContentLoaded listener;