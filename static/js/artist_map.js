document.addEventListener("DOMContentLoaded", () => {
    // check if map element and locations are present
    const mapElement = document.getElementById("map");
    const locations = window.artistMapData;
    
    if (!mapElement || !locations)
        return;

    // initialize the map
    const map = L.map(mapElement).setView([20, 0], 2); // center the map at (20, 0) with zoom level 2

    // add the tile layer
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: "&copy; OpenStreetMap &copy; Groupie Tracker",
        subdomains: 'abcd', // provides multiple tile servers for better performance
        maxZoom: 19, // maximum zoom level
    }).addTo(map);

    // create an array to store the bounds of the markers
    const bounds = [];
    locations.forEach(location => {
        const lat = parseFloat(location.lat);
        const lon = parseFloat(location.lon);
        // Filter out invalid coordinates
        if (!isNaN(lat) && !isNaN(lon)) {
            const marker = L.marker([lat, lon])
                .addTo(map);
        }
    });
}