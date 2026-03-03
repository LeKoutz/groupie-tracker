# Groupie Tracker

A web application that displays information about your favorite artists and bands by fetching data from multiple API endpoints.

## Features

- **Asynchronous API Integration**: Non-blocking data fetching from multiple endpoints with graceful handling of incomplete data
- **Advanced Filtering System**:
    - Real-time dual range sliders for Creation Date and First Album Year with synchronized numeric inputs
    - Member count filtering with checkbox selection
    - Location-based filtering with autocomplete
    - Instant grid updates as filters are adjusted
    - Reset functionality to quickly clear all filters
    - Mobile-optimized filter sidebar with toggle button
- **Geolocalization & Mapping**: 
    - Integration with OpenStreetMap (Nominatim API) to convert tour locations into geographic coordinates
    - Intelligent caching system with persistence (`locations.json`) to minimize API hits
    - Asynchronous background geocoding to pre-populate location data
- **Progressive Loading**: Server starts immediately; redirects to loading page while data is being fetched
- **Progressive Enhancement**: Search-bar functionality works with vanilla form submission, enhanced with JavaScript for dynamic suggestions
- **Zero external dependencies**: Pure Go backend with only standard packages

## Visual Enhancements
   * Interactive Maps: Leaflet.js powered maps showing concert locations with chronological tour paths.
   * Dynamic Path Highlighting: Visualizes previous and next tour stops when clicking on specific location markers.
   * Smooth Scrolling: Fluid page navigation for a more polished feel.
   * Interactive UI: Bouncy "spring" hover effects on images and card highlights for better feedback.
   * Mobile-Ready: Optimized responsive layouts for error pages and artist grids.
   * Accessible Keyboard Navigation: ESC to close sidebar/suggestions, arrow keys to navigate search results

## Tech Stack

- **Backend**: Go
- **Frontend**:
    - HTML/CSS
    - Javascript
- **Deployment**: Railway

## Installation

1. Clone the repository:
```bash
git clone https://platform.zone01.gr/git/gkoutzos/groupie-tracker-filters.git
cd groupie-tracker-filters
```

2. Run the application:
```bash
go run main.go
```

3. Open your browser and navigate to `http://localhost:8080` (or whatever port is set in your PORT environment variable)

## Deployed

Check out the [live](https://groupie-tracker-production-e572.up.railway.app/) application

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributors

- [gkoutzos](https://platform.zone01.gr/git/gkoutzos)
- [cktistak](https://platform.zone01.gr/git/cktistak)
