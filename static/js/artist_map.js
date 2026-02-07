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
  let markerCount = 0; // for valid markers
  
    locations.forEach((location, index) => {
        const lat = parseFloat(location.lat);
        const lon = parseFloat(location.lon);
        // Filter out invalid coordinates
      if (!isNaN(lat) && !isNaN(lon)) {
            markerCount++;
            const marker = L.marker([lat, lon]).addTo(map);
            // replace hyphens with spaces and convert to uppercase for the display name
            const displayName = location.name.replace(/[-]/g, ` `).toUpperCase();
            marker.bindPopup(`<b>${markerCount}.${displayName}</b>`); // add the display name to the marker

            // Highlight path to next location when clicked
            marker.on('popupopen', () => {
                // Remove any old temporary lines
                if (window.currentLine) map.removeLayer(window.currentLine);

                // Is there a next location?
                if (index < locations.length - 1) {
                    const nextLoc = locations[index + 1];
                    const nextLat = parseFloat(nextLoc.lat);
                    const nextLon = parseFloat(nextLoc.lon);

                    if (!isNaN(nextLat) && !isNaN(nextLon)) {
                        // Draw a line to the NEXT one
                        window.currentLine = L.polyline([[lat, lon], [nextLat, nextLon]], {
                            color: '#ee0c0cff', // Red for "Next Step"
                            weight: 3,
                            opacity: 0.7,
                            dashArray: '10,10'
                        }).addTo(map);
                    }
                }
            });

            bounds.push([lat, lon]); // add the marker to the bounds
        }
    });

    if (bounds.length > 0) {
        // Connect markers with lines (Chronological order from backend)
        L.polyline(bounds, {
            color: '#97CE4C',
            weight: 3,
            opacity: 0.7,
            dashArray: '10, 10', // Dashed line for a "tour path" look
            lineJoin: 'round'
        }).addTo(map);
        // fit the map to the bounds
        map.fitBounds(bounds, { padding: [50, 50] });
    }
});